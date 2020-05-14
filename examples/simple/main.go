package main

import (
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/prop"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter
)

func main() {
	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(p *prop.Media) {},
	})
	if err != nil {
		panic(err)
	}

	t := s.GetVideoTracks()[0]
	defer t.Stop()
	videoTrack := t.(*mediadevices.VideoTrack)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		videoReader := videoTrack.NewReader()
		mimeWriter := multipart.NewWriter(w)

		contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary())
		w.Header().Add("Content-Type", contentType)

		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")

		for {
			frame, err := videoReader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				panic(err)
			}

			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				panic(err)
			}

			err = jpeg.Encode(partWriter, frame, nil)
			if err != nil {
				panic(err)
			}
		}
	})

	log.Println(http.ListenAndServe(":1313", nil))
}
