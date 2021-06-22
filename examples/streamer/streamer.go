package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/pion/webrtc/v3"

	md "github.com/pion/mediadevices"
	//h264 "github.com/pion/mediadevices/pkg/codec/x264"
	h264 "github.com/pion/mediadevices/pkg/codec/openh264"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/driver"
	"github.com/pion/mediadevices/pkg/prop"
)

const (
	mtu = 1000
)

func must(err error) bool {
	if err != nil {
		//fmt.Printf("ERROR: %s\n", err.Error())
		return true
	}
	return false
}

type Stream struct {
	Addr     string
	Active   bool
	DeviceId string
	Started  time.Time
	Sent     uint64
	Stop     chan interface{}
}

func (s Stream) String() string {
	return fmt.Sprintf("%s active: %t, device id: %s, start time: %s, bytes sent: %s", s.Addr, s.Active, s.DeviceId, strconv.FormatInt(s.Started.Unix(), 10), strconv.FormatUint(s.Sent, 10))
}

func (s Stream) Serialize() string {
	return fmt.Sprintf("s/%s/%t/%s/%s/%s", s.Addr, s.Active, s.DeviceId, strconv.FormatInt(s.Started.Unix(), 10), strconv.FormatUint(s.Sent, 10))
}

var Streams = make(map[string]*Stream)
var Devices = make(map[string]*md.MediaDeviceInfo)

func init() {
	Enumerate()
}

func State() string {
	state := ""
	for _, stream := range Streams {
		if stream != nil {
			state += stream.Serialize()
			state += "\n"
		}
		fmt.Printf("%s\n", state)
	}
	for _, device := range Devices {
		if device != nil {
			state += device.Label
			state += "\n"
		}
	}

	return state
}

func Enumerate() map[string]*md.MediaDeviceInfo {
	drivers := driver.GetManager().Query(func(drv driver.Driver) bool {
		return true
	})
	devices := make(map[string]*md.MediaDeviceInfo, len(drivers))

	sending := ""

	for _, drv := range drivers {
		var kind md.MediaDeviceType
		deviceID := drv.ID()
		switch {
		case driver.FilterVideoRecorder()(drv):
			kind = md.VideoInput
		case driver.FilterAudioRecorder()(drv):
			kind = md.AudioInput
		default:
			continue
		}
		drvInfo := drv.Info()
		deviceInfo := md.MediaDeviceInfo{DeviceID: deviceID, Kind: kind, Label: drvInfo.Label, Name: drvInfo.Name, Manufacturer: drvInfo.Manufacturer, ModelID: drvInfo.ModelID, DeviceType: drvInfo.DeviceType}
		devices[deviceID] = &deviceInfo
		fmt.Printf("%s\n", deviceInfo.String())

		if driver.FilterVideoRecorder()(drv) && sending == "" {
			sending = deviceID
		}
	}

	// TODO: check if device not available anymore
	Devices = devices

	go Start("192.168.1.20:15000", sending)

	return devices
}

func Start(addr, deviceId string) {
	device, ok := Devices[deviceId]
	if !ok {
		fmt.Printf("[%s->%s] no such device\n", deviceId, addr)
		return
	}

	a, err := net.ResolveUDPAddr("udp", addr)
	if must(err) {
		fmt.Printf("[%s->%s] bad addr\n", deviceId, addr)
		return
	}

	conn, err := net.DialUDP("udp", nil, a)
	if must(err) {
		fmt.Printf("[%s->%s] can't connect to addr\n", deviceId, addr)
		return
	}

	constraints := md.MediaStreamConstraints{}

	var codecName string
	var payloadType uint8

	switch device.Kind {
	case md.VideoInput:
		h264Params, err := h264.NewParams()
		if must(err) {
			fmt.Printf("[%s->%s] can't make x264 params\n", deviceId, addr)
			return
		}
		h264Params.BitRate = 3_000_000
		//h264Params.Preset = x264.PresetUltrafast
		h264Params.KeyFrameInterval = 1

		constraints.Codec = md.NewCodecSelector(md.WithVideoEncoders(&h264Params))
		constraints.Video = func(c *md.MediaTrackConstraints) {
			c.DeviceID = prop.StringExact(deviceId)
			c.Width = prop.IntRanged{Min: 1280, Max: 1920, Ideal: 1280}
		}

		codecName = webrtc.MimeTypeH264
		payloadType = 125 // corresponding value from sdpOffer to Kurento
	case md.AudioInput:
		opusParams, err := opus.NewParams()
		must(err)
		if must(err) {
			fmt.Printf("[%s->%s] can't make Opus params\n", deviceId, addr)
			return
		}
		opusParams.BitRate = 256_000

		constraints.Audio = func(c *md.MediaTrackConstraints) { c.DeviceID = prop.StringExact(deviceId) }
		constraints.Codec = md.NewCodecSelector(md.WithAudioEncoders(&opusParams))

		codecName = webrtc.MimeTypeOpus
		payloadType = 96
	default:
		return
	}

	mediaStream, err := md.GetUserMedia(constraints)
	if must(err) {
		fmt.Printf("[%s->%s] can't get media stream: %s\n", deviceId, addr, err.Error())
		return
	}

	track := mediaStream.GetTracks()[0]
	defer track.Close()

	rtpReader, err := track.NewRTPReader(codecName, rand.Uint32(), mtu) //nolint:gosec
	if must(err) {
		fmt.Printf("[%s->%s] can't make rtp reader: %s\n", deviceId, addr, err.Error())
		return
	}
	defer rtpReader.Close()

	stop := make(chan interface{})
	defer close(stop)

	stream := Stream{DeviceId: deviceId, Addr: addr, Started: time.Now(), Stop: stop}
	Streams[addr] = &stream
	fmt.Printf("[%s->%s] new stream: %s\n", deviceId, addr, stream.String())

	buff := make([]byte, mtu)
	for {
		select {
		case <-stop:
			stream.Active = false
			fmt.Printf("[%s->%s] stop stream: %s\n", deviceId, addr, stream.String())
			return
		default:
			pkts, release, err := rtpReader.Read()

			if must(err) {
				return
			}

			stream.Active = true

			for _, pkt := range pkts {
				pkt.PayloadType = payloadType
				n, err := pkt.MarshalTo(buff)
				if must(err) {
					continue
				}

				b, err := conn.Write(buff[:n])
				if must(err) {
					continue
				}
				stream.Sent += uint64(b)
				//fmt.Printf("sent: %d\n", b)
			}
			release()
		}
	}
}

func Stop(addr string) {
	stream, ok := Streams[addr]
	if ok {
		close(stream.Stop)
	}
}
