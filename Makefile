# glide is required to install dependencies. If make target depends
# on it and it is not installed, make will be terminated.
GLIDE := $(shell which glide)
ifeq ($(GLIDE),)
    GLIDE = must-install-glide
endif

$(GLIDE):
	@echo "glide is not installed."
	@echo "See https://github.com/Masterminds/glide#install for installation instructions."
	@exit 1

.PHONY: all
all: dep build

.PHONY: dep
dep: glide.yaml glide.lock | $(GLIDE)
	glide install

.PHONY: build
build:
	go build -o ${CURDIR}/bin/dweller ./cmd/dweller

.PHONY: image
image:
	docker build -t fukt/dweller .

.PHONY: codegen
codegen:
	vendor/k8s.io/code-generator/generate-groups.sh client,informer,lister,deepcopy \
		github.com/fukt/dweller/pkg/client \
		github.com/fukt/dweller/pkg/apis \
		"dweller:v1alpha1"


