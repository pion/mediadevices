#include <iostream>
#include <unistd.h>

#include <dshow.h>
#include <qedit.h>
#include <mmsystem.h>

#include "camera_windows.hpp"
#include "_cgo_export.h"


imageProp* getProp(camera* cam, int i)
{
  return &cam->props[i];
}

char* getName(cameraList* list, int i)
{
  return list->name[i];
}


// printErr shows string representation of HRESULT.
// This is for debugging.
void printErr(HRESULT hr)
{
  char buf[128];
  AMGetErrorTextA(hr, buf, 128);
  fprintf(stderr, "%s\n", buf);
}

// getCameraName returns name of the device.
// returned pointer must be released by free() after use.
char* getCameraName(IMoniker* moniker)
{
  IPropertyBag* pPropBag = nullptr;
  VARIANT varName;
  VariantInit(&varName);
  
  // Try to get FriendlyName
  if (SUCCEEDED(moniker->BindToStorage(nullptr, nullptr, IID_IPropertyBag, (void**)&pPropBag)))
  {
    if (SUCCEEDED(pPropBag->Read(L"FriendlyName", &varName, 0)))
    {
      std::string nameStr = utf16Decode(varName.bstrVal);
      char* ret = (char*)malloc(nameStr.size() + 1);
      memcpy(ret, nameStr.c_str(), nameStr.size() + 1);
      VariantClear(&varName);
      pPropBag->Release();
      return ret;
    }
    pPropBag->Release();
  }
  VariantClear(&varName);
  
  // Fallback to display name
  LPOLESTR name;
  if (FAILED(moniker->GetDisplayName(nullptr, nullptr, &name)))
    return nullptr;

  std::string nameStr = utf16Decode(name);
  char* ret = (char*)malloc(nameStr.size() + 1);
  memcpy(ret, nameStr.c_str(), nameStr.size() + 1);

  LPMALLOC comalloc;
  CoGetMalloc(1, &comalloc);
  comalloc->Free(name);

  return ret;
}

// listCamera stores information of the devices to cameraList*.
int listCamera(cameraList* list, const char** errstr)
{
  ICreateDevEnum* sysDevEnum = nullptr;
  IEnumMoniker* enumMon = nullptr;

  if (FAILED(CoCreateInstance(
          CLSID_SystemDeviceEnum, nullptr, CLSCTX_INPROC,
          IID_ICreateDevEnum, (void**)&sysDevEnum)))
  {
    *errstr = errEnumDevice;
    goto fail;
  }

  if (FAILED(sysDevEnum->CreateClassEnumerator(
          CLSID_VideoInputDeviceCategory, &enumMon, 0)))
  {
    *errstr = errEnumDevice;
    goto fail;
  }

  safeRelease(&sysDevEnum);

  if (enumMon == nullptr)
  {
    list->num = 0;
    list->name = nullptr;
    return 0;
  }

  {
    IMoniker* moniker;
    list->num = 0;
    while (enumMon->Next(1, &moniker, nullptr) == S_OK)
    {
      moniker->Release();
      list->num++;
    }

    enumMon->Reset();
    list->name = new char*[list->num];

    int i = 0;
    while (enumMon->Next(1, &moniker, nullptr) == S_OK)
    {
      list->name[i] = getCameraName(moniker);
      moniker->Release();
      i++;
    }
  }

  safeRelease(&enumMon);
  return 0;

fail:
  safeRelease(&sysDevEnum);
  safeRelease(&enumMon);
  return 1;
}

// freeCameraList frees all resources stored in cameraList*.
int freeCameraList(cameraList* list, const char** errstr)
{
  if (list->name != nullptr)
  {
    for (int i = 0; i < list->num; ++i)
    {
      delete list->name[i];
    }
    delete list->name;
  }
  return 1;
}

