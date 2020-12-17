GIT_URL=https://github.com/xiph/opus.git
VERSION=v1.3.1
SRC_DIR=src
LIB_DIR=lib
INCLUDE_DIR=include
ROOT_DIR=${PWD}
LIB_PREFIX=libopus

mkdir -p ${LIB_DIR} ${INCLUDE_DIR}

git clone --depth=1 --branch=${VERSION} ${GIT_URL} ${SRC_DIR}
cd ${SRC_DIR}
${MEDIADEVICES_TOOLCHAIN_BIN} cmake -DOPUS_STACK_PROTECTOR=OFF .
${MEDIADEVICES_TOOLCHAIN_BIN} make -j2
mv ${LIB_PREFIX}.a ${ROOT_DIR}/${LIB_DIR}/${LIB_PREFIX}_${MEDIADEVICES_TARGET_PLATFORM}.a
mkdir -p ${ROOT_DIR}/${INCLUDE_DIR}
cp include/*.h ${ROOT_DIR}/${INCLUDE_DIR}
git clean -dfx
git reset --hard
