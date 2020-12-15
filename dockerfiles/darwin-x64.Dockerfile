FROM dockcross/base

ENV OSX_CROSS_PATH=/osxcross

COPY --from=dockercore/golang-cross "${OSX_CROSS_PATH}/." "${OSX_CROSS_PATH}/"
ENV PATH=${OSX_CROSS_PATH}/target/bin:$PATH

COPY init.sh /tmp/init.sh
RUN bash /tmp/init.sh

ENV CC=x86_64-apple-darwin14-clang \
    CXX=x86_64-apple-darwin14-clang++ \
    CPP=x86_64-apple-darwin14-clang++ \
    AR=x86_64-apple-darwin14-ar \
    AS=x86_64-apple-darwin14-as \
    LD=x86_64-apple-darwin14-ld

COPY darwin-x64.cmake ${OSX_CROSS_PATH}/
ENV CMAKE_TOOLCHAIN_FILE ${OSX_CROSS_PATH}/darwin-x64.cmake

ARG IMAGE=lherman/cross-darwin-x64
ARG VERSION=latest
ENV DEFAULT_DOCKCROSS_IMAGE ${IMAGE}:${VERSION}
