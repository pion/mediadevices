FROM dockercore/golang-cross

RUN apt-get update -qq \
  && apt-get install -y \
    g++-mingw-w64 \
    nasm \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

VOLUME /go/src/github.com/pion/mediadevices
WORKDIR /go/src/github.com/pion/mediadevices
