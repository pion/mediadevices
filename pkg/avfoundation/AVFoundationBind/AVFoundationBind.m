// MIT License
// 
// Copyright (c) 2019-2020 Pion
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Naming Convention (let "name" as an actual variable name):
//   - mName: "name" is a member of an Objective C object
//   - pName: "name" is a C pointer
//   - refName: "name" is an Objective C object reference

#import <Foundation/Foundation.h>
#import <AVFoundation/AVFoundation.h>
#import "AVFoundationBind.h"
#include <string.h>

#define CHK(condition, status) \
    do { \
        if(!(condition)) { \
            retStatus = status; \
            goto cleanup; \
        } \
    } while(0)

#define CHK_STATUS(status) \
    do { \
        if(status != STATUS_OK) { \
            retStatus = status; \
            goto cleanup; \
        } \
    } while(0)

@interface VideoDataDelegate : NSObject<AVCaptureVideoDataOutputSampleBufferDelegate>

@property (readonly) AVBindDataCallback mCallback;
@property (readonly) void *mPUserData;

- (void)captureOutput:(AVCaptureOutput *)captureOutput
didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer
       fromConnection:(AVCaptureConnection *)connection;

@end

@implementation VideoDataDelegate

- (id) init: (AVBindDataCallback) callback
withUserData: (void*) pUserData {
    self = [super init];
    _mCallback = callback;
    _mPUserData = pUserData;

    return self;
}

- (void)captureOutput:(AVCaptureOutput *)captureOutput
didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer
       fromConnection:(AVCaptureConnection *)connection {
    if (CMSampleBufferGetNumSamples(sampleBuffer) != 1 ||
        !CMSampleBufferIsValid(sampleBuffer) ||
        !CMSampleBufferDataIsReady(sampleBuffer)) {
      return;
    }
    
    CVImageBufferRef imageBuffer = CMSampleBufferGetImageBuffer(sampleBuffer);
    if (imageBuffer == NULL) {
      return;
    }
    
    imageBuffer = CVBufferRetain(imageBuffer);
    CVReturn ret =
        CVPixelBufferLockBaseAddress(imageBuffer, kCVPixelBufferLock_ReadOnly);
    if (ret != kCVReturnSuccess) {
      return;
    }
    
    size_t heightY = CVPixelBufferGetHeightOfPlane(imageBuffer, 0);
    size_t bytesPerRowY = CVPixelBufferGetBytesPerRowOfPlane(imageBuffer, 0);
    
    size_t heightUV = CVPixelBufferGetHeightOfPlane(imageBuffer, 1);
    size_t bytesPerRowUV = CVPixelBufferGetBytesPerRowOfPlane(imageBuffer, 1);
    
    int len = (int)((heightY * bytesPerRowY) + (2 * heightUV * bytesPerRowUV));
    void *buf = CVPixelBufferGetBaseAddressOfPlane(imageBuffer, 0);
    _mCallback(_mPUserData, buf, len);
    
    CVPixelBufferUnlockBaseAddress(imageBuffer, 0);
    CVBufferRelease(imageBuffer);
}

@end

@interface AudioDataDelegate : NSObject<AVCaptureAudioDataOutputSampleBufferDelegate>

@property (readonly) AVBindDataCallback mCallback;

- (void)captureOutput:(AVCaptureOutput *)captureOutput
didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer
       fromConnection:(AVCaptureConnection *)connection;

@end

@implementation AudioDataDelegate

- (id) init: (AVBindDataCallback) callback {
    self = [super init];
    _mCallback = callback;
    return self;
}

- (void)captureOutput:(AVCaptureOutput *)captureOutput
didOutputSampleBuffer:(CMSampleBufferRef)sampleBuffer
       fromConnection:(AVCaptureConnection *)connection {
    // TODO
}

@end

STATUS frameFormatToFourCC(AVBindFrameFormat format, FourCharCode *pFourCC) {
    STATUS retStatus = STATUS_OK;
    switch (format) {
        case AVBindFrameFormatNV21:
            *pFourCC = kCVPixelFormatType_420YpCbCr8BiPlanarFullRange;
            break;
        case AVBindFrameFormatUYVY:
            *pFourCC = kCVPixelFormatType_422YpCbCr8;
            break;
        // TODO: Add the rest of frame formats
        default:
            retStatus = STATUS_UNSUPPORTED_FRAME_FORMAT;
    }
    return retStatus;
}

