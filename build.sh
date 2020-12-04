#!/bin/bash

MEDIADEVICES_TOOLCHAIN_OWNER=lherman
MEDIADEVICES_TOOLCHAIN_PREFIX=cross
MEDIADEVICES_SCRIPT_PATH=$(realpath ./scripts)
MEDIADEVICES_TOOLCHAIN_PATH=${MEDIADEVICES_SCRIPT_PATH}/${MEDIADEVICES_TOOLCHAIN_PREFIX}
MEDIADEVICES_DOCKERFILES_PATH=dockerfiles

# Reference: https://github.com/dockcross/dockcross#cross-compilers
MEDIADEVICES_TARGET_PLATFORMS=(
  linux-armv7
  linux-arm64
  linux-x64
  windows-x64
  darwin-x64
)

if [[ -z ${VERBOSE} ]]; then
  MEDIADEVICES_OUTPUT=/dev/null
else
  MEDIADEVICES_OUTPUT=/dev/stdout
fi

install_toolchains() {
  bash ${MEDIADEVICES_DOCKERFILES_PATH}/build.sh &> ${MEDIADEVICES_OUTPUT}
  for platform in ${MEDIADEVICES_TARGET_PLATFORMS[@]}
  do
    mkdir -p ${MEDIADEVICES_TOOLCHAIN_PATH}
    image=${MEDIADEVICES_TOOLCHAIN_OWNER}/${MEDIADEVICES_TOOLCHAIN_PREFIX}-${platform}
    bin_path=${MEDIADEVICES_TOOLCHAIN_PATH}/${MEDIADEVICES_TOOLCHAIN_PREFIX}-${platform}
    docker run ${image} > ${bin_path}
    chmod +x ${bin_path}
  done
}

build() {
  sub_builds=$(find pkg -type f -name "build.sh")
  for sub_build in ${sub_builds[@]}
  do
    sub_build=$(realpath ${sub_build})
    sub_build_dir=$(dirname ${sub_build})
    current_dir=${PWD}
    cd $sub_build_dir
    for platform in ${MEDIADEVICES_TARGET_PLATFORMS[@]}
    do
      export MEDIADEVICES_TOOLCHAIN_BIN=${MEDIADEVICES_TOOLCHAIN_PATH}/${MEDIADEVICES_TOOLCHAIN_PREFIX}-${platform}
      # convert '-' to '_' since '_' is more common in library names
      export MEDIADEVICES_TARGET_PLATFORM=${platform//-/_}
      export MEDIADEVICES_TARGET_OS=$(echo $MEDIADEVICES_TARGET_PLATFORM | cut -d'_' -f1)
      export MEDIADEVICES_TARGET_ARCH=${platform//${MEDIADEVICES_TARGET_OS}-/}

      echo "Building ${sub_build_dir}:"
      echo "  PLATFORM      : ${MEDIADEVICES_TARGET_PLATFORM}"
      echo "  OS            : ${MEDIADEVICES_TARGET_OS}"
      echo "  ARCH          : ${MEDIADEVICES_TARGET_ARCH}"
      echo "  TOOLCHAIN_BIN : ${MEDIADEVICES_TOOLCHAIN_BIN}"
      echo ""

      ${sub_build} &> ${MEDIADEVICES_OUTPUT}
    done
    cd ${current_dir}
  done
}

if [[ $# > 0 && $1 != "all" ]]; then
  MEDIADEVICES_TARGET_PLATFORMS=($1)
fi

install_toolchains
build
