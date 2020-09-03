package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/pion/mediadevices"
	mio "github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/prop"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			constraint.Width = prop.Int(600)
			constraint.Height = prop.Int(400)
		},
	})
	must(err)

	t := s.GetVideoTracks()[0]
	defer t.Stop()
	videoTrack := t.(*mediadevices.VideoTrack)
	videoSource := videoTrack.Source()
	jpegBroadcaster := mio.NewBroadcaster(mio.ReaderFunc(func() (interface{}, error) {
		var buf bytes.Buffer

		img, err := videoSource.Read()
		if err != nil {
			return nil, err
		}

		err = jpeg.Encode(&buf, img, nil)
		return buf.Bytes(), err
	}))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		videoReader := jpegBroadcaster.NewReader(func(src interface{}) interface{} { return src })
		mimeWriter := multipart.NewWriter(w)

		contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary())
		w.Header().Add("Content-Type", contentType)

		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")

		for {
			frame, err := videoReader.Read()
			if err == io.EOF {
				return
			}
			must(err)

			partWriter, err := mimeWriter.CreatePart(partHeader)
			must(err)

			data, _ := frame.([]byte)
			partWriter.Write(data)
		}
	})

	fmt.Println("listening on http://localhost:1313")
	log.Println(http.ListenAndServe("localhost:1313", nil))
}
