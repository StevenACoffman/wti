APP=git-wti
GOBIN?=${HOME}/go/bin
GOPRIVATE?=github.com/StevenACoffman
INSTALLPATH?=${GOBIN}/${APP}
.PHONY: clean
clean: ## - Cleans old binary
	@printf "\033[32m\xE2\x9c\x93 Cleaning your code\n\033[0m"
	mkdir -p ${GOBIN}
	rm -f ${INSTALLPATH} || true
.PHONY: build
build: clean ## - Build the application
	@printf "\033[32m\xE2\x9c\x93 Building your code\n\033[0m"
	PATH="${GOBIN}:${PATH}" GOPRIVATE=$(GOPRIVATE) go build -trimpath \
	-ldflags='-w -s -extldflags "-static"' -a \
	-o ${INSTALLPATH} ./cmd/wti/main.go
.PHONY: run
run: build ## - Runs your application after building
	@printf "\033[32m\xE2\x9c\x93 Running your code\n\033[0m"
	@PATH="${GOBIN}:${PATH}" GOPRIVATE=$(GOPRIVATE) go run cmd/wti/main.go

.PHONY: krun
krun: ## - Just Runs your already compiled application, for Kathy
	@printf "\033[32m\xE2\x9c\x93 Running your code\n\033[0m"
	@PATH="${GOPATH}/bin:${PATH}" git-wti

.PHONY: mego
mego: run ## - Runs your application after building but has a cute name if you say it out loud
	@echo Done!

.PHONY: help
## help: Prints this help message
help: ## - Show help message
	@printf "\033[32m\xE2\x9c\x93 usage: make [target]\n\n\033[0m"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
