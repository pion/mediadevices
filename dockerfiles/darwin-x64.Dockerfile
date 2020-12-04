FROM dockcross/base

ENV OSX_CROSS_PATH=/osxcross

COPY --from=dockercore/golang-cross "${OSX_CROSS_PATH}/." "${OSX_CROSS_PATH}/"
ENV PATH=${OSX_CROSS_PATH}/target/bin:$PATH

COPY init.sh /tmp/init.sh
RUN bash /tmp/init.sh

ENV CC=o64-clang \
    CXX=o64-clang++ \
    AR=llvm-ar

ARG IMAGE=lherman/cross-darwin-x64
ARG VERSION=latest
ENV DEFAULT_DOCKCROSS_IMAGE ${IMAGE}:${VERSION}
