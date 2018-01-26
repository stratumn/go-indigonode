SHELL=/bin/bash
NIX_OS_ARCHS?=darwin-amd64 linux-amd64
WIN_OS_ARCHS?=windows-amd64
VERSION=$(shell ./version.sh)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_PATH=$(shell git rev-parse --show-toplevel)
GITHUB_REPO=$(shell basename $(GIT_PATH))
GITHUB_USER=$(shell dirname `git remote get-url --all origin` | sed 's/.*[:/]//')
GIT_TAG=$(VERSION)
CMD=$(GITHUB_REPO)
DIST_DIR=dist
RELEASE_NAME=$(GIT_TAG)
RELEASE_NOTES_FILE=RELEASE_NOTES.md
TEXT_FILES=LICENSE RELEASE_NOTES.md CHANGE_LOG.md
DOCKER_USER=$(GITHUB_USER)
DOCKER_FILE=Dockerfile
COVERAGE_FILE=coverage.txt
COVERHTML_FILE=coverhtml.txt
CLEAN_PATHS=$(DIST_DIR) $(COVERAGE_FILE) $(COVERHTML_FILE)
TMP_DIR:=$(shell mktemp -d)
DEPGRAPH_FILE=doc/depgraph.svg

GO_CMD=go
GO_LINT_CMD=gometalinter
GO_CYCLO_CMD=gocyclo
KEYBASE_CMD=keybase
GITHUB_RELEASE_COMMAND=github-release
DOCKER_CMD=docker
GRAPHPKG_CMD=graphpkg

GITHUB_RELEASE_FLAGS=--user '$(GITHUB_USER)' --repo '$(GITHUB_REPO)' --tag '$(GIT_TAG)'
GITHUB_RELEASE_RELEASE_FLAGS=$(GITHUB_RELEASE_FLAGS) --name '$(RELEASE_NAME)' --description "$$(cat $(RELEASE_NOTES_FILE))"

GO_LIST=$(GO_CMD) list
GO_BUILD=$(GO_CMD) build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -ldflags '-X github.com/$(GITHUB_USER)/$(GITHUB_REPO)/release.Version=$(VERSION) -X github.com/$(GITHUB_USER)/$(GITHUB_REPO)/release.GitCommit=$(GIT_COMMIT)'
GO_TEST=$(GO_CMD) test
GO_BENCHMARK=$(GO_TEST) -bench .
GO_LINT=$(GO_LINT_CMD) -d --vendor --enable-gc --skip=test --deadline=2m --disable="vetshadow" --disable="maligned" --disable="ineffassign" --disable="gocyclo" --disable="gas"
GO_CYCLO=$(GO_CYCLO_CMD)
KEYBASE_SIGN=$(KEYBASE_CMD) pgp sign
GITHUB_RELEASE_RELEASE=$(GITHUB_RELEASE_COMMAND) release $(GITHUB_RELEASE_RELEASE_FLAGS)
GITHUB_RELEASE_UPLOAD=$(GITHUB_RELEASE_COMMAND) upload $(GITHUB_RELEASE_FLAGS)
GITHUB_RELEASE_EDIT=$(GITHUB_RELEASE_COMMAND) edit $(GITHUB_RELEASE_RELEASE_FLAGS)
DOCKER_BUILD=$(DOCKER_CMD) build
DOCKER_PUSH=$(DOCKER_CMD) push

PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor)
MOCK_FILES=$(shell find . -name 'mock*.go' | grep -v vendor)
TEST_PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor | grep -v './grpc/' | grep -v 'mock' | grep -v 'test')
COVERAGE_PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor | grep -v './grpc/' | grep -v 'mock' | grep -v 'test')
COVERAGE_SOURCES=$(shell find . -name '*.go' | grep -v './grpc/' | grep -v 'mock' | grep -v 'test')
LINT_PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor | grep -v './grpc' | grep -v 'mock' | grep -v 'test')
BUILD_SOURCES=$(shell find . -name '*.go' | grep -v './grpc/' | grep -v 'mock' | grep -v 'test' | grep -v '_test.go')
CYCLO_SOURCES=$(shell find . -name '*.go' | grep -v vendor | grep -v './grpc/' | grep -v 'mock' | grep -v 'test')
GRPC_PROTOS=$(shell find grpc -name '*.proto')
GRPC_GO=$(GRPC_PROTOS:.proto=.pb.go)

PROTOS=$(shell find pb -name '*.proto')
PROTOS_GO=$(PROTOS:.proto=.pb.go)

