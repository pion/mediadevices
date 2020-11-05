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
	"os"

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
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s host:port\n", os.Args[0])
		return
	}
	dest := os.Args[1]

	mediaStream, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			constraint.Width = prop.Int(600)
			constraint.Height = prop.Int(400)
		},
	})
	must(err)

	track := mediaStream.GetVideoTracks()[0]
	videoTrack := track.(*mediadevices.VideoTrack)
	defer videoTrack.Close()

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

	fmt.Printf("listening on %s\n", dest)
	log.Println(http.ListenAndServe(dest, nil))
}
