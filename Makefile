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
  darwin-x64 \
  darwin-arm64
cmd_build := build
cmd_test := test
examples_dir := examples
codec_dir := pkg/codec
codec_list := $(shell ls $(codec_dir)/*/Makefile)
codec_list := $(codec_list:$(codec_dir)/%/Makefile=%)
targets := $(foreach codec, $(codec_list), $(addprefix $(cmd_build)-$(codec)-, $(supported_platforms)))
pkgs_without_ext_device := $(shell go list ./... | grep -v mmal | grep -v vaapi)
pkgs_without_cgo := $(shell go list ./... | grep -v pkg/codec | grep -v pkg/driver | grep -v pkg/avfoundation)

define BUILD_TEMPLATE
ifneq (,$$(findstring $(2)-$(3),$$(supported_platforms)))
$$(cmd_build)-$(1)-$(2)-$(3): toolchain-$(2)-$(3)
	$$(MAKE) --directory=$$(codec_dir)/$(1) \
		MEDIADEVICES_TOOLCHAIN_BIN=$$(toolchain_path)/$(docker_prefix)-$(2)-$(3) \
		MEDIADEVICES_TARGET_PLATFORM=$(2)-$(3) \
		MEDIADEVICES_TARGET_OS=$(2) \
		MEDIADEVICES_TARGET_ARCH=$(3)
endif
endef

.PHONY: all
all: $(cmd_test) $(cmd_build)

# Subcommand:
# 	make build[-<codec_name>-<os>-<arch>]
#
# Description:
# 	Build codec dependencies to multiple platforms.
#
# Examples:
#   * make build: build all codecs for all supported platforms
#   * make build-opus-darwin-x64: only build opus for darwin-x64 platform
$(cmd_build): $(targets)

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

# Subcommand:
# 	make test
#
# Description:
# 	Run a series of tests
$(cmd_test):
	go vet $(pkgs_without_ext_device)
	go build $(pkgs_without_ext_device)
	# go build without CGO
	CGO_ENABLED=0 go build $(pkgs_without_cgo)
	# go build with CGO
	CGO_ENABLED=1 go build $(pkgs_without_ext_device)
	$(MAKE) --directory=$(examples_dir)
	go test -v -race -coverprofile=coverage.txt -covermode=atomic $(pkgs_without_ext_device)
