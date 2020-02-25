vendor_dir = cvendor
src_dir = $(vendor_dir)/src
lib_dir = $(vendor_dir)/lib
include_dir = $(vendor_dir)/include

make_args.x86_64-windows = \
	CC=x86_64-w64-mingw32-gcc \
	CXX=x86_64-w64-mingw32-g++ \
	ARCH=x86_64 \
	OS=mingw_nt
make_args.x86_64-darwin = \
	CC=o64-clang \
	CXX=o64-clang++ \
	AR=llvm-ar \
	ARCH=x86_64 \
	OS=darwin

.PHONY: vendor
vendor: \
	$(include_dir)/openh264 \
	cross-libraries

$(include_dir)/openh264: $(src_dir)/openh264
	mkdir -p $@
	cp $^/codec/api/svc/*.h $@

$(lib_dir)/openh264/libopenh264.x86_64-linux.a: $(src_dir)/openh264
	$(MAKE) -C $^ clean \
		&& $(MAKE) -C $^ libraries
	mkdir -p $(dir $@)
	cp $^/libopenh264.a $@

$(lib_dir)/openh264/libopenh264.x86_64-windows.a: $(src_dir)/openh264
	$(MAKE) -C $^ clean \
		&& $(MAKE) -C $^ $(make_args.x86_64-windows) libraries
	mkdir -p $(dir $@)
	cp $^/libopenh264.a $@

$(lib_dir)/openh264/libopenh264.x86_64-darwin.a: $(src_dir)/openh264
	$(MAKE) -C $^ clean \
		&& $(MAKE) -C $^ $(make_args.x86_64-darwin) libraries
	mkdir -p $(dir $@)
	cp $^/libopenh264.a $@

.PHONY: cross-libraries
cross-libraries:
	docker build -t mediadevices-libs-builder -f libs-builder.Dockerfile .
	docker run --rm \
		-v $(CURDIR):/go/src/github.com/pion/mediadevices \
		mediadevices-libs-builder make $(lib_dir)/openh264/libopenh264.x86_64-linux.a
	docker run --rm \
		-v $(CURDIR):/go/src/github.com/pion/mediadevices \
		mediadevices-libs-builder make $(lib_dir)/openh264/libopenh264.x86_64-windows.a
	docker run --rm \
		-v $(CURDIR):/go/src/github.com/pion/mediadevices \
		mediadevices-libs-builder make $(lib_dir)/openh264/libopenh264.x86_64-darwin.a
