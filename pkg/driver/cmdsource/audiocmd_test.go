package cmdsource

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/pion/mediadevices/pkg/prop"
)

// const minInt32 int32 = -2147483648
const maxInt32 int32 = 2147418112

func ValueInRange(input int64, min int64, max int64) bool {
	return input >= min && input <= max
}

var ffmpegAudioFormatMap = map[string]prop.Audio{
	"f32be": {
		IsFloat:       true,
		IsBigEndian:   true,
		IsInterleaved: true,
		SampleSize:    4, // 4*8 = 32 bits
	},
	"f32le": {
		IsFloat:       true,
		IsBigEndian:   false,
		IsInterleaved: true,
		SampleSize:    4, // 4*8 = 32 bits
	},
	"s16be": {
		IsFloat:       false,
		IsBigEndian:   true,
		IsInterleaved: true,
		SampleSize:    2, // 2*8 = 16 bits
	},
	"s16le": {
		IsFloat:       false,
		IsBigEndian:   false,
		IsInterleaved: true,
		SampleSize:    2, // 2*8 = 16 bits
	},
}

func RunAudioCmdTest(t *testing.T, freq int, duration float32, sampleRate int, channelCount int, sampleBufferSize int, format string) {
	command := fmt.Sprintf("ffmpeg -f lavfi -i sine=frequency=%d:duration=%f:sample_rate=%d -af arealtime,volume=8 -ac %d -f %s -", freq, duration, sampleRate, channelCount, format)
	timeout := uint32(10) // 10 seconds
	audioProps := ffmpegAudioFormatMap[format]
	audioProps.ChannelCount = channelCount
	audioProps.SampleRate = sampleRate
	properties := []prop.Media{
		{
			DeviceID: "ffmpeg audio",
			Audio:    audioProps,
		},
	}

	fmt.Println("Testing audio source command: " + command)

	// Make sure ffmpeg is installed before continuting the test
	err := exec.Command("ffmpeg", "-version").Run()
	if err != nil {
		t.Skip("ffmpeg command not found in path. Skipping test. Err: ", err)
	}

	// Create the audio cmd source
	audioCmdSource := &audioCmdSource{
		cmdSource:         newCmdSource(command, properties, timeout),
		bufferSampleCount: sampleBufferSize,
	}

	// check if the command split correctly
	if audioCmdSource.cmdArgs[0] != "ffmpeg" {
		t.Fatal("command parsing failed")
	}

	err = audioCmdSource.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer audioCmdSource.Close()

	// Get the audio reader from the audio cmd source
	reader, err := audioCmdSource.AudioRecord(properties[0])
	if err != nil {
		t.Fatal(err)
	}

	// Read the first chunk
	chunk, _, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}

	// Check if the chunk has the correct number of channels
	if chunk.ChunkInfo().Channels != channelCount {
		t.Errorf("chunk has incorrect number of channels")
	}

	// Check if the chunk has the correct sample rate
	if chunk.ChunkInfo().SamplingRate != sampleRate {
		t.Errorf("chunk has incorrect sample rate")
	}

	println("Samples:")
	for i := 0; i < chunk.ChunkInfo().Len; i++ {
		fmt.Printf("%d\n", chunk.At(i, 0).Int())
	}

	// Test the waveform value at the 1st sample in the chunk (should be "near" 0, because it is a sine wave)
	sampleIdx := 1
	channelIdx := 0
	min := int64(0)
	max := int64(267911168)
	if value := chunk.At(sampleIdx, channelIdx).Int(); ValueInRange(value, min, max) == false {
		t.Errorf("chan #%d, chunk #%d has incorrect value, expected %d-%d, got %d", channelIdx, sampleIdx, min, max, value)
	}

	// Test the waveform value at the 1/4th the way through the sine wave (should be near max in 32 bit int)
	samplesPerSinewaveCycle := sampleRate / freq
	sampleIdx = samplesPerSinewaveCycle / 4 // 1/4 of a cycle
	channelIdx = 0
	min = int64(maxInt32) - int64(267911168)
	max = 0xFFFFFFFF
	if value := chunk.At(sampleIdx, channelIdx).Int(); ValueInRange(value, min, max) == false {
		t.Errorf("chan #%d, chunk #%d has incorrect value, expected %d-%d, got %d", channelIdx, sampleIdx, min, max, value)
	}

	err = audioCmdSource.Close()
	if err != nil && err.Error() != "exit status 255" { // ffmpeg returns 255 when it is stopped normally
		t.Fatal(err)
	}

	audioCmdSource.Close() // should not panic
}

func TestWavIntLeAudioCmdOut(t *testing.T) {
	RunAudioCmdTest(t, 440, 1, 44100, 1, 256, "s16le")
}

func TestWavIntBeAudioCmdOut(t *testing.T) {
	RunAudioCmdTest(t, 120, 1, 44101, 1, 256, "s16be")
}

func TestWavFloatLeAudioCmdOut(t *testing.T) {
	RunAudioCmdTest(t, 220, 1, 44102, 1, 256, "f32le")
}

func TestWavFloatBeAudioCmdOut(t *testing.T) {
	RunAudioCmdTest(t, 110, 1, 44103, 1, 256, "f32be")
}
