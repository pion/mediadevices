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
#import "DeviceObserver.h"


extern void goDeviceEventCallback(void *pUserData, int eventType, DeviceInfo *pDevice);

void deviceEventBridge(void *pUserData, DeviceEventType eventType, DeviceInfo *pDevice) {
    goDeviceEventCallback(pUserData, (int)eventType, pDevice);
}

@interface DeviceObserverDelegate : NSObject {
    DeviceEventCallback mCallback;
    void *mUserData;
    AVCaptureDeviceDiscoverySession *mDiscoverySession;
    BOOL mObserving;
}
@end

@implementation DeviceObserverDelegate

- (instancetype)initWithCallback:(DeviceEventCallback)callback userData:(void *)pUserData {
    self = [super init];
    if (self) {
        mCallback = callback;
        mUserData = pUserData;
        mObserving = NO;

        NSArray *refDeviceTypes = @[
            AVCaptureDeviceTypeBuiltInWideAngleCamera,
            AVCaptureDeviceTypeExternal
        ];

        mDiscoverySession = [[AVCaptureDeviceDiscoverySession
            discoverySessionWithDeviceTypes:refDeviceTypes
            mediaType:AVMediaTypeVideo
            position:AVCaptureDevicePositionUnspecified] retain];
    }
    return self;
}

- (void)startObserving {
    if (mObserving) return;

    [mDiscoverySession addObserver:self
                        forKeyPath:@"devices"
                           options:(NSKeyValueObservingOptionOld | NSKeyValueObservingOptionNew)
                           context:nil];

    mObserving = YES;
}

- (void)stopObserving {
    if (!mObserving) return;

    [mDiscoverySession removeObserver:self forKeyPath:@"devices"];
    mObserving = NO;
}

- (void)observeValueForKeyPath:(NSString *)keyPath
                      ofObject:(id)object
                        change:(NSDictionary<NSKeyValueChangeKey,id> *)change
                       context:(void *)pContext {

    if (![keyPath isEqualToString:@"devices"]) return;

    NSArray<AVCaptureDevice *> *refOldDevices = change[NSKeyValueChangeOldKey];
    NSArray<AVCaptureDevice *> *refNewDevices = change[NSKeyValueChangeNewKey];

    if ([refOldDevices isKindOfClass:[NSNull class]]) refOldDevices = @[];
    if ([refNewDevices isKindOfClass:[NSNull class]]) refNewDevices = @[];

    // Build sets of device UIDs for comparison
    NSMutableSet *refOldUIDs = [NSMutableSet set];
    NSMutableDictionary *refOldDeviceMap = [NSMutableDictionary dictionary];
    for (AVCaptureDevice *refDevice in refOldDevices) {
        [refOldUIDs addObject:refDevice.uniqueID];
        refOldDeviceMap[refDevice.uniqueID] = refDevice;
    }

    NSMutableSet *refNewUIDs = [NSMutableSet set];
    NSMutableDictionary *refNewDeviceMap = [NSMutableDictionary dictionary];
    for (AVCaptureDevice *refDevice in refNewDevices) {
        [refNewUIDs addObject:refDevice.uniqueID];
        refNewDeviceMap[refDevice.uniqueID] = refDevice;
    }

    // Find added devices
    NSMutableSet *refAddedUIDs = [refNewUIDs mutableCopy];
    [refAddedUIDs minusSet:refOldUIDs];

    // Find removed devices
    NSMutableSet *refRemovedUIDs = [refOldUIDs mutableCopy];
    [refRemovedUIDs minusSet:refNewUIDs];

    // Notify about added devices
    for (NSString *uid in refAddedUIDs) {
        AVCaptureDevice *refDevice = refNewDeviceMap[uid];
        DeviceInfo info;
        memset(&info, 0, sizeof(info));
        strncpy(info.uid, refDevice.uniqueID.UTF8String, MAX_DEVICE_UID_CHARS - 1);
        strncpy(info.name, refDevice.localizedName.UTF8String, MAX_DEVICE_NAME_CHARS - 1);

        if (mCallback) {
            mCallback(mUserData, DeviceEventConnected, &info);
        }
    }

    // Notify about removed devices
    for (NSString *uid in refRemovedUIDs) {
        AVCaptureDevice *refDevice = refOldDeviceMap[uid];
        DeviceInfo info;
        memset(&info, 0, sizeof(info));
        strncpy(info.uid, refDevice.uniqueID.UTF8String, MAX_DEVICE_UID_CHARS - 1);
        strncpy(info.name, refDevice.localizedName.UTF8String, MAX_DEVICE_NAME_CHARS - 1);

        if (mCallback) {
            mCallback(mUserData, DeviceEventDisconnected, &info);
        }
    }

    [refAddedUIDs release];
    [refRemovedUIDs release];
}

