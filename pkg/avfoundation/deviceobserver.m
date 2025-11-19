#import <Foundation/Foundation.h>
#import <AVFoundation/AVFoundation.h>
#import "deviceobserver.h"

static STATUS STATUS_OK = (STATUS)NULL;

extern void goDeviceEventCallback(void *userData, int eventType, DeviceInfo *device);

void deviceEventBridge(void *userData, DeviceEventType eventType, DeviceInfo *device) {
    goDeviceEventCallback(userData, (int)eventType, device);
}

@interface DeviceObserverDelegate : NSObject {
    DeviceEventCallback _callback;
    void *_userData;
    AVCaptureDeviceDiscoverySession *_discoverySession;
    BOOL _observing;
}
@end

@implementation DeviceObserverDelegate

- (instancetype)initWithCallback:(DeviceEventCallback)callback userData:(void *)userData {
    self = [super init];
    if (self) {
        _callback = callback;
        _userData = userData;
        _observing = NO;

        NSArray *deviceTypes = @[
            AVCaptureDeviceTypeBuiltInWideAngleCamera,
            AVCaptureDeviceTypeExternal
        ];

        _discoverySession = [[AVCaptureDeviceDiscoverySession
            discoverySessionWithDeviceTypes:deviceTypes
            mediaType:AVMediaTypeVideo
            position:AVCaptureDevicePositionUnspecified] retain];
    }
    return self;
}

- (void)startObserving {
    if (_observing) return;

    [_discoverySession addObserver:self
                        forKeyPath:@"devices"
                           options:(NSKeyValueObservingOptionOld | NSKeyValueObservingOptionNew)
                           context:nil];

    _observing = YES;
}

- (void)stopObserving {
    if (!_observing) return;

    [_discoverySession removeObserver:self forKeyPath:@"devices"];
    _observing = NO;
}

- (void)observeValueForKeyPath:(NSString *)keyPath
                      ofObject:(id)object
                        change:(NSDictionary<NSKeyValueChangeKey,id> *)change
                       context:(void *)context {

    if (![keyPath isEqualToString:@"devices"]) return;

    NSArray<AVCaptureDevice *> *oldDevices = change[NSKeyValueChangeOldKey];
    NSArray<AVCaptureDevice *> *newDevices = change[NSKeyValueChangeNewKey];

    if ([oldDevices isKindOfClass:[NSNull class]]) oldDevices = @[];
    if ([newDevices isKindOfClass:[NSNull class]]) newDevices = @[];

    // Build sets of device UIDs for comparison
    NSMutableSet *oldUIDs = [NSMutableSet set];
    NSMutableDictionary *oldDeviceMap = [NSMutableDictionary dictionary];
    for (AVCaptureDevice *device in oldDevices) {
        [oldUIDs addObject:device.uniqueID];
        oldDeviceMap[device.uniqueID] = device;
    }

    NSMutableSet *newUIDs = [NSMutableSet set];
    NSMutableDictionary *newDeviceMap = [NSMutableDictionary dictionary];
    for (AVCaptureDevice *device in newDevices) {
        [newUIDs addObject:device.uniqueID];
        newDeviceMap[device.uniqueID] = device;
    }

    // Find added devices
    NSMutableSet *addedUIDs = [newUIDs mutableCopy];
    [addedUIDs minusSet:oldUIDs];

    // Find removed devices
    NSMutableSet *removedUIDs = [oldUIDs mutableCopy];
    [removedUIDs minusSet:newUIDs];

    // Notify about added devices
    for (NSString *uid in addedUIDs) {
        AVCaptureDevice *device = newDeviceMap[uid];
        DeviceInfo info;
        memset(&info, 0, sizeof(info));
        strncpy(info.uid, device.uniqueID.UTF8String, MAX_DEVICE_UID_CHARS - 1);
        strncpy(info.name, device.localizedName.UTF8String, MAX_DEVICE_NAME_CHARS - 1);

        if (_callback) {
            _callback(_userData, DeviceEventConnected, &info);
        }
    }

    // Notify about removed devices
    for (NSString *uid in removedUIDs) {
        AVCaptureDevice *device = oldDeviceMap[uid];
        DeviceInfo info;
        memset(&info, 0, sizeof(info));
        strncpy(info.uid, device.uniqueID.UTF8String, MAX_DEVICE_UID_CHARS - 1);
        strncpy(info.name, device.localizedName.UTF8String, MAX_DEVICE_NAME_CHARS - 1);

        if (_callback) {
            _callback(_userData, DeviceEventDisconnected, &info);
        }
    }

    [addedUIDs release];
    [removedUIDs release];
}

- (void)dealloc {
    [self stopObserving];
    [_discoverySession release];
    [super dealloc];
}

@end

// Global observer instance
static DeviceObserverDelegate *gObserver = nil;

STATUS DeviceObserverInit(DeviceEventCallback callback, void *userData) {
    @autoreleasepool {
        if (gObserver != nil) {
            return "observer already initialized";
        }

        gObserver = [[DeviceObserverDelegate alloc] initWithCallback:callback userData:userData];
        if (gObserver == nil) {
            return "failed to create observer";
        }

        return STATUS_OK;
    }
}

STATUS DeviceObserverStart(void) {
    @autoreleasepool {
        if (gObserver == nil) {
            return "observer not initialized";
        }

        [gObserver startObserving];
        return STATUS_OK;
    }
}

STATUS DeviceObserverStop(void) {
    @autoreleasepool {
        if (gObserver == nil) {
            return "observer not initialized";
        }

        [gObserver stopObserving];
        return STATUS_OK;
    }
}

STATUS DeviceObserverDestroy(void) {
    @autoreleasepool {
        if (gObserver == nil) {
            return "observer not initialized";
        }

        [gObserver stopObserving];
        [gObserver release];
        gObserver = nil;

        return STATUS_OK;
    }
}

STATUS DeviceObserverGetDevices(DeviceInfo *devices, int *count) {
    @autoreleasepool {
        if (devices == NULL || count == NULL) {
            return "invalid arguments";
        }

        // Use discovery session for device enumeration
        NSArray *deviceTypes = @[
            AVCaptureDeviceTypeBuiltInWideAngleCamera,
            AVCaptureDeviceTypeExternal
        ];

        AVCaptureDeviceDiscoverySession *session = [AVCaptureDeviceDiscoverySession
            discoverySessionWithDeviceTypes:deviceTypes
            mediaType:AVMediaTypeVideo
            position:AVCaptureDevicePositionUnspecified];

        int i = 0;
        for (AVCaptureDevice *device in session.devices) {
            if (i >= MAX_DEVICES) break;

            memset(&devices[i], 0, sizeof(DeviceInfo));
            strncpy(devices[i].uid, device.uniqueID.UTF8String, MAX_DEVICE_UID_CHARS - 1);
            strncpy(devices[i].name, device.localizedName.UTF8String, MAX_DEVICE_NAME_CHARS - 1);
            i++;
        }

        *count = i;
        return STATUS_OK;
    }
}

STATUS DeviceObserverRunFor(double seconds) {
    @autoreleasepool {
        // Add a timer to keep the run loop alive
        NSTimer *timer = [NSTimer scheduledTimerWithTimeInterval:seconds
                                                          target:[NSDate class]
                                                        selector:@selector(date)
                                                        userInfo:nil
                                                         repeats:NO];
        [[NSRunLoop currentRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:seconds]];
        [timer invalidate];
        return STATUS_OK;
    }
}