NIX_EXECS=$(foreach os-arch, $(NIX_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD))
WIN_EXECS=$(foreach os-arch, $(WIN_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD).exe)
EXECS=$(NIX_EXECS) $(WIN_EXECS)
DOCKER_EXEC=$(DIST_DIR)/linux-amd64/$(CMD)
SIGNATURES=$(foreach exec, $(EXECS), $(exec).sig)
NIX_ZIP_FILES=$(foreach os-arch, $(NIX_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD).zip)
WIN_ZIP_FILES=$(foreach os-arch, $(WIN_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD).zip)
ZIP_FILES=$(NIX_ZIP_FILES) $(WIN_ZIP_FILES)

TEST_LIST=$(foreach package, $(TEST_PACKAGES), test_$(package))
LINT_LIST=$(foreach package, $(LINT_PACKAGES), lint_$(package))
GITHUB_UPLOAD_LIST=$(foreach file, $(ZIP_FILES), github_upload_$(firstword $(subst ., ,$(file))))
CLEAN_LIST=$(foreach path, $(CLEAN_PATHS), clean_$(path))


# == .PHONY ===================================================================
.PHONY: gx dep gometalinter graphpck deps generate protobuf unit system test coverage lint benchmark cyclo build git_tag github_draft github_upload github_publish docker_image docker_push clean license_headers $(TEST_LIST) $(BENCHMARK_LIST) $(LINT_LIST) $(GITHUB_UPLOAD_LIST) $(CLEAN_LIST)

# == all ======================================================================
all: build

# == deps =====================================================================
gx:
	go get -u github.com/whyrusleeping/gx
	go get -u github.com/whyrusleeping/gx-go

dep:
	go get -u github.com/golang/dep/cmd/dep

gometalinter:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

graphpkg:
	go get -u github.com/davecheney/graphpkg

gofast:
	go get github.com/gogo/protobuf/protoc-gen-gofast

deps: gx dep gometalinter graphpkg gofast
	gx install
	dep ensure

# == generate =================================================================

generate:
	@for d in $(PACKAGES); do \
		echo $(GO_CMD) generate $$d; \
		$(GO_CMD) generate $$d 2> /dev/null; \
	done
	
	@for f in $(MOCK_FILES); do \
		sed -i'.bak' 's|github.com/stratumn/alice/vendor/||g' $$f; \
		rm $$f.bak; \
	done

# == protobuf =================================================================

protobuf: $(GRPC_GO) $(PROTOS_GO)

grpc/%.pb.go: grpc/%.proto
	protoc -I $(GOPATH)/src/ github.com/$(GITHUB_USER)/$(GITHUB_REPO)/$< --gofast_out=plugins=grpc:$(GOPATH)/src
	sed -i'.bak' 's|golang.org/x/net/context|context|g' $@
	rm $@.bak

pb/%.pb.go: pb/%.proto
	protoc -I $(GOPATH)/src/ github.com/$(GITHUB_USER)/$(GITHUB_REPO)/$< --gofast_out=$(GOPATH)/src

# == doc ======================================================================

doc: $(DEPGRAPH_FILE)

$(DEPGRAPH_FILE): $(BUILD_SOURCES)
	$(GRAPHPKG_CMD) -stdout -match github.com/$(GITHUB_USER)/$(GITHUB_REPO) github.com/$(GITHUB_USER)/$(GITHUB_REPO) > $@

# == release ==================================================================
release: test lint clean build git_tag github_draft github_upload github_publish docker_image docker_push

# == test =====================================================================
test: unit system

unit: $(TEST_LIST)

$(TEST_LIST): test_%:
	@$(GO_TEST) $*

system:
	@$(GO_TEST) github.com/$(GITHUB_USER)/$(GITHUB_REPO)/test/system

# == coverage =================================================================
coverage: $(COVERAGE_FILE)

$(COVERAGE_FILE): $(COVERAGE_SOURCES)
	@for d in $(COVERAGE_PACKAGES); do \
	    $(GO_TEST) -coverprofile=profile.out -covermode=atomic $$d || exit 1; \
	    if [ -f profile.out ]; then \
	        cat profile.out >> $(COVERAGE_FILE); \
	        rm profile.out; \
	    fi \
	done

coverhtml:
	echo 'mode: set' > $(COVERHTML_FILE)
	@for d in $(TEST_PACKAGES); do \
	    $(GO_TEST) -coverprofile=profile.out $$d || exit 1; \
	    if [ -f profile.out ]; then \
	        tail -n +2 profile.out >> $(COVERHTML_FILE); \
	        rm profile.out; \
	    fi \
	done
	$(GO_CMD) tool cover -html $(COVERHTML_FILE)

