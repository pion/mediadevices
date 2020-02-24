FROM mediadevices-base

ENV TARGET_OS=windows \
    PYTHONUNBUFFERED=1 \
    # Somehow libopus.dll is installed under /usr/local/bin, so we need to add this
    # path to WINEPATH so that wine can find this library
    WINEPATH=/usr/local/bin \
    # Go and C default environments
    GOOS=windows \
    GOARCH=amd64 \
    CGO_ENABLED=1 \
    CC=x86_64-w64-mingw32-gcc

COPY scripts /usr/bin/ 

RUN install-dependencies && \
    dpkg --add-architecture i386 && \
    apt-get update -qq && \
    apt-get install -y gnupg2 software-properties-common && \
    wget -qO - https://dl.winehq.org/wine-builds/winehq.key | apt-key add - && \
    apt-add-repository https://dl.winehq.org/wine-builds/debian/ && \
    wget -O- -q https://download.opensuse.org/repositories/Emulators:/Wine:/Debian/Debian_10/Release.key | apt-key add - && \
    echo "deb http://download.opensuse.org/repositories/Emulators:/Wine:/Debian/Debian_10 ./" | tee /etc/apt/sources.list.d/wine-obs.list && \
    apt-get update -qq && \
    apt-get install -y --install-recommends winehq-stable && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/*