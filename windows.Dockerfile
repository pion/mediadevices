FROM fedora

ENV PYTHONUNBUFFERED=1 \
    # Somehow libopus.dll is installed under /usr/local/bin, so we need to add this
    # path to WINEPATH so that wine can find this library
    WINEPATH=/usr/local/bin \
    # Go and C default environments
    GOOS=windows \
    GOARCH=amd64 \
    CGO_ENABLED=1 \
    GO111MODULE=on \
    CC=x86_64-w64-mingw32-gcc \
    CXX=x86_64-w64-mingw32-g++

COPY scripts /usr/bin/ 

RUN dnf install -y \
        diffutils \
        golang \
        make \
        mingw64-gcc \
        mingw64-gcc-c++ \
        mingw64-opus && \
        nasm \
        pkg-config \
        python3 \
        wine \
    dnf clean all && \
    install-dependencies && \
    rm -rf /tmp/*

VOLUME /go/src/github.com/pion/mediadevices
WORKDIR /go/src/github.com/pion/mediadevices