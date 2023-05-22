package cmdsource

import (
	"fmt"
	"image/color"
	"os/exec"
	"testing"

	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

// var ycbcrWhite := color.YCbCr{235, 128, 128}
var ycbcrPink = color.YCbCr{Y: 198, Cb: 123, Cr: 155}
var ffmpegFrameFormatMap = map[frame.Format]string{
	frame.FormatI420: "yuv420p",
	frame.FormatNV21: "nv21",
	frame.FormatNV12: "nv12",
	frame.FormatYUY2: "yuyv422",
	frame.FormatUYVY: "uyvy422",
	frame.FormatZ16:  "gray",
}

func RunVideoCmdTest(t *testing.T, width int, height int, frameRate float32, frameFormat frame.Format, inputColor string, expectedColor color.Color) {

	command := fmt.Sprintf("ffmpeg -hide_banner -f lavfi -i color=c=%s:size=%dx%d:rate=%f -vf realtime -f rawvideo -pix_fmt %s -", inputColor, width, height, frameRate, ffmpegFrameFormatMap[frameFormat])

	// Example using injected environment variables instead of hardcoding the command:
	// command := fmt.Sprintf("sh -c 'ffmpeg -hide_banner -f lavfi -i color=c=%s:size=\"$MEDIA_DEVICES_Width\"x\"$MEDIA_DEVICES_Height\":rate=\"$MEDIA_DEVICES_FrameRate\" -vf realtime -f rawvideo -pix_fmt %s -'", inputColor, ffmpegFrameFormatMap[frameFormat])

	timeout := uint32(10) // 10 seconds
	properties := []prop.Media{
		{
			DeviceID: "ffmpeg 1",
			Video: prop.Video{
				Width:       width,
				Height:      height,
				FrameFormat: frameFormat,
				FrameRate:   frameRate,
			},
		},
	}

	fmt.Println("Testing video source command: " + command)

	// Make sure ffmpeg is installed before continuting the test
	err := exec.Command("ffmpeg", "-version").Run()
	if err != nil {
		t.Skip("ffmpeg command not found in path. Skipping test. Err: ", err)
	}

	// Create a new video command source
	videoCmdSource := &videoCmdSource{
		cmdSource:  newCmdSource(command, properties, timeout),
		showStdErr: true,
		label:      "test_source",
	}

	// if videoCmdSource.cmdArgs[0] != "ffmpeg" {
	// 	t.Fatal("command parsing failed")
	// }

	err = videoCmdSource.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer videoCmdSource.Close()

	reader, err := videoCmdSource.VideoRecord(properties[0])
	if err != nil {
		t.Fatal(err)
	}
	img, _, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != width || img.Bounds().Dy() != height {
		t.Logf("image resolution output is not correct, got: (%d, %d) | expected: (%d, %d)", img.Bounds().Dx(), img.Bounds().Dy(), width, height)
		t.Fatal()
	}

	// test color at upper left corner
	if pxlColor := img.At(0, 0); pxlColor != expectedColor {
		t.Errorf("Image pixel output at 0,0 is not correct. Got: %+v | Expected: %+v", pxlColor, expectedColor)
	}

	// test color at center of image
	x := width / 2
	y := height / 2
	if pxlColor := img.At(x, y); pxlColor != expectedColor {
		t.Errorf("Image pixel output at %d,%d is not correct. Got: %+v | Expected: %+v", x, y, pxlColor, expectedColor)
	}

	// test color at lower right corner
	x = width - 1
	y = height - 1
	if pxlColor := img.At(x, y); pxlColor != expectedColor {
		t.Errorf("Image pixel output at %d,%d is not correct. Got: %+v | Expected: %+v", x, y, pxlColor, expectedColor)
	}

	err = videoCmdSource.Close()
	if err != nil {
		t.Fatal(err)
	}
	videoCmdSource.Close() // should not panic
	println()              // add a new line to separate the output from the end of the test
}

func TestI420VideoCmdOut(t *testing.T) {
	RunVideoCmdTest(t, 640, 480, 30, frame.FormatI420, "pink", ycbcrPink)
}

func TestNV21VideoCmdOut(t *testing.T) {
	RunVideoCmdTest(t, 640, 480, 30, frame.FormatNV21, "pink", ycbcrPink)
}

func TestNV12VideoCmdOut(t *testing.T) {
	RunVideoCmdTest(t, 640, 480, 30, frame.FormatNV12, "pink", ycbcrPink)
}

func TestYUY2VideoCmdOut(t *testing.T) {
	RunVideoCmdTest(t, 640, 480, 30, frame.FormatYUY2, "pink", ycbcrPink)
}

func TestUYVYVideoCmdOut(t *testing.T) {
	RunVideoCmdTest(t, 640, 480, 30, frame.FormatUYVY, "pink", ycbcrPink)
}

func TestZ16VideoCmdOut(t *testing.T) {
	RunVideoCmdTest(t, 640, 480, 30, frame.FormatZ16, "white", color.Gray16{65535})
}
