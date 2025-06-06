TARGET_NAMESPACE=pipelines-as-code
HUGO_VERSION=0.96.0
GOLANGCI_LINT=golangci-lint
GOFUMPT=gofumpt
TKN_BINARY_NAME := tkn
TKN_BINARY_URL := https://tekton.dev/docs/cli/\#installation
LDFLAGS=
OUTPUT_DIR=bin
GO           = go
TIMEOUT_UNIT = 20m
TIMEOUT_E2E  = 45m
DEFAULT_GO_TEST_FLAGS := -race -failfast
GO_TEST_FLAGS :=

SHELL := bash
TOPDIR := $(shell git rev-parse --show-toplevel)
TMPDIR := $(TOPDIR)/tmp
HUGO_BIN := $(TMPDIR)/hugo/hugo
PY_FILES := $(shell find . -type f -regex ".*\.py" -print)
SH_FILES := $(shell find hack/ -type f -regex ".*\.sh" -print)
YAML_FILES := $(shell find . -not -regex '^./vendor/.*' -type f -regex ".*y[a]ml" -print)
MD_FILES := $(shell find . -type f -regex ".*md"  -not -regex '^./vendor/.*'  -not -regex '^./.vale/.*'  -not -regex "^./docs/themes/.*" -not -regex "^./.git/.*" -print)


ifeq ($(PAC_VERSION),)
	PAC_VERSION="$(shell git describe --tags --exact-match 2>/dev/null || echo nightly-`date +'%Y%m%d'`-`git rev-parse --short HEAD`)"
endif
FLAGS += -ldflags "-X github.com/openshift-pipelines/pipelines-as-code/pkg/params/version.Version=$(PAC_VERSION) $(LDFLAGS) -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/settings.TknBinaryName=$(TKN_BINARY_NAME) -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/settings.TknBinaryURL=$(TKN_BINARY_URL)"


##@ General
all: allbinaries test lint ## compile all binaries, test and lint
check: lint test ## run lint and test

.PHONY: help
help: ## print this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

FORCE:
.PHONY: vendor
vendor: ## generate vendor directory
	@echo Generating vendor directory
	@go mod tidy -compat=1.17 && go mod vendor

##@ Build
allbinaries: $(OUTPUT_DIR)/pipelines-as-code-controller $(OUTPUT_DIR)/pipelines-as-code-watcher $(OUTPUT_DIR)/tkn-pac ## compile all binaries

$(OUTPUT_DIR)/%: cmd/% FORCE ## compile binaries
	go build -mod=vendor $(FLAGS)  -v -o $@ ./$<

windows: ## compile windows binaries
	env GOOS=windows GOARCH=amd64 go build -mod=vendor $(FLAGS)  -v -o ./bin/tkn-pac.exe ./cmd/tkn-pac/main.go

##@ Testing
test: test-unit ## Run test-unit
test-clean:  ## Clean testcache
	@echo "Cleaning test cache"
	@go clean -testcache
.PHONY: test test-unit
test-no-cache: test-clean test-unit ## Run test-unit without caching
test-unit: ## Run unit tests
	@echo "Running unit tests..."
	$(GO) test $(DEFAULT_GO_TEST_FLAGS) $(GO_TEST_FLAGS) -timeout $(TIMEOUT_UNIT) ./pkg/...

.PHONY: test-e2e-cleanup
test-e2e-cleanup: ## cleanup test e2e namespace/pr left open
	@./hack/dev/e2e-tests-cleanup.sh

.PHONY: test-e2e
test-e2e:  test-e2e-cleanup ## run e2e tests
	env GODEBUG=asynctimerchan=1 \
		$(GO) test $(DEFAULT_GO_TEST_FLAGS) $(GO_TEST_FLAGS) -timeout $(TIMEOUT_E2E)  -failfast -count=1 -tags=e2e ./test

.PHONY: html-coverage
html-coverage: ## generate html coverage
	@mkdir -p tmp
	@go test -coverprofile=tmp/c.out ./.../ && go tool cover -html=tmp/c.out

##@ Linting
.PHONY: lint
lint: lint-go lint-yaml lint-md lint-python lint-shell ## run all linters

.PHONY: lint-go
lint-go: ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--timeout $(TIMEOUT_UNIT)

.PHONY: lint-yaml
lint-yaml: ${YAML_FILES} ## runs yamllint on all yaml files
	@echo "Linting yaml files..."
	@yamllint -c .yamllint $(YAML_FILES)


.PHONY: lint-md
lint-md: ## runs markdownlint and vale on all markdown files
	@echo "Linting markdown files..."
	@markdownlint $(MD_FILES)
	@echo "Grammar check with vale of documentation..."
	@vale docs/content --output=line --glob='*.md'
	@echo "CodeSpell on docs content"
	@codespell docs/content

