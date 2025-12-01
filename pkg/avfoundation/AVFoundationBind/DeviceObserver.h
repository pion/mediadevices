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

#ifndef DEVICEOBSERVER_H
#define DEVICEOBSERVER_H

#include "AVFoundationBind.h"

typedef const char* STATUS;

typedef struct {
    char uid[MAX_DEVICE_UID_CHARS + 1]; // +1 for null terminator
    char name[MAX_DEVICE_NAME_CHARS + 1];
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

// Start observing device events (notifications will be delivered via the run loop)
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
