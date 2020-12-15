FROM dockcross/windows-static-x64-posix

ARG IMAGE=lherman/cross-windows-x64
ARG VERSION=latest
ENV DEFAULT_DOCKCROSS_IMAGE ${IMAGE}:${VERSION}

COPY init.sh /tmp/init.sh
RUN bash /tmp/init.sh
RUN ln -s /usr/src/mxe/usr/bin/x86_64-w64-mingw32.static.posix-gcc /usr/bin/x86_64-w64-mingw32-gcc && \
    ln -s /usr/src/mxe/usr/bin/x86_64-w64-mingw32.static.posix-g++ /usr/bin/x86_64-w64-mingw32-g++ && \
    ln -s /usr/src/mxe/usr/bin/x86_64-w64-mingw32.static.posix-ar /usr/bin/x86_64-w64-mingw32-ar
