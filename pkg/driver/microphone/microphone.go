package microphone

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
	"unsafe"

	"github.com/gen2brain/malgo"
	"github.com/pion/mediadevices/internal/logging"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/io/audio"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/mediadevices/pkg/wave"
)

const (
	maxDeviceIDLength = 20
	// TODO: should replace this with a more flexible approach
	sampleRateStep    = 1000
	initialBufferSize = 1024
)

var logger = logging.NewLogger("mediadevices/driver/microphone")
var ctx *malgo.AllocatedContext
var hostEndian binary.ByteOrder
var (
	errUnsupportedFormat = errors.New("the provided audio format is not supported")
)

type microphone struct {
	malgo.DeviceInfo
	chunkChan chan []byte
}

func init() {
	var err error
	/*
		backends := []malgo.Backend{
			malgo.BackendPulseaudio,
			malgo.BackendAlsa,
		}
	*/
	ctx, err = malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		logger.Debugf("%v\n", message)
	})
	if err != nil {
		panic(err)
	}

	devices, err := ctx.Devices(malgo.Capture)
	if err != nil {
		panic(err)
	}

	for _, device := range devices {
		// TODO: Detect default device and prioritize it
		driver.GetManager().Register(newMicrophone(device), driver.Info{
			Label:      device.ID.String(),
			DeviceType: driver.Microphone,
		})
	}

	// Decide which endian
	switch v := *(*uint16)(unsafe.Pointer(&([]byte{0x12, 0x34}[0]))); v {
	case 0x1234:
		hostEndian = binary.BigEndian
	case 0x3412:
		hostEndian = binary.LittleEndian
	default:
		panic(fmt.Sprintf("failed to determine host endianness: %x", v))
	}
}

func newMicrophone(info malgo.DeviceInfo) *microphone {
	return &microphone{
		DeviceInfo: info,
	}
}

func (m *microphone) Open() error {
	m.chunkChan = make(chan []byte, 1)
	return nil
}

func (m *microphone) Close() error {
	if m.chunkChan != nil {
		close(m.chunkChan)
		m.chunkChan = nil
	}
	return nil
}

func (m *microphone) AudioRecord(inputProp prop.Media) (audio.Reader, error) {
	var config malgo.DeviceConfig
	var callbacks malgo.DeviceCallbacks

	decoder, err := wave.NewDecoder(&wave.RawFormat{
		SampleSize:  inputProp.SampleSize,
		IsFloat:     inputProp.IsFloat,
		Interleaved: inputProp.IsInterleaved,
	})
	if err != nil {
		return nil, err
	}

	config.DeviceType = malgo.Capture
	config.PerformanceProfile = malgo.LowLatency
	config.Capture.Channels = uint32(inputProp.ChannelCount)
	config.SampleRate = uint32(inputProp.SampleRate)
	if inputProp.SampleSize == 4 && inputProp.IsFloat {
		config.Capture.Format = malgo.FormatF32
	} else if inputProp.SampleSize == 2 && !inputProp.IsFloat {
		config.Capture.Format = malgo.FormatS16
	} else {
		return nil, errUnsupportedFormat
	}

	onRecvChunk := func(_, chunk []byte, framecount uint32) {
		m.chunkChan <- chunk
	}
	callbacks.Data = onRecvChunk

	device, err := malgo.InitDevice(ctx.Context, config, callbacks)
	if err != nil {
		return nil, err
	}

	err = device.Start()
	if err != nil {
		return nil, err
	}

	return audio.ReaderFunc(func() (wave.Audio, func(), error) {
		chunk, ok := <-m.chunkChan
		if !ok {
			device.Stop()
			device.Uninit()
			return nil, func() {}, io.EOF
		}

		decodedChunk, err := decoder.Decode(hostEndian, chunk, inputProp.ChannelCount)
		// FIXME: the decoder should also fill this information
		decodedChunk.(*wave.Float32Interleaved).Size.SamplingRate = inputProp.SampleRate
		return decodedChunk, func() {}, err
	}), nil
}

func (m *microphone) Properties() []prop.Media {
	var supportedProps []prop.Media
	logger.Debug("Querying properties")

	var isBigEndian bool
	// miniaudio only uses the host endian
	if hostEndian == binary.BigEndian {
		isBigEndian = true
	}

	for ch := m.MinChannels; ch <= m.MaxChannels; ch++ {
		for sampleRate := m.MinSampleRate; sampleRate <= m.MaxSampleRate; sampleRate += sampleRateStep {
			for i := 0; i < int(m.FormatCount); i++ {
				format := m.Formats[i]

				supportedProp := prop.Media{
					Audio: prop.Audio{
						ChannelCount: int(ch),
						SampleRate:   int(sampleRate),
						IsBigEndian:  isBigEndian,
						// miniaudio only supports interleaved at the moment
						IsInterleaved: true,
					},
				}

				switch malgo.FormatType(format) {
				case malgo.FormatF32:
					supportedProp.SampleSize = 4
					supportedProp.IsFloat = true
				case malgo.FormatS16:
					supportedProp.SampleSize = 2
					supportedProp.IsFloat = false
				}

				supportedProps = append(supportedProps, supportedProp)
			}
		}
	}

	// FIXME: remove this hardcoded value. Malgo doesn't support "ma_context_get_device_info" API yet. The above iterations
	//        will always return nothing as of now
	supportedProps = append(supportedProps, prop.Media{
		Audio: prop.Audio{
			Latency:       time.Millisecond * 20,
			ChannelCount:  1,
			SampleRate:    48000,
			SampleSize:    4,
			IsFloat:       true,
			IsBigEndian:   isBigEndian,
			IsInterleaved: true,
		},
	})
	return supportedProps
}
