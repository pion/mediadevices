#ifndef DEVICEOBSERVER_H
#define DEVICEOBSERVER_H

#define MAX_DEVICES 16
#define MAX_DEVICE_UID_CHARS 128
#define MAX_DEVICE_NAME_CHARS 128

typedef const char* STATUS;

typedef struct {
    char uid[MAX_DEVICE_UID_CHARS];
    char name[MAX_DEVICE_NAME_CHARS];
} DeviceInfo;

typedef enum {
    DeviceEventConnected = 0,
    DeviceEventDisconnected = 1
} DeviceEventType;

// Callback function type for device events
// userData: user-provided context pointer
// eventType: connected or disconnected
// device: device info
typedef void (*DeviceEventCallback)(void *userData, DeviceEventType eventType, DeviceInfo *device);

// Initialize the device observer with a callback
// Returns NULL on success, error string on failure
STATUS DeviceObserverInit(DeviceEventCallback callback, void *userData);

// Start observing device events (runs on current thread's run loop)
STATUS DeviceObserverStart(void);

// Stop observing device events
STATUS DeviceObserverStop(void);

// Cleanup the device observer
STATUS DeviceObserverDestroy(void);

// Get current list of video devices
// devices: output array (must have space for MAX_DEVICES)
// count: output count of devices found
STATUS DeviceObserverGetDevices(DeviceInfo *devices, int *count);

// Run the run loop for a specified duration (in seconds)
// This allows the observer to receive notifications
STATUS DeviceObserverRunFor(double seconds);

#endif // DEVICEOBSERVER_H
