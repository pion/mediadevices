package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/driver/camera"
)

func main() {
	fmt.Println("This example demonstrates query-based camera device discovery on Darwin.")
	fmt.Println("The background observer automatically updates the manager's device list")
	fmt.Println("when cameras are connected or disconnected.")

	camera.StartObserver()
	defer camera.DestroyObserver()

	scanner := bufio.NewScanner(os.Stdin)
	queryCount := 0

	queryDevices(0)

	for {
		fmt.Print("\nPress Enter to query (or 'q' to exit): ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if strings.ToLower(input) == "q" {
			break
		}

		queryCount++
		queryDevices(queryCount)
	}
}

func queryDevices(count int) {
	if count > 0 {
		fmt.Printf("Query #%d\n", count)
	}

	devices := driver.GetManager().Query(driver.FilterVideoRecorder())

	if len(devices) == 0 {
		fmt.Println("No video devices found.")
	} else {
		fmt.Printf("Found %d video device(s):\n", len(devices))
		for i, d := range devices {
			info := d.Info()
			fmt.Printf("  %d. %s [%s]\n", i+1, info.Name, info.Label)
		}
	}
}