# == benchmark ================================================================
benchmark:
	@$(GO_BENCHMARK) github.com/$(GITHUB_USER)/$(GITHUB_REPO)/test/benchmark/...

# == lint =====================================================================
lint:
	@$(GO_LINT) ./...

# == cyclomatic complexity ====================================================
cyclo: $(CYCLO_SOURCES)
	@$(GO_CYCLO) -over 9 $(CYCLO_SOURCES)

# == build ====================================================================
build: $(EXECS)

BUILD_OS_ARCH=$(word 2, $(subst /, ,$@))
BUILD_OS=$(firstword $(subst -, ,$(BUILD_OS_ARCH)))
BUILD_ARCH=$(lastword $(subst -, ,$(BUILD_OS_ARCH)))

$(EXECS): $(BUILD_SOURCES)
	GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) $(GO_BUILD) -o $@

# == sign =====================================================================
sign: $(SIGNATURES)

%.sig: %
	$(KEYBASE_SIGN) -d -i $* -o $@

# == zip ======================================================================
zip: $(ZIP_FILES)

ZIP_TMP_OS_ARCH_DIR=$(TMP_DIR)/$(BUILD_OS_ARCH)
ZIP_TMP_CMD_DIR=$(ZIP_TMP_OS_ARCH_DIR)/$(CMD)

%.zip: %.exe %.exe.sig
	mkdir -p $(ZIP_TMP_CMD_DIR)
	cp $*.exe $(ZIP_TMP_CMD_DIR)
	cp $*.exe.sig $(ZIP_TMP_CMD_DIR)
	cp $(TEXT_FILES) $(ZIP_TMP_CMD_DIR)
	mv $(ZIP_TMP_CMD_DIR)/LICENSE $(ZIP_TMP_CMD_DIR)/LICENSE.txt
	cd $(ZIP_TMP_OS_ARCH_DIR) && zip -r $(CMD){.zip,} 1>/dev/null
	cp $(ZIP_TMP_CMD_DIR).zip $@

%.zip: % %.sig
	mkdir -p $(ZIP_TMP_CMD_DIR)
	cp $* $(ZIP_TMP_CMD_DIR)
	cp $*.sig $(ZIP_TMP_CMD_DIR)
	cp $(TEXT_FILES) $(ZIP_TMP_CMD_DIR)
	cd $(ZIP_TMP_OS_ARCH_DIR) && zip -r $(CMD){.zip,} 1>/dev/null
	cp $(ZIP_TMP_CMD_DIR).zip $@

# == git_tag ==================================================================
git_tag:
	git tag $(GIT_TAG)
	git push origin --tags

# == github_draft =============================================================
github_draft:
	@if [[ $prerelease != "false" ]]; then \
		echo $(GITHUB_RELEASE_RELEASE) --draft --pre-release; \
		$(GITHUB_RELEASE_RELEASE) --draft --pre-release; \
	else \
		echo $(GITHUB_RELEASE_RELEASE) --draft; \
		$(GITHUB_RELEASE_RELEASE) --draft; \
	fi

# == github_upload ============================================================
github_upload: $(GITHUB_UPLOAD_LIST)

$(GITHUB_UPLOAD_LIST): github_upload_%: %.zip
	$(GITHUB_RELEASE_UPLOAD) --file $*.zip --name $(CMD)-$(BUILD_OS_ARCH).zip

# == github_publish ===========================================================
github_publish:
	@if [[ "$(PRERELEASE)" != "false" ]]; then \
		echo $(GITHUB_RELEASE_EDIT) --pre-release; \
		$(GITHUB_RELEASE_EDIT) --pre-release; \
	else \
		echo $(GITHUB_RELEASE_EDIT); \
		$(GITHUB_RELEASE_EDIT); \
	fi

# == docker_image =============================================================
docker_image: $(DOCKER_EXEC)
	$(DOCKER_BUILD) -f $(DOCKER_FILE) -t $(DOCKER_USER)/$(CMD):$(VERSION) -t $(DOCKER_USER)/$(CMD):latest .

# == docker_push ==============================================================
docker_push:
	$(DOCKER_PUSH) $(DOCKER_USER)/$(CMD):$(VERSION)
	$(DOCKER_PUSH) $(DOCKER_USER)/$(CMD):latest

# == license_headers ==========================================================
license_headers:
	./update-license-headers.sh

# == clean ====================================================================
clean: $(CLEAN_LIST)

$(CLEAN_LIST): clean_%:
	rm -rf $*
