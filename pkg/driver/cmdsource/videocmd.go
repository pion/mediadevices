package cmdsource

import (
	"fmt"
	"image"
	"io"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
)

type videoCmdSource struct {
	cmdSource
	showStdErr bool
	label      string
}

func AddVideoCmdSource(label string, command string, mediaProperties []prop.Media, readTimeout uint32, showStdErr bool) error {
	videoCmdSource := &videoCmdSource{
		cmdSource:  newCmdSource(command, mediaProperties, readTimeout),
		label:      label,
		showStdErr: showStdErr,
	}
	if len(videoCmdSource.cmdArgs) == 0 || videoCmdSource.cmdArgs[0] == "" {
		return errInvalidCommand // no command specified
	}

	err := driver.GetManager().Register(videoCmdSource, driver.Info{
		Label:      label,
		DeviceType: driver.CmdSource,
		Priority:   driver.PriorityNormal,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *videoCmdSource) VideoRecord(inputProp prop.Media) (video.Reader, error) {
	getFrameSize, ok := frame.FrameSizeMap[inputProp.FrameFormat]
	if !ok {
		return nil, errUnsupportedFormat
	}
	frameSize := getFrameSize(inputProp.Width, inputProp.Height)

	decoder, err := frame.NewDecoder(inputProp.FrameFormat)
	if err != nil {
		return nil, err
	}

	if c.showStdErr {
		// get the command's standard error
		stdErr, err := c.execCmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		// send standard error to the console as debug logs prefixed with "{command} stdErr >"
		go c.logStdIoWithPrefix(fmt.Sprintf("%s stdErr> ", c.label+":"+c.cmdArgs[0]), stdErr)
	}
	// get the command's standard output
	stdOut, err := c.execCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// add environment variables to the command for each media property
	c.addEnvVarsFromStruct(inputProp.Video, c.showStdErr)

	// start the command
	if err := c.execCmd.Start(); err != nil {
		return nil, err
	}

	var buf []byte = make([]byte, frameSize)
	doneChan := make(chan error)
	// fmt.Printf("frameSize: %d\n", frameSize)
	r := video.ReaderFunc(func() (img image.Image, release func(), err error) {
		go func() {
			if _, err := io.ReadFull(stdOut, buf); err == io.ErrUnexpectedEOF {
				doneChan <- io.EOF
			} else if err != nil {
				doneChan <- err
			}
			doneChan <- nil
		}()

		select {
		case err := <-doneChan:
			if err != nil {
				return nil, nil, err
			} else {
				return decoder.Decode(buf, inputProp.Width, inputProp.Height)
			}
		case <-time.After(time.Duration(c.readTimeout) * time.Second):
			return nil, func() {}, errReadTimeout
		}
	})

	return r, nil
}
