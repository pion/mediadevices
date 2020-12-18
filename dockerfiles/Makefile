dockerfiles := $(wildcard *.Dockerfile)
supported_platforms := $(dockerfiles:.Dockerfile=)

.PHONY: all
all: $(supported_platforms)

%: %.Dockerfile guard-MEDIADEVICES_DOCKER_OWNER guard-MEDIADEVICES_DOCKER_PREFIX
	docker build -t "$(MEDIADEVICES_DOCKER_OWNER)/$(MEDIADEVICES_DOCKER_PREFIX)-$@" -f "$<" . 

guard-%:
	@if [ -z ${$*} ]; then echo "$* is a required environment variable"; exit 1; fi
