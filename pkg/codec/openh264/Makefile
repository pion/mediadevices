git_url := https://github.com/cisco/openh264.git
version := v2.1.1
src_root_dir := src
lib_dir := lib
include_dir := include/openh264
lib_prefix := libopenh264
src_dir := $(src_root_dir)/$(MEDIADEVICES_TARGET_PLATFORM)
output_path := $(lib_dir)/$(lib_prefix)-$(MEDIADEVICES_TARGET_PLATFORM).a

# OS and Arch mapping to OpenH264 parameters
ifeq (windows,$(MEDIADEVICES_TARGET_OS))
	os := mingw_nt
else
	os := $(MEDIADEVICES_TARGET_OS)
endif

ifneq (,$(findstring $(MEDIADEVICES_TARGET_ARCH),armv6 armv7 armv8))
	arch := arm
else ifeq (x64,$(MEDIADEVICES_TARGET_ARCH))
	arch := x86_64
else
	arch := $(MEDIADEVICES_TARGET_ARCH)
endif

.PHONY: all
all: guard-MEDIADEVICES_TARGET_PLATFORM guard-MEDIADEVICES_TARGET_PLATFORM \
	guard-MEDIADEVICES_TARGET_OS guard-MEDIADEVICES_TARGET_ARCH \
	$(output_path) headers

headers: | $(src_dir) $(include_dir)
	@cp $(src_dir)/codec/api/svc/*.h $(include_dir)

$(output_path): $(src_dir)/$(lib_prefix).a | $(lib_dir)
	@cp $< $@

$(src_dir)/$(lib_prefix).a: | $(src_dir)
	$(MEDIADEVICES_TOOLCHAIN_BIN) make --directory=$(src_dir) $(lib_prefix).a \
		OS=$(os) ARCH=$(arch)

$(src_dir): | $(src_root_dir)
	git clone --depth=1 --branch=$(version) $(git_url) $@

$(src_root_dir) $(lib_dir) $(include_dir):
	@mkdir -p $@

guard-%:
	@if [ -z ${$*} ]; then echo "$* is a required environment variable"; exit 1; fi
