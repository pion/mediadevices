package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pion/mediadevices/pkg/avfoundation"
)

func main() {
	fmt.Println("=== AVFoundation Background Observer Example ===")

	// Start the background observer
	if err := avfoundation.StartObserver(); err != nil {
		fmt.Printf("Error starting observer: %v\n", err)
		os.Exit(1)
	}
	defer avfoundation.StopObserver()

	fmt.Println("Observer started. Monitoring for device changes...")

	// Get initial devices
	lastDevices, err := avfoundation.GetDevices()
	if err != nil {
		fmt.Printf("Error getting devices: %v\n", err)
		return
	}
	printDevices(lastDevices)

	// Poll for changes
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			currentDevices, err := avfoundation.GetDevices()
			if err != nil {
				fmt.Printf("Error getting devices: %v\n", err)
				continue
			}

			if detectChanges(lastDevices, currentDevices) {
				fmt.Println("\nChange detected!")
				printDevices(currentDevices)
				lastDevices = currentDevices
			}
		case <-sigChan:
			fmt.Println("\nShutting down...")
			return
		}
	}
}

func printDevices(devices []avfoundation.Device) {
	fmt.Printf("Found %d device(s):\n", len(devices))
	for i, d := range devices {
		fmt.Printf("  %d. %s [%s]\n", i+1, d.Name, d.UID)
	}
}

func detectChanges(old, new []avfoundation.Device) bool {
	oldMap := make(map[string]avfoundation.Device)
	for _, d := range old {
		oldMap[d.UID] = d
	}

	newMap := make(map[string]avfoundation.Device)
	for _, d := range new {
		newMap[d.UID] = d
	}

	changed := false

	// Check for added
	for uid, d := range newMap {
		if _, exists := oldMap[uid]; !exists {
			fmt.Printf("  + ADDED: %s [%s]\n", d.Name, d.UID)
			changed = true
		}
	}

	// Check for removed
	for uid, d := range oldMap {
		if _, exists := newMap[uid]; !exists {
			fmt.Printf("  - REMOVED: %s [%s]\n", d.Name, d.UID)
			changed = true
		}
	}

	return changed
}
