FROM dockcross/linux-arm64

ARG IMAGE=lherman/cross-linux-arm64
ARG VERSION=latest
ENV DEFAULT_DOCKCROSS_IMAGE ${IMAGE}:${VERSION}

COPY init.sh /tmp/init.sh
RUN bash /tmp/init.sh