// selectCamera stores pointer to the selected device IMoniker* according to the configs in camera*.
int selectCamera(camera* cam, IMoniker** monikerSelected, const char** errstr)
{
  ICreateDevEnum* sysDevEnum = nullptr;
  IEnumMoniker* enumMon = nullptr;

  if (FAILED(CoCreateInstance(
          CLSID_SystemDeviceEnum, nullptr, CLSCTX_INPROC,
          IID_ICreateDevEnum, (void**)&sysDevEnum)))
  {
    *errstr = errEnumDevice;
    goto fail;
  }

  if (FAILED(sysDevEnum->CreateClassEnumerator(
          CLSID_VideoInputDeviceCategory, &enumMon, 0)))
  {
    *errstr = errEnumDevice;
    goto fail;
  }

  safeRelease(&sysDevEnum);

  if (enumMon == nullptr)
  {
    *errstr = errEnumDevice;
    return 0;
  }

  {
    IMoniker* moniker;
    while (enumMon->Next(1, &moniker, nullptr) == S_OK)
    {
      char* name = getCameraName(moniker);
      if (strcmp(cam->name, name) != 0)
      {
        free(name);
        safeRelease(&moniker);
        continue;
      }
      free(name);
      *monikerSelected = moniker;
      safeRelease(&enumMon);
      return 1;
    }
  }

  safeRelease(&enumMon);
  return 0;

fail:
  safeRelease(&sysDevEnum);
  safeRelease(&enumMon);
  return 1;
}

// listResolution stores list of the device to camera*.
int listResolution(camera* cam, const char** errstr)
{
  cam->props = nullptr;

  IMoniker* moniker = nullptr;
  IBaseFilter* captureFilter = nullptr;
  ICaptureGraphBuilder2* captureGraph = nullptr;
  IAMStreamConfig* config = nullptr;
  IPin* src = nullptr;
  LPOLESTR name;

  if (!selectCamera(cam, &moniker, errstr))
  {
    goto fail;
  }

  moniker->BindToObject(0, 0, IID_IBaseFilter, (void**)&captureFilter);
  safeRelease(&moniker);

  src = getPin(captureFilter, PINDIR_OUTPUT);
  if (src == nullptr)
  {
    *errstr = errGetConfig;
    goto fail;
  }

  // Getting IAMStreamConfig is stub on Wine. Requires real Windows.
  if (FAILED(src->QueryInterface(
          IID_IAMStreamConfig, (void**)&config)))
  {
    *errstr = errGetConfig;
    goto fail;
  }
  safeRelease(&src);

  {
    int count = 0, size = 0;
    if (FAILED(config->GetNumberOfCapabilities(&count, &size)))
    {
      *errstr = errGetConfig;
      goto fail;
    }
    cam->props = new imageProp[count];

    int iProp = 0;
    for (int i = 0; i < count; ++i)
    {
      VIDEO_STREAM_CONFIG_CAPS caps;
      AM_MEDIA_TYPE* mediaType;
      if (FAILED(config->GetStreamCaps(i, &mediaType, (BYTE*)&caps)))
        continue;

      if (mediaType->majortype != MEDIATYPE_Video ||
          mediaType->formattype != FORMAT_VideoInfo ||
          mediaType->pbFormat == nullptr)
        continue;

      VIDEOINFOHEADER* videoInfoHdr = (VIDEOINFOHEADER*)mediaType->pbFormat;
      cam->props[iProp].width = videoInfoHdr->bmiHeader.biWidth;
      cam->props[iProp].height = videoInfoHdr->bmiHeader.biHeight;
      cam->props[iProp].fcc = videoInfoHdr->bmiHeader.biCompression;
      iProp++;
    }
    cam->numProps = iProp;
  }
  safeRelease(&config);
  safeRelease(&captureGraph);
  safeRelease(&captureFilter);
  safeRelease(&moniker);
  return 0;

fail:
  safeRelease(&src);
  safeRelease(&config);
  safeRelease(&captureGraph);
  safeRelease(&captureFilter);
  safeRelease(&moniker);
  return 1;
}