STATUS frameFormatFromFourCC(FourCharCode fourCC, AVBindFrameFormat *pFormat) {
    STATUS retStatus = STATUS_OK;
    switch (fourCC) {
         case kCVPixelFormatType_420YpCbCr8BiPlanarFullRange:
             *pFormat = AVBindFrameFormatNV21;
             break;
         case kCVPixelFormatType_422YpCbCr8:
             *pFormat = AVBindFrameFormatUYVY;
             break;
         // TODO: Add the rest of frame formats
         default:
             retStatus = STATUS_UNSUPPORTED_FRAME_FORMAT;
     }
    return retStatus;
}


STATUS AVBindDevices(AVBindMediaType mediaType, PAVBindDevice *ppDevices, int *pLen) {
    static AVBindDevice devices[MAX_DEVICES];
    STATUS retStatus = STATUS_OK;
    NSAutoreleasePool *refPool = [[NSAutoreleasePool alloc] init];
    CHK(mediaType == AVBindMediaTypeVideo || mediaType == AVBindMediaTypeAudio, STATUS_UNSUPPORTED_MEDIA_TYPE);
    CHK(ppDevices != NULL && pLen != NULL, STATUS_NULL_ARG);
    
    PAVBindDevice pDevice;
    AVMediaType _mediaType = mediaType == AVBindMediaTypeVideo ? AVMediaTypeVideo : AVMediaTypeAudio;
    NSArray *refAllTypes = @[
        AVCaptureDeviceTypeBuiltInWideAngleCamera,
        AVCaptureDeviceTypeBuiltInMicrophone,
        AVCaptureDeviceTypeExternalUnknown
    ];
    AVCaptureDeviceDiscoverySession *refSession = [AVCaptureDeviceDiscoverySession
        discoverySessionWithDeviceTypes: refAllTypes
        mediaType: _mediaType
        position: AVCaptureDevicePositionUnspecified];
    
    int i = 0;
    for (AVCaptureDevice *refDevice in refSession.devices) {
        if (i >= MAX_DEVICES) {
            break;
        }
        
        pDevice = devices + i;
        strncpy(pDevice->uid, refDevice.uniqueID.UTF8String, MAX_DEVICE_UID_CHARS);
        pDevice->uid[MAX_DEVICE_UID_CHARS] = '\0';
        i++;
    }
    
    *ppDevices = devices;
    *pLen = i;
    
cleanup:
    [refPool drain];
    return retStatus;
}

struct AVBindSession {
    AVBindDevice device;
    AVCaptureSession *refCaptureSession;
    AVBindMediaProperty properties[MAX_PROPERTIES];
};


STATUS AVBindSessionInit(AVBindDevice device, PAVBindSession *ppSessionResult) {
    STATUS retStatus = STATUS_OK;
    CHK(ppSessionResult != NULL, STATUS_NULL_ARG);
    PAVBindSession pSession = malloc(sizeof(AVBindSession));
    pSession->device = device;
    pSession->refCaptureSession = NULL;
    *ppSessionResult = pSession;
    
cleanup:
    return retStatus;
}

STATUS AVBindSessionFree(PAVBindSession *ppSession) {
    STATUS retStatus = STATUS_OK;
    CHK(ppSession != NULL, STATUS_NULL_ARG);
    PAVBindSession pSession = *ppSession;
    if (pSession->refCaptureSession != NULL) {
        [pSession->refCaptureSession release];
        pSession->refCaptureSession = NULL;
    }
    free(pSession);
    *ppSession = NULL;

cleanup:
    return retStatus;
}

