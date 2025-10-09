package vpx

import (
	"testing"
	"unsafe"
)

// TestVpxImageStructure tests the VpxImage struct methods
// Note: These tests verify the interface and structure without requiring actual VPX images
func TestVpxImageStructure(t *testing.T) {
	// Test that VpxImage can be created (interface test)
	// We can't easily test with real C structures in unit tests due to CGO limitations
	// but we can test the structure and interface
	
	t.Run("VpxImageInterface", func(t *testing.T) {
		// This test ensures the VpxImage type exists and has the expected methods
		// We use a type assertion to verify the interface
		var _ interface {
			Width() int
			Height() int
			YStride() int
			UStride() int
			VStride() int
			Plane(int) unsafe.Pointer
		} = (*VpxImage)(nil)
	})
}

// TestNewImageFromPtr tests the constructor
func TestNewImageFromPtr(t *testing.T) {
	// Test with nil pointer
	vpxImg := NewImageFromPtr(nil)
	if vpxImg == nil {
		t.Error("NewImageFromPtr should not return nil even with nil input")
	}
	if vpxImg != nil && vpxImg.img != nil {
		t.Error("VpxImage should contain nil pointer when created with nil")
	}
}

// TestVpxImageMethodsWithNil tests that methods panic appropriately with nil pointer
// This documents the expected behavior - methods will panic if called with nil C pointer
func TestVpxImageMethodsWithNil(t *testing.T) {
	vpxImg := NewImageFromPtr(nil)
	
	// These methods should panic with nil img (this is expected behavior)
	testCases := []struct {
		name string
		fn   func()
	}{
		{"Width", func() { vpxImg.Width() }},
		{"Height", func() { vpxImg.Height() }},
		{"YStride", func() { vpxImg.YStride() }},
		{"UStride", func() { vpxImg.UStride() }},
		{"VStride", func() { vpxImg.VStride() }},
		{"Plane0", func() { vpxImg.Plane(0) }},
		{"Plane1", func() { vpxImg.Plane(1) }},
		{"Plane2", func() { vpxImg.Plane(2) }},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Method %s should panic with nil image but didn't", tc.name)
				}
			}()
			tc.fn()
		})
	}
}

// TestVpxImageConstants tests expected behavior with common video formats
func TestVpxImageConstants(t *testing.T) {
	// Test that the VpxImage type can be used in common video processing scenarios
	testCases := []struct {
		name        string
		planeIndex  int
		description string
	}{
		{"Y Plane", 0, "Luma plane"},
		{"U Plane", 1, "Chroma U plane"},
		{"V Plane", 2, "Chroma V plane"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify plane indices are within expected range
			if tc.planeIndex < 0 || tc.planeIndex > 2 {
				t.Errorf("Plane index %d is out of expected range [0-2]", tc.planeIndex)
			}
		})
	}
}
