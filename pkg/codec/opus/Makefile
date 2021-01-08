git_url := https://github.com/xiph/opus.git
version := v1.3.1
src_root_dir := src
lib_dir := lib
include_dir := include
lib_prefix := libopus
src_dir := $(src_root_dir)/$(MEDIADEVICES_TARGET_PLATFORM)
output_path := $(lib_dir)/$(lib_prefix)-$(MEDIADEVICES_TARGET_PLATFORM).a

.PHONY: all
all: guard-MEDIADEVICES_TARGET_PLATFORM guard-MEDIADEVICES_TOOLCHAIN_BIN $(output_path) headers

headers: | $(src_dir) $(include_dir)
	@cp $(src_dir)/include/*.h $(include_dir)

$(output_path): $(src_dir)/$(lib_prefix).a | $(lib_dir)
	@cp $< $@

$(src_dir)/$(lib_prefix).a: | $(src_dir)
	cd $(src_dir) && \
		$(MEDIADEVICES_TOOLCHAIN_BIN) cmake -DOPUS_STACK_PROTECTOR=OFF -DCMAKE_C_FLAGS="-fpic" && \
		$(MEDIADEVICES_TOOLCHAIN_BIN) make VERBOSE=1

$(src_dir): | $(src_root_dir)
	git clone --depth=1 --branch=$(version) $(git_url) $@

$(src_root_dir) $(lib_dir) $(include_dir):
	@mkdir -p $@

guard-%:
	@if [ -z ${$*} ]; then echo "$* is a required environment variable"; exit 1; fi
