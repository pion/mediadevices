package frame

import (
	"bytes"
	"image/jpeg"
	"testing"
)

func TestAddMotionDht(t *testing.T) {
	uninitializedHuffmanTableFrame, err := jpeg.Decode(bytes.NewReader(UninitializedHuffmanTable))

	// Decode fails with an uninitialized Huffman table error for sample input
	expectedErrorMessage := "invalid JPEG format: uninitialized Huffman table"
	if err.Error() != expectedErrorMessage {
		t.Fatalf("Wrong decode error result,\nexpected:\n%+v\ngot:\n%+v", expectedErrorMessage, err)
	}

	// Decode passes after adding default Huffman table to
	defaultHuffmanTableFrame, err := jpeg.Decode(bytes.NewReader(addMotionDht(UninitializedHuffmanTable)))
	if err != nil {
		t.Fatalf("Expected decode function to pass after adding default Huffman table. Failed with %v\n", err)
	}

	// Adding default Huffman table to a valid frame without a Huffman table changes the table
	if uninitializedHuffmanTableFrame == defaultHuffmanTableFrame {
		t.Fatalf("Expected addMotionDht to update frame. Instead returned original frame")
	}

	// Check that an improperly constructed frame does not get updated by addMotionDht
	randomBytes := []byte{1, 2, 3, 4}
	frame1, err := jpeg.Decode(bytes.NewReader(randomBytes))
	if err == nil {
		t.Fatalf("Expected decode function to fail with random bytes but passed.")
	}

	frame2, err := jpeg.Decode(bytes.NewReader(addMotionDht(randomBytes)))
	if err == nil {
		t.Fatalf("Expected decode function to fail with random bytes but passed.")
	}

	if frame1 != frame2 {
		t.Fatalf("addMotionDht updated the frame despite being improperly constructed")
	}
}

func TestDecodeMJPEG(t *testing.T) {
	_, _, err := decodeMJPEG(UninitializedHuffmanTable, 640, 480)
	if err != nil {
		t.Fatalf("Expected decode function to pass. Failed with %v\n", err)
	}
}
