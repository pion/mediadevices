examples := $(shell find * -maxdepth 0 -type d)
examples := $(filter-out internal,$(examples))

.PHONY: all $(examples)
all: $(examples)

$(examples):
	cd $@ && go build -mod=mod
