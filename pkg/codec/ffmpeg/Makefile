version=n7.0
srcPath=tmp/$(version)/src
installPath=tmp/$(version)
CGO_CFLAGS := -I$(CURDIR)/$(installPath)/include/
CGO_LDFLAGS := -L$(CURDIR)/$(installPath)/lib/
PKG_CONFIG_PATH := $(CURDIR)/$(installPath)/lib/pkgconfig
configure := --enable-libx264 --enable-gpl

# Main test target - depends on FFmpeg being built
test: $(installPath)/lib/libavcodec.a
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go test -v .

# Separate target for building FFmpeg (used by CI when cache miss)
build-ffmpeg: $(installPath)/lib/libavcodec.a
	@echo "FFmpeg build completed"

# Clean incomplete builds before starting
clean-incomplete:
	@if [ -d "$(srcPath)" ] && [ ! -f "$(installPath)/lib/libavcodec.a" ]; then \
		echo "Cleaning incomplete build..."; \
		rm -rf $(srcPath); \
	fi

# FFmpeg build rule
$(installPath)/lib/libavcodec.a: clean-incomplete $(srcPath)/Makefile
	cd $(srcPath) && make -j4
	cd $(srcPath) && make install
	@echo "Installation completed, checking results..."
	@ls -la $(installPath)/lib/ || echo "lib directory not found"
	@ls -la $(installPath)/include/ || echo "include directory not found"

$(srcPath)/Makefile: $(srcPath)/.git
	cd $(srcPath) && ./configure --prefix=$(CURDIR)/$(installPath) $(configure)

$(srcPath)/.git:
	rm -rf $(srcPath)
	mkdir -p $(srcPath)
	cd $(srcPath) && git clone https://github.com/FFmpeg/FFmpeg .
	cd $(srcPath) && git checkout $(version)

.PHONY: test build-ffmpeg clean-incomplete
