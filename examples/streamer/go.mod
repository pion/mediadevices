module github.com/pion/mediadevices/examples/streamer

go 1.16

require (
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/pion/mediadevices v0.2.0
	github.com/pion/webrtc/v3 v3.0.29
)

replace github.com/pion/mediadevices v0.2.0 => ../../
