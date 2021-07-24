record:
	 gst-launch-1.0 -v pulsesrc ! audio/x-raw, format=S16LE, rate=48000, channels=2 ! filesink location=audio.raw


play: audio.raw
	gst-launch-1.0 filesrc location=$^ ! rawaudioparse use-sink-caps=false \
			 format=pcm pcm-format=s16le sample-rate=48000 num-channels=2 ! \
			 audioconvert ! audioresample ! autoaudiosink