// openCamera opens a camera and stores interface handler to camera*.
// camera* should be freed by freeCamera() after use.
int openCamera(camera* cam, const char** errstr)
{
  cam->grabber = nullptr;
  cam->mediaControl = nullptr;
  cam->callback = nullptr;

  IMoniker* moniker = nullptr;
  IGraphBuilder* graphBuilder = nullptr;
  IBaseFilter* captureFilter = nullptr;
  IMediaControl* mediaControl = nullptr;
  IBaseFilter* grabberFilter = nullptr;
  ISampleGrabber* grabber = nullptr;
  IBaseFilter* nullFilter = nullptr;
  IPin* src = nullptr;
  IPin* dst = nullptr;
  IPin* end = nullptr;
  IPin* nul = nullptr;

  if (!selectCamera(cam, &moniker, errstr))
  {
    goto fail;
  }
  moniker->BindToObject(0, 0, IID_IBaseFilter, (void**)&captureFilter);
  safeRelease(&moniker);

  if (FAILED(CoCreateInstance(
          CLSID_FilterGraph, nullptr, CLSCTX_INPROC,
          IID_IGraphBuilder, (void**)&graphBuilder)))
  {
    *errstr = errGraphBuilder;
    goto fail;
  }

  if (FAILED(graphBuilder->QueryInterface(
          IID_IMediaControl, (void**)&mediaControl)))
  {
    *errstr = errNoControl;
    goto fail;
  }

  if (FAILED(graphBuilder->AddFilter(captureFilter, L"capture")))
  {
    *errstr = errAddFilter;
    goto fail;
  }

  if (FAILED(CoCreateInstance(
          CLSID_SampleGrabber, nullptr, CLSCTX_INPROC,
          IID_IBaseFilter, (void**)&grabberFilter)))
  {
    *errstr = errGrabber;
    goto fail;
  }

  if (FAILED(grabberFilter->QueryInterface(IID_ISampleGrabber, (void**)&grabber)))
  {
    *errstr = errGrabber;
    goto fail;
  }

  {
    AM_MEDIA_TYPE mediaType;
    memset(&mediaType, 0, sizeof(mediaType));
    mediaType.majortype = MEDIATYPE_Video;
    // Accept any format by leaving subtype as zeros (equivalent to GUID_NULL)
    memset(&mediaType.subtype, 0, sizeof(GUID));
    mediaType.formattype = FORMAT_VideoInfo;
    // Let DirectShow negotiate format automatically
    if (FAILED(grabber->SetMediaType(&mediaType)))
    {
      *errstr = errGrabber;
      goto fail;
    }
  }

  if (FAILED(graphBuilder->AddFilter(grabberFilter, L"grabber")))
  {
    *errstr = errAddFilter;
    goto fail;
  }

  if (FAILED(CoCreateInstance(
          CLSID_NullRenderer, nullptr, CLSCTX_INPROC,
          IID_IBaseFilter, (void**)&nullFilter)))
  {
    *errstr = errTerminator;
    goto fail;
  }

  if (FAILED(graphBuilder->AddFilter(nullFilter, L"bull")))
  {
    *errstr = errAddFilter;
    goto fail;
  }

  HRESULT hr;
  src = getPin(captureFilter, PINDIR_OUTPUT);
  dst = getPin(grabberFilter, PINDIR_INPUT);
  if (src == nullptr || dst == nullptr ||
      FAILED(hr = graphBuilder->Connect(src, dst)))
  {
    *errstr = errConnectFilters;
    goto fail;
  }

  safeRelease(&src);
  safeRelease(&dst);

  // FIX: Don't connect to null renderer
  // Null renderer causes DirectShow to pause after one frame
  /*
  end = getPin(grabberFilter, PINDIR_OUTPUT);
  nul = getPin(nullFilter, PINDIR_INPUT);
  if (end == nullptr || nul == nullptr ||
      FAILED(hr = graphBuilder->Connect(end, nul)))
  {
    *errstr = errConnectFilters;
    goto fail;
  }

  safeRelease(&end);
  safeRelease(&nul);
  */

  safeRelease(&nullFilter);
  safeRelease(&captureFilter);
  safeRelease(&grabberFilter);
  safeRelease(&graphBuilder);

  {
    SampleGrabberCallback* cb = new SampleGrabberCallback(cam);
    grabber->SetCallback(cb, 1);
    cam->grabber = (void*)grabber;
    cam->mediaControl = (void*)mediaControl;
    cam->callback = (void*)cb;

    grabber->SetBufferSamples(true);
    mediaControl->Run();
  }

  return 0;

fail:
  safeRelease(&src);
  safeRelease(&dst);
  safeRelease(&end);
  safeRelease(&nul);
  safeRelease(&nullFilter);
  safeRelease(&grabber);
  safeRelease(&grabberFilter);
  safeRelease(&mediaControl);
  safeRelease(&captureFilter);
  safeRelease(&graphBuilder);
  safeRelease(&moniker);
  return 1;
}

// SampleCB is not used in this app.
HRESULT SampleGrabberCallback::SampleCB(double sampleTime, IMediaSample* sample)
{
  return S_OK;
}

