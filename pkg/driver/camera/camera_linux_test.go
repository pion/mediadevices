package camera

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/pion/mediadevices/pkg/driver"
)

func TestDiscover(t *testing.T) {
	const (
		shortName  = "unittest-video0"
		shortName2 = "unittest-video1"
		longName   = "unittest-long-device-name:0:1:2:3"
	)

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	byPathDir := filepath.Join(dir, "v4l", "by-path")
	if err := os.MkdirAll(byPathDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, shortName), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, shortName2), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(
		filepath.Join(dir, shortName),
		filepath.Join(byPathDir, longName),
	); err != nil {
		t.Fatal(err)
	}

	discovered := make(map[string]struct{})
	discover(discovered, filepath.Join(byPathDir, "*"))
	discover(discovered, filepath.Join(dir, "unittest-video*"))

	drvs := driver.GetManager().Query(func(d driver.Driver) bool {
		// Ignore real cameras.
		return d.Info().DeviceType == driver.Camera && strings.Contains(d.Info().Label, "unittest")
	})
	if len(drvs) != 2 {
		t.Fatalf("Expected 2 driver, got %d drivers", len(drvs))
	}

	labels := []string{
		drvs[0].Info().Label,
		drvs[1].Info().Label,
	}

	// Returned drivers are unordered. Sort to get static result.
	sort.Sort(sort.StringSlice(labels))

	expected := longName + LabelSeparator + shortName
	if label := labels[0]; label != expected {
		t.Errorf("Expected label: %s, got: %s", expected, label)
	}

	expectedNoLink := shortName2 + LabelSeparator + shortName2
	if label := labels[1]; label != expectedNoLink {
		t.Errorf("Expected label: %s, got: %s", expectedNoLink, label)
	}
}

func TestDiscoverByID(t *testing.T) {
	const (
		shortName  = "id-unittest-video0"
		shortName2 = "id-unittest-video1"
		longName   = "id-unittest-long-device-name:0:1:2:3"
	)

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	byIdDir := filepath.Join(dir, "v4l", "by-id")
	if err := os.MkdirAll(byIdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, shortName), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, shortName2), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(
		filepath.Join(dir, shortName),
		filepath.Join(byIdDir, longName),
	); err != nil {
		t.Fatal(err)
	}

	discovered := make(map[string]struct{})
	discover(discovered, filepath.Join(byIdDir, "*"))
	discover(discovered, filepath.Join(dir, "id-unittest-video*"))

	drvs := driver.GetManager().Query(func(d driver.Driver) bool {
		// Ignore real cameras.
		return d.Info().DeviceType == driver.Camera && strings.Contains(d.Info().Label, "id-unittest")
	})
	if len(drvs) != 2 {
		t.Fatalf("Expected 2 driver, got %d drivers", len(drvs))
	}

	labels := []string{
		drvs[0].Info().Label,
		drvs[1].Info().Label,
	}

	// Returned drivers are unordered. Sort to get static result.
	sort.Sort(sort.StringSlice(labels))

	expected := longName + LabelSeparator + shortName
	if label := labels[0]; label != expected {
		t.Errorf("Expected label: %s, got: %s", expected, label)
	}

	expectedNoLink := shortName2 + LabelSeparator + shortName2
	if label := labels[1]; label != expectedNoLink {
		t.Errorf("Expected label: %s, got: %s", expectedNoLink, label)
	}
}

func TestGetCameraReadTimeout(t *testing.T) {
	var expected uint32 = 5
	value := getCameraReadTimeout()
	if value != expected {
		t.Errorf("Expected: %d, got: %d", expected, value)
	}

	envVarName := "PION_MEDIADEVICES_CAMERA_READ_TIMEOUT"
	os.Setenv(envVarName, "text")
	value = getCameraReadTimeout()
	if value != expected {
		t.Errorf("Expected: %d, got: %d", expected, value)
	}

	os.Setenv(envVarName, "-1")
	value = getCameraReadTimeout()
	if value != expected {
		t.Errorf("Expected: %d, got: %d", expected, value)
	}

	os.Setenv(envVarName, "1")
	expected = 1
	value = getCameraReadTimeout()
	if value != expected {
		t.Errorf("Expected: %d, got: %d", expected, value)
	}
}
