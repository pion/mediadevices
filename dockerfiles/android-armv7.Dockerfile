FROM dockcross/android-arm

ARG IMAGE=lherman/cross-android-armv7
ARG VERSION=latest
ENV DEFAULT_DOCKCROSS_IMAGE ${IMAGE}:${VERSION}

ENV ANDROID_NDK_HOME /opt/android-ndk
ENV ANDROID_NDK_VERSION r22

RUN apt-get update -qq && apt-get install -y -q --no-install-recommends \
  unzip 


  RUN mkdir /opt/android-ndk-tmp && \
  cd /opt/android-ndk-tmp && \
  wget -q https://dl.google.com/android/repository/android-ndk-${ANDROID_NDK_VERSION}-linux-x86_64.zip && \
# uncompress
  unzip -q android-ndk-${ANDROID_NDK_VERSION}-linux-x86_64.zip && \
# move to its final location
  mv ./android-ndk-${ANDROID_NDK_VERSION} ${ANDROID_NDK_HOME} && \
# remove temp dir
  cd ${ANDROID_NDK_HOME} && \
  rm -rf /opt/android-ndk-tmp

# add to PATH
ENV PATH ${PATH}:${ANDROID_NDK_HOME}

COPY init.sh /tmp/init.sh
RUN bash /tmp/init.sh