// BufferCB receives image from DirectShow.
HRESULT SampleGrabberCallback::BufferCB(double sampleTime, BYTE* buf, LONG len)
{
  BYTE* gobuf = (BYTE*)cam_->buf;
  const int nPix = cam_->width * cam_->height;
  const int expectedNV12 = nPix + nPix / 2;
  const int expectedYUY2 = nPix * 2;
  
  if (abs(len - expectedNV12) <= 10)
  {
    // NV12 → I420: Copy Y, de-interleave UV
    memcpy(gobuf, buf, nPix);
    const BYTE* uv = buf + nPix;
    BYTE* u = gobuf + nPix;
    BYTE* v = u + nPix / 4;
    
    int uvSize = nPix / 4;
    int i = 0;
    
    // Process 4 pixels at a time
    for (; i < uvSize - 3; i += 4)
    {
      u[i] = uv[i * 2];
      v[i] = uv[i * 2 + 1];
      u[i + 1] = uv[(i + 1) * 2];
      v[i + 1] = uv[(i + 1) * 2 + 1];
      u[i + 2] = uv[(i + 2) * 2];
      v[i + 2] = uv[(i + 2) * 2 + 1];
      u[i + 3] = uv[(i + 3) * 2];
      v[i + 3] = uv[(i + 3) * 2 + 1];
    }
    
    // Handle remaining pixels
    for (; i < uvSize; i++)
    {
      u[i] = uv[i * 2];
      v[i] = uv[i * 2 + 1];
    }
  }
  else if (abs(len - expectedYUY2) <= 10)
  {
    // YUY2 → I420: Extract Y, subsample and average UV vertically
    BYTE* y = gobuf;
    BYTE* u = gobuf + nPix;
    BYTE* v = u + nPix / 4;
    
    for (int row = 0; row < cam_->height; row += 2)
    {
      const BYTE* src1 = buf + row * cam_->width * 2;
      const BYTE* src2 = src1 + cam_->width * 2;
      
      for (int col = 0; col < cam_->width; col += 2)
      {
        y[0] = src1[0];
        y[1] = src1[2];
        y[cam_->width] = src2[0];
        y[cam_->width + 1] = src2[2];
        y += 2;
        
        *u++ = (src1[1] + src2[1]) / 2;
        *v++ = (src1[3] + src2[3]) / 2;
        
        src1 += 4;
        src2 += 4;
      }
      y += cam_->width;
    }
  }
  else
  {
    fprintf(stderr, "Unexpected buffer size: %d (expected NV12=%d or YUY2=%d)\n", 
            len, expectedNV12, expectedYUY2);
    return S_OK;
  }

  imageCallback((size_t)cam_);
  return S_OK;
}

// freeCamera closes device and frees all resources allocated by openCamera().
void freeCamera(camera* cam)
{
  if (cam->mediaControl)
    ((IMediaControl*)cam->mediaControl)->Stop();

  safeRelease((ISampleGrabber**)&cam->grabber);
  safeRelease((IMediaControl**)&cam->mediaControl);

  if (cam->callback)
  {
    ((SampleGrabberCallback*)cam->callback)->Release();
    delete ((SampleGrabberCallback*)cam->callback);
    cam->callback = nullptr;
  }

  if (cam->props)
  {
    delete cam->props;
    cam->props = nullptr;
  }
}

// utf16Decode returns UTF-8 string from UTF-16 string.
std::string utf16Decode(LPOLESTR olestr)
{
  std::wstring wstr(olestr);
  const int len = WideCharToMultiByte(
      CP_UTF8, 0,
      wstr.data(), (int)wstr.size(),
      nullptr, 0, nullptr, nullptr);
  std::string str(len, 0);
  WideCharToMultiByte(
      CP_UTF8, 0,
      wstr.data(), (int)wstr.size(),
      (LPSTR)str.data(), len, nullptr, nullptr);
  return str;
}

// getPin is a helper to get I/O pin of DirectShow filter.
IPin* getPin(IBaseFilter* filter, PIN_DIRECTION dir)
{
  IEnumPins* enumPins;
  if (FAILED(filter->EnumPins(&enumPins)))
    return nullptr;

  IPin* pin;
  while (enumPins->Next(1, &pin, nullptr) == S_OK)
  {
    PIN_DIRECTION d;
    pin->QueryDirection(&d);
    if (d == dir)
    {
      enumPins->Release();
      return pin;
    }
    pin->Release();
  }
  enumPins->Release();
  return nullptr;
}