STATUS AVBindSessionOpen(PAVBindSession pSession,
                  AVBindMediaProperty property,
                  AVBindDataCallback dataCallback,
                  void *pUserData) {
    STATUS retStatus = STATUS_OK;
    NSAutoreleasePool *refPool = [[NSAutoreleasePool alloc] init];
    CHK(pSession != NULL && dataCallback != NULL, STATUS_NULL_ARG);
    
    AVCaptureDeviceInput *refInput;
    NSError *refErr = NULL;
    NSString *refUID = [NSString stringWithUTF8String: pSession->device.uid];
    AVCaptureDevice *refDevice = [AVCaptureDevice deviceWithUniqueID: refUID];
    
    refInput = [[AVCaptureDeviceInput alloc] initWithDevice: refDevice error: &refErr];
    CHK(refErr == NULL, STATUS_DEVICE_INIT_FAILED);
    
    AVCaptureSession *refCaptureSession = [[AVCaptureSession alloc] init];
    refCaptureSession.sessionPreset = AVCaptureSessionPresetMedium;
    [refCaptureSession addInput: refInput];

    if ([refDevice hasMediaType: AVMediaTypeVideo]) {
        VideoDataDelegate *pDelegate = [[VideoDataDelegate alloc]
                                        init: dataCallback
                                        withUserData: pUserData];
        
        AVCaptureVideoDataOutput *pOutput = [[AVCaptureVideoDataOutput alloc] init];
        FourCharCode fourCC;
        CHK_STATUS(frameFormatToFourCC(property.frameFormat, &fourCC));

        pOutput.videoSettings = @{
            (id)kCVPixelBufferWidthKey: @(property.width),
            (id)kCVPixelBufferHeightKey: @(property.height),
            (id)kCVPixelBufferPixelFormatTypeKey: @(fourCC),
        };
        pOutput.alwaysDiscardsLateVideoFrames = YES;
        dispatch_queue_t queue =
            dispatch_queue_create("captureQueue", DISPATCH_QUEUE_SERIAL);
        [pOutput setSampleBufferDelegate:pDelegate queue:queue];
        [refCaptureSession addOutput: pOutput];
    } else {
        // TODO: implement audio pipeline
    }
    
    pSession->refCaptureSession = [refCaptureSession retain];
    [refCaptureSession startRunning];
    
cleanup:
    [refPool drain];
    return retStatus;
}


STATUS AVBindSessionClose(PAVBindSession pSession) {
    STATUS retStatus = STATUS_OK;
    CHK(pSession != NULL, STATUS_NULL_ARG);
    CHK(pSession->refCaptureSession != NULL, STATUS_OK);
    
    [pSession->refCaptureSession stopRunning];
    [pSession->refCaptureSession release];
    pSession->refCaptureSession = NULL;
        
cleanup:
    return retStatus;
}

STATUS AVBindSessionProperties(PAVBindSession pSession, PAVBindMediaProperty *ppProperties, int *pLen) {
    STATUS retStatus = STATUS_OK;
    NSAutoreleasePool *refPool = [[NSAutoreleasePool alloc] init];
    CHK(pSession != NULL && ppProperties != NULL && pLen != NULL, STATUS_NULL_ARG);
    
    NSString *refDeviceUID = [NSString stringWithUTF8String: pSession->device.uid];
    AVCaptureDevice *refDevice = [AVCaptureDevice deviceWithUniqueID: refDeviceUID];
    FourCharCode fourCC;
    CMVideoFormatDescriptionRef videoFormat;
    CMVideoDimensions videoDimensions;

    memset(pSession->properties, 0, sizeof(pSession->properties));
    PAVBindMediaProperty pProperty = pSession->properties;
    int len = 0;
    for (AVCaptureDeviceFormat *refFormat in refDevice.formats) {
        // TODO: Probably gives a warn to the user
        if (len >= MAX_PROPERTIES) {
            break;
        }
        
        if ([refFormat.mediaType isEqual:AVMediaTypeVideo]) {
            fourCC = CMFormatDescriptionGetMediaSubType(refFormat.formatDescription);
            if (frameFormatFromFourCC(fourCC, &pProperty->frameFormat) != STATUS_OK) {
                continue;
            }
            
            videoFormat = (CMVideoFormatDescriptionRef) refFormat.formatDescription;
            videoDimensions = CMVideoFormatDescriptionGetDimensions(videoFormat);
            pProperty->height = videoDimensions.height;
            pProperty->width = videoDimensions.width;
        } else {
            // TODO: Get audio properties
        }
        
        pProperty++;
        len++;
    }
    
    *ppProperties = pSession->properties;
    *pLen = len;
    
cleanup:
    
    [refPool drain];
    return retStatus;
}
