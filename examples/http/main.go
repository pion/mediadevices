// This is an example of using mediadevices to broadcast your camera through http.
// The example doesn't aim to be performant, but rather it strives to be simple.
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
	videoTrack := t.(*mediadevices.VideoTrack)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		videoReader := videoTrack.NewReader(false)
		mimeWriter := multipart.NewWriter(w)

		contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary())
		w.Header().Add("Content-Type", contentType)

		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")

		for {
			frame, release, err := videoReader.Read()
			if err == io.EOF {
				return
			}
			must(err)

			err = jpeg.Encode(&buf, frame, nil)
			// Since we're done with img, we need to release img so that that the original owner can reuse
			// this memory.
			release()
			must(err)

			partWriter, err := mimeWriter.CreatePart(partHeader)
			must(err)

			_, err = partWriter.Write(buf.Bytes())
			buf.Reset()
			must(err)
		}
	})

	fmt.Println("listening on http://localhost:1313")
	log.Println(http.ListenAndServe("localhost:1313", nil))
}
