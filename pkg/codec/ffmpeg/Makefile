version=n7.0
srcPath=tmp/$(version)/src
CGO_CFLAGS := -I$(CURDIR)/tmp/$(version)/include/
CGO_LDFLAGS := -L$(CURDIR)/tmp/$(version)/lib/
PKG_CONFIG_PATH := $(CURDIR)/tmp/$(version)/lib/pkgconfig
configure := --enable-libx264 --enable-gpl

test: $(srcPath)
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go test -v .

$(srcPath):
	rm -rf $(srcPath)
	mkdir -p $(srcPath)
	cd $(srcPath) && git clone https://github.com/FFmpeg/FFmpeg .
	cd $(srcPath) && git checkout $(version)
	cd $(srcPath) && ./configure --prefix=.. $(configure)
	cd $(srcPath) && make -j4
	cd $(srcPath) && make install
