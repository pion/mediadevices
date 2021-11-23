FROM dockercore/golang-cross as m1cross

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update -qq && apt-get install -y -q --no-install-recommends \
    cmake \
    git \
    libssl-dev \
    libxml2-dev \
    libz-dev \
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
