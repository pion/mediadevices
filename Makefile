# When there's no argument provided to make, all codecs will be built to all 
# supported platforms. Otherwise, it'll build only the provided target.
#
# Usage:
#   make [<codec_name>-<os>-<arch>]
#
# Examples:
#   * make: build all codecs for all supported platforms
#   * make opus-darwin-x64: only build opus for darwin-x64 platform

docker_owner := lherman
docker_prefix := cross
toolchain_dockerfiles := dockerfiles
script_path := $(realpath scripts)
toolchain_path := $(script_path)/$(docker_prefix)
os_list := \
	linux \
	windows \
	darwin
arch_list := \
	armv7 \
	arm64 \
	x64
supported_platforms := \
  linux-armv7 \
  linux-arm64 \
  linux-x64 \
  windows-x64 \
  darwin-x64
codec_dir := pkg/codec
codec_list := $(shell ls $(codec_dir)/*/Makefile)
codec_list := $(codec_list:$(codec_dir)/%/Makefile=%)
targets := $(foreach codec, $(codec_list), $(addprefix $(codec)-, $(supported_platforms)))

define BUILD_TEMPLATE
ifneq (,$$(findstring $(2)-$(3),$$(supported_platforms)))
$(1)-$(2)-$(3): toolchain-$(2)-$(3)
	$$(MAKE) --directory=$$(codec_dir)/$(1) \
		MEDIADEVICES_TOOLCHAIN_BIN=$$(toolchain_path)/$(docker_prefix)-$(2)-$(3) \
		MEDIADEVICES_TARGET_PLATFORM=$(2)-$(3) \
		MEDIADEVICES_TARGET_OS=$(2) \
		MEDIADEVICES_TARGET_ARCH=$(3)
endif
endef

.PHONY: all
all: $(targets)

toolchain-%: $(toolchain_dockerfiles)
	$(MAKE) --directory=$< "$*" \
		MEDIADEVICES_DOCKER_OWNER=$(docker_owner) \
		MEDIADEVICES_DOCKER_PREFIX=$(docker_prefix)
	@mkdir -p $(toolchain_path)
	@docker run $(docker_owner)/$(docker_prefix)-$* > \
		$(toolchain_path)/$(docker_prefix)-$*
	@chmod +x $(toolchain_path)/$(docker_prefix)-$*

$(foreach codec, $(codec_list), \
	$(foreach os, $(os_list), \
		$(foreach arch, $(arch_list), \
			$(eval $(call BUILD_TEMPLATE,$(codec),$(os),$(arch))))))