- (void)dealloc {
    [self stopObserving];
    [mDiscoverySession release];
    [super dealloc];
}

@end

// Global observer instance
static DeviceObserverDelegate *refObserver = nil;

STATUS DeviceObserverInit(DeviceEventCallback callback, void *pUserData) {
    @autoreleasepool {
        if (refObserver != nil) {
            return "observer already initialized";
        }

        refObserver = [[DeviceObserverDelegate alloc] initWithCallback:callback userData:pUserData];
        if (refObserver == nil) {
            return "failed to create observer";
        }

        return STATUS_OK;
    }
}

STATUS DeviceObserverStart(void) {
    @autoreleasepool {
        if (refObserver == nil) {
            return "observer not initialized";
        }

        [refObserver startObserving];
        return STATUS_OK;
    }
}

STATUS DeviceObserverStop(void) {
    @autoreleasepool {
        if (refObserver == nil) {
            return "observer not initialized";
        }

        [refObserver stopObserving];
        return STATUS_OK;
    }
}

STATUS DeviceObserverDestroy(void) {
    @autoreleasepool {
        if (refObserver == nil) {
            return "observer not initialized";
        }

        [refObserver stopObserving];
        [refObserver release];
        refObserver = nil;

        return STATUS_OK;
    }
}

STATUS DeviceObserverGetDevices(DeviceInfo *pDevices, int *pCount) {
    @autoreleasepool {
        if (pDevices == NULL || pCount == NULL) {
            return "invalid arguments";
        }

        // Use discovery session for device enumeration
        NSArray *refDeviceTypes = @[
            AVCaptureDeviceTypeBuiltInWideAngleCamera,
            AVCaptureDeviceTypeExternal
        ];

        AVCaptureDeviceDiscoverySession *refSession = [AVCaptureDeviceDiscoverySession
            discoverySessionWithDeviceTypes:refDeviceTypes
            mediaType:AVMediaTypeVideo
            position:AVCaptureDevicePositionUnspecified];

        int i = 0;
        for (AVCaptureDevice *refDevice in refSession.devices) {
            if (i >= MAX_DEVICES) break;

            memset(&pDevices[i], 0, sizeof(DeviceInfo));
            strncpy(pDevices[i].uid, refDevice.uniqueID.UTF8String, MAX_DEVICE_UID_CHARS - 1);
            strncpy(pDevices[i].name, refDevice.localizedName.UTF8String, MAX_DEVICE_NAME_CHARS - 1);
            i++;
        }

        *pCount = i;
        return STATUS_OK;
    }
}

STATUS DeviceObserverRunFor(double seconds) {
    @autoreleasepool {
        // Add a timer to keep the run loop alive
        NSTimer *refTimer = [NSTimer scheduledTimerWithTimeInterval:seconds
                                                             target:[NSDate class]
                                                           selector:@selector(date)
                                                           userInfo:nil
                                                            repeats:NO];
        [[NSRunLoop currentRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:seconds]];
        [refTimer invalidate];
        return STATUS_OK;
    }
}
