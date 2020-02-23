#include <iostream>
#include <unistd.h>

#include <dshow.h>
#include <qedit.h>

#include "camera_windows.hpp"
#include "_cgo_export.h"

void printErr(HRESULT hr)
{
  char buf[128];
  AMGetErrorTextA(hr, buf, 128);
  fprintf(stderr, "%s\n", buf);
}

int openCamera(camera* cam, const char** errstr)
{
  cam->grabber = nullptr;
  cam->mediaControl = nullptr;
  cam->callback = nullptr;

  ICreateDevEnum* sysDevEnum = nullptr;
  IEnumMoniker* enumMon = nullptr;
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

  {
    IMoniker* moniker;
    if (enumMon->Next(1, &moniker, nullptr) == S_OK)
    {
      LPOLESTR name;
      if (FAILED(moniker->GetDisplayName(nullptr, nullptr, &name)))
      {
        *errstr = errEnumDevice;
        safeRelease(&moniker);
        goto fail;
      }
      std::string nameStr = utf16Decode(name);
      cam->name = (char*)malloc(nameStr.size() + 1);
      memcpy(cam->name, nameStr.c_str(), nameStr.size() + 1);

      moniker->BindToObject(0, 0, IID_IBaseFilter, (void**)&captureFilter);
      safeRelease(&moniker);
    }
  }

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
    mediaType.subtype = MEDIASUBTYPE_RGB24;
    mediaType.formattype = FORMAT_VideoInfo;

    VIDEOINFOHEADER videoInfoHdr;
    memset(&videoInfoHdr, 0, sizeof(VIDEOINFOHEADER));
    videoInfoHdr.bmiHeader.biSize = sizeof(BITMAPINFOHEADER);
    videoInfoHdr.bmiHeader.biWidth = cam->width;
    videoInfoHdr.bmiHeader.biHeight = cam->height;
    videoInfoHdr.bmiHeader.biPlanes = 1;
    videoInfoHdr.bmiHeader.biBitCount = 24;
    videoInfoHdr.bmiHeader.biCompression = BI_RGB;
    mediaType.pbFormat = (BYTE*)&videoInfoHdr;
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

  src = getPin(captureFilter, PINDIR_OUTPUT);
  dst = getPin(grabberFilter, PINDIR_INPUT);
  if (src == nullptr || dst == nullptr ||
      FAILED(graphBuilder->Connect(src, dst)))
  {
    *errstr = errConnectFilters;
    goto fail;
  }

  safeRelease(&src);
  safeRelease(&dst);

  end = getPin(grabberFilter, PINDIR_OUTPUT);
  nul = getPin(nullFilter, PINDIR_INPUT);
  if (end == nullptr || nul == nullptr ||
      FAILED(graphBuilder->Connect(end, nul)))
  {
    *errstr = errConnectFilters;
    goto fail;
  }

  safeRelease(&end);
  safeRelease(&nul);

  safeRelease(&nullFilter);
  safeRelease(&captureFilter);
  safeRelease(&grabberFilter);
  safeRelease(&graphBuilder);

  {
    AM_MEDIA_TYPE actualMediaType;
    grabber->GetConnectedMediaType(&actualMediaType);
    VIDEOINFOHEADER* videoInfoHdr = (VIDEOINFOHEADER*)actualMediaType.pbFormat;
  }

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
  safeRelease(&enumMon);
  safeRelease(&sysDevEnum);
  return 1;
}

HRESULT SampleGrabberCallback::SampleCB(double sampleTime, IMediaSample* sample)
{
  return S_OK;
}

HRESULT SampleGrabberCallback::BufferCB(double sampleTime, BYTE* buf, LONG len)
{
  BYTE* gobuf = (BYTE*)cam_->buf;
  const int nPix = cam_->width * cam_->height;
  if (len < nPix * 3)
  {
    fprintf(stderr, "Wrong frame buffer size: %v < %v\n", len, nPix * 3);
    return S_OK;
  }
  // Windows image is upside down
  int i = 0;
  for (int y = 0; y < cam_->height; ++y)
  {
    int j = (cam_->height - y - 1) * cam_->width * 3;
    for (int x = 0; x < cam_->width; ++x)
    {
      gobuf[i + 0] = buf[j + 2];
      gobuf[i + 1] = buf[j + 1];
      gobuf[i + 2] = buf[j + 0];
      gobuf[i + 3] = 0xff;
      i += 4;
      j += 3;
    }
  }
  imageCallback((size_t)cam_);
  return S_OK;
}

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

  if (cam->name)
    free(cam->name);
}

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
