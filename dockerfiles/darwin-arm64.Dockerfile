FROM dockercore/golang-cross as m1cross

ARG DEBIAN_FRONTEND=noninteractive

# Fix broken buster repositories due to Debian 10 being archived
RUN printf '%s\n' \
  'deb http://archive.debian.org/debian buster main contrib non-free' \
  'deb http://archive.debian.org/debian-security buster/updates main contrib non-free' \
  > /etc/apt/sources.list

RUN apt-get -o Acquire::Check-Valid-Until=false update -qq && \
    apt-get install -y -q --no-install-recommends \
        cmake \
        git \
        python3 \
        libssl-dev \
        libxml2-dev \
        zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

ENV SDK_VERSION=11.3 \
    TARGET_DIR=/osxcross/target \
    UNATTENDED=1

WORKDIR /work
RUN git clone --depth=1 https://github.com/tpoechtrager/osxcross.git /work \
  && cd /work/tarballs \
  && wget -q https://github.com/phracker/MacOSX-SDKs/releases/download/${SDK_VERSION}/MacOSX${SDK_VERSION}.sdk.tar.xz

# Build cross compile toolchain for Apple silicon
RUN ./build.sh


FROM dockcross/base

ENV OSX_CROSS_PATH=/osxcross

COPY --from=m1cross "${OSX_CROSS_PATH}/." "${OSX_CROSS_PATH}/"
ENV PATH=${OSX_CROSS_PATH}/target/bin:$PATH

COPY init.sh /tmp/init.sh
RUN bash /tmp/init.sh

ENV CC=arm64-apple-darwin20.4-clang \
    CXX=arm64-apple-darwin20.4-clang++ \
    CPP=arm64-apple-darwin20.4-clang++ \
    AR=arm64-apple-darwin20.4-ar \
    AS=arm64-apple-darwin20.4-as \
    LD=arm64-apple-darwin20.4-ld

COPY darwin-arm64.cmake ${OSX_CROSS_PATH}/
ENV CMAKE_TOOLCHAIN_FILE ${OSX_CROSS_PATH}/darwin-arm64.cmake

ARG IMAGE=lherman/cross-darwin-arm64
ARG VERSION=latest
ENV DEFAULT_DOCKCROSS_IMAGE ${IMAGE}:${VERSION}
