package camera

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pion/mediadevices/pkg/driver"
)

func TestDiscover(t *testing.T) {
	const (
		shortName  = "video0"
		shortName2 = "video1"
		longName   = "long-device-name:0:1:2:3"
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
	discover(discovered, filepath.Join(dir, "video*"))

	drvs := driver.GetManager().Query(func(d driver.Driver) bool {
		return d.Info().DeviceType == driver.Camera
	})
	if len(drvs) != 2 {
		t.Fatalf("Expected 2 driver, got %d drivers", len(drvs))
	}

	expected := longName + LabelSeparator + shortName
	if label := drvs[0].Info().Label; label != expected {
		t.Errorf("Expected label: %s, got: %s", expected, label)
	}

	expectedNoLink := shortName2 + LabelSeparator + shortName2
	if label := drvs[1].Info().Label; label != expectedNoLink {
		t.Errorf("Expected label: %s, got: %s", expectedNoLink, label)
	}
}
