package cmdsource

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

type audioCmdSource struct {
	cmdSource
	bufferSampleCount int
	showStdErr        bool
	label             string
}

func AddAudioCmdSource(label string, command string, mediaProperties []prop.Media, readTimeout uint32, sampleBufferSize int, showStdErr bool) error {
	audioCmdSource := &audioCmdSource{
		cmdSource:         newCmdSource(command, mediaProperties, readTimeout),
		bufferSampleCount: sampleBufferSize,
		label:             label,
		showStdErr:        showStdErr,
	}
	if len(audioCmdSource.cmdArgs) == 0 || audioCmdSource.cmdArgs[0] == "" {
		return errInvalidCommand // no command specified
	}

	// register this audio source with the driver manager
	err := driver.GetManager().Register(audioCmdSource, driver.Info{
		Label:      label,
		DeviceType: driver.CmdSource,
		Priority:   driver.PriorityNormal,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *audioCmdSource) AudioRecord(inputProp prop.Media) (audio.Reader, error) {
	decoder, err := wave.NewDecoder(&wave.RawFormat{
		SampleSize:  inputProp.SampleSize,
		IsFloat:     inputProp.IsFloat,
		Interleaved: inputProp.IsInterleaved,
	})

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
		go c.logStdIoWithPrefix(fmt.Sprintf("%s stderr > ", c.label+":"+c.cmdArgs[0]), stdErr)
	}

	// get the command's standard output
	stdOut, err := c.execCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// add environment variables to the command for each media property
	c.addEnvVarsFromStruct(inputProp.Audio, c.showStdErr)

	// start the command
	if err := c.execCmd.Start(); err != nil {
		return nil, err
	}

	// claclulate the sample size and chunk buffer size (as a multple of the sample size)
	sampleSize := inputProp.ChannelCount * inputProp.SampleSize
	chunkSize := c.bufferSampleCount * sampleSize
	var endienness binary.ByteOrder = binary.LittleEndian
	if inputProp.IsBigEndian {
		endienness = binary.BigEndian
	}

	var chunkBuf []byte = make([]byte, chunkSize)
	doneChan := make(chan error)
	r := audio.ReaderFunc(func() (chunk wave.Audio, release func(), err error) {
		go func() {
			if _, err := io.ReadFull(stdOut, chunkBuf); err == io.ErrUnexpectedEOF {
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
				decodedChunk, err := decoder.Decode(endienness, chunkBuf, inputProp.ChannelCount)
				if err != nil {
					return nil, nil, err
				}
				// FIXME: the decoder should also fill this information
				switch decodedChunk := decodedChunk.(type) {
				case *wave.Float32Interleaved:
					decodedChunk.Size.SamplingRate = inputProp.SampleRate
				case *wave.Int16Interleaved:
					decodedChunk.Size.SamplingRate = inputProp.SampleRate
				default:
					panic("unsupported format")
				}
				return decodedChunk, func() {}, err
			}
		case <-time.After(time.Duration(c.readTimeout) * time.Second):
			return nil, func() {}, errReadTimeout
		}
	})

	return r, nil
}