.PHONY: lint-python
lint-python: ${PY_FILES} ## runs pylint on all python files
	@echo "Linting python files..."
	@ruff check $(PY_FILES)
	@ruff format --check $(PY_FILES)

.PHONY: lint-shell
lint-shell: ${SH_FILES} ## runs shellcheck on all python files
	@echo "Linting shell script files..."
	@shellcheck $(SH_FILES)

.PHONY: gitlint
gitlint: ## Run gitlint
	@gitlint --commit "`git log --format=format:%H --no-merges -1`" --ignore "Merge branch"

.PHONY: pre-commit
pre-commit: ## Run pre-commit hooks script manually
	@pre-commit run --all-files

##@ Linters Fixing
.PHONY: fix-linters
fix-linters: fix-golangci-lint fix-python-errors fix-markdownlint fix-trailing-spaces fumpt ## run all linters fixes

.PHONY: fix-markdownlint
fix-markdownlint: ## run markdownlint and fix on all markdown file
	@echo "Fixing markdown files..."
	@markdownlint --fix $(MD_FILES)

.PHONY: fix-trailing-spaces
fix-trailing-spaces: ## remove trailing spaces on all markdown and yaml file
	@sed --in-place 's/[[:space:]]\+$$//' $(MD_FILES) $(YAML_FILES)
	@[[ -n `git status --porcelain $(MD_FILES) $(YAML_FILES)` ]] && { echo "Markdowns and Yaml files has been cleaned 🧹. Cleaned Files: ";git status --porcelain $(MD_FILES) $(YAML_FILES) ;} || echo "Markdown and YAML are clean ✨"

.PHONE: fix-python-errors
fix-python-errors: ## fix all python errors generated by ruff
	@echo "Fixing python files..."
	@ruff check --fix $(PY_FILES)
	@ruff format --check $(PY_FILES)
	@[[ -n `git status --porcelain $(PY_FILES)` ]] && { echo "Python files has been cleaned 🧹. Cleaned Files: ";git status --porcelain $(PY_FILES) ;} || echo "Python files are clean ✨"

.PHONY: fix-golangci-lint
fix-golangci-lint: ## run golangci-lint and fix on all go files
	@echo "Fixing some golangi-lint files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--timeout $(TIMEOUT_UNIT) \
							--fix
	@[[ -n `git status --porcelain` ]] && { echo "Go files has been cleaned 🧹. Cleaned Files: ";git status --porcelain ;} || echo "Go files are clean ✨"

.PHONY: fmt 
fmt: ## formats the GO code(excludes vendors dir)
	@go fmt `go list ./... | grep -v /vendor/`

.PHONY: fumpt 
fumpt: ## formats the GO code with gofumpt(excludes vendors dir)
	@find test pkg -name '*.go'|xargs -P4 $(GOFUMPT) -w -extra

##@ Local Development
.PHONY: dev
dev: ## deploys dev setup locally
	./hack/dev/kind/install.sh

.PHONY: dev-redeploy
dev-redeploy: ## redeploy pac in local setup
	./hack/dev/kind/install.sh -p

##@ Generated files
check-generated: # check if all files that needs to be generated are generated
	@git status -uno |grep -E "modified:[ ]*(vendor/|.*.golden$)" && \
		{ echo "Vendor directory or Golden files has not been generated properly, commit the change first" ; \
		  git status -uno ;	exit 1 ;} || true

.PHONY: update-golden
update-golden: ## run unit tests (updating golden files)
	@echo "Running unit tests to update golden files..."
	@./hack/update-golden.sh

.PHONY: update-schemas
update-schemas: ## update openapi schemas
	@echo "Running update schemas..."
	@./hack/update-schemas.sh

.PHONY: generated
generated: update-golden fumpt ## generate all files that needs to be generated

##@ Docs

.PHONY: download-hugo
download-hugo: ## Download hugo software
	./hack/download-hugo.sh $(HUGO_VERSION) $(TMPDIR)/hugo

.PHONY: dev-docs
dev-docs: download-hugo ## preview live your docs with hugo
	@$(HUGO_BIN) server -s docs/ &
	@echo "Sleeping for 2s....";sleep 2 ; \
	if type -p xdg-open 2>/dev/null >/dev/null; then \
		xdg-open http://localhost:1313; \
	elif type -p open 2>/dev/null >/dev/null; then \
		open http://localhost:1313; \
	fi

##@ Misc

.PHONY: clean
clean: ## clean build artifacts
	rm -fR bin


