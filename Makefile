VERSION ?= 0.0.1

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
    detected_OS := Windows
	POWERSHELL = C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe  -ExecutionPolicy Unrestricted -NoProfile
	MODULE = $(shell $(POWERSHELL) -Command "(Get-Content .\go.mod | Select-String 'module') -split ' ' | Select-Object -Last 1")
	BUILD_DATE = $(shell $(POWERSHELL) -Command "(Get-Date).ToUniversalTime().ToString('ddd MMM dd HH:mm:ss UTC yyyy')")
	MOCKERY = mockery.exe
	SWAGGER = swagger
	SWAGDIR = internal/windows
	SWAGGERFILES = $(SWAGDIR)/controller/routes.go internal/models/rest/*.go internal/models/get-vhd.go
	SOURCE_FILES_RAW = $(shell $(POWERSHELL) -File zbuild/get-windowsdeps.ps1)
	SOURCE_FILES = $(shell echo | set /p="$(SOURCE_FILES_RAW)")
	LOGGING_FILES_RAW = $(shell $(POWERSHELL) -File zbuild/get-loggingdeps.ps1)
	LOGGING_FILES = $(shell echo | set /p="$(LOGGING_FILES_RAW)")
	LINT_TARGETS = powershell
	MOCK_TARGETS = internal/windows/powershell/runner.go
	TEST_TARGETS = powershell install-module
	BIN_TARGET = khypervprovider.exe
	MAIN_DIR = ./cmd/khypervprovider
	CGO =
	PSMODULE_TARGET = powershell
else
    detected_OS := $(shell uname)  # same as "uname -s"
	MODULE=$(shell grep -oP 'module\s+\K[\w\/\.]+' go.mod)
	BUILD_DATE = $(shell date -u)
	MOCKERY = mockery
	SWAGGER =
	SOURCE_FILES = $(shell find ./cmd/csi ./internal/common ./internal/constants ./internal/linux ./internal/models -type f -name '*.go' -print )
	LOGGING_FILES = $(shell find ./internal/logging/ -type d -name wineventlog -prune -o -type f -name '*.go' -not -name 'win*.go' -print)
	LINT_TARGETS =
	TEST_TARGETS =
	MOCK_TARGETS = internal/linux/driver/mounter.go internal/linux/kvp/metadata.go
	BIN_TARGET = hyperv-csi-plugin
	MAIN_DIR = ./cmd/csi
	CGO = CGO_ENABLED=0
	PSMODULE_TARGET =

endif
COMMIT_HASH = $(shell git rev-parse --short HEAD)

.PHONY: help



help: ## Display this help.
ifeq ($(detected_OS),Windows)
	@powershell -NoProfile -Command " \
	  Write-Host ''; \
	  Write-Host 'Usage:'; \
	  $$pattern = '^[a-zA-Z0-9_-]+:.*?##'; \
	  $$category = '^##@'; \
	  Get-Content $(MAKEFILE_LIST) | ForEach-Object { \
	    if ($$_ -match $$pattern) { \
	      $$parts = $$_ -split ':.*##'; \
	      Write-Host ('  {0,-15} {1}' -f $$parts[0], $$parts[1]) -ForegroundColor Cyan \
	    } elseif ($$_ -match $$category) { \
	      Write-Host ('{0}' -f ($$_ -replace '^##@', '')) -ForegroundColor White \
	    } \
	  } \
	"

else  # Linux

	@awk 'BEGIN {FS = ":.*##"; \
	    printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
	    /^[a-zA-Z_0-9-]+:.*?##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2} \
	    /^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)}' $(MAKEFILE_LIST)

endif

##@ Environment

bin/$(MOCKERY):
ifeq ($(detected_OS),Windows)
	@if not exist bin mkdir bin
	@curl -L -o mockery.tar.gz https://github.com/vektra/mockery/releases/download/v3.5.5/mockery_3.5.5_Windows_x86_64.tar.gz
	@7z x mockery.tar.gz -otmp
	@7z x tmp\mockery.tar -obin
	@del mockery.tar.gz
	@rmdir /s /q tmp
else
	@mkdir -p bin
	@curl -L -o mockery.tar.gz https://github.com/vektra/mockery/releases/download/v3.5.5/mockery_3.5.5_Linux_x86_64.tar.gz
	@tar -xzf mockery.tar.gz -C bin
	@rm mockery.tar.gz
endif

.PHONY: install-tools
install-tools: bin/$(MOCKERY)  ## Install build tools


##@ Build

# ------------ MOCKS -------------
temp/mocks.build: install-tools $(MOCK_TARGETS) go.mod
ifeq ($(detected_OS),Windows)
	@if not exist temp mkdir temp
else
	@mkdir -p temp
endif
	@bin/$(MOCKERY)
	@echo "mocks built" > temp/mocks.build

.PHONY: mocks ## Generate mocks
mocks: temp/mocks.build

# ------------ SWAGGER -------------
$(SWAGDIR)/swaggerui/docs.go: $(SWAGGERFILES)
	swag fmt -g (SWAGDIR)/controller/routes.go
	swag init -g $(SWAGDIR)/controller/routes.go --output $(SWAGDIR)/swaggerui

.PHONY: swagger
swagger: $(SWAGDIR)/swaggerui/docs.go ## Build swagger UI components.

# ------------ POWERSHELL -------------
ifeq ($(detected_OS),Windows)

PS_FILES := $(shell $(POWERSHELL) -Command 'Get-ChildItem -Recurse -Filter *.ps* -Path powershell-modules | ForEach-Object { Write-Host -NoNewLine $$_.FullName ""}')

cmd/khypervprovider/psmodule/khyperv-csi.$(VERSION).nupkg: $(PS_FILES)
	@$(POWERSHELL) -Command "Remove-Item -Force cmd\khypervprovider\psmodule\*.nupkg"
	@$(POWERSHELL) -File powershell-modules/build-module.ps1 -Version $(VERSION) -Target "$@"
	@go generate ./cmd/khypervprovider/psmodule

.PHONY: powershell
powershell: cmd/khypervprovider/psmodule/khyperv-csi.$(VERSION).nupkg ## (Windows) Build khyperv-csi PowerShell Module

endif

# ------------ EXECUTABLE -------------

$(BIN_TARGET): $(SWAGGER) $(SOURCE_FILES) $(LOGGING_FILES)
	$(CGO) go build -o $(BIN_TARGET) -ldflags "-s -w -X $(MODULE)/internal/common.Version=$(VERSION) -X $(MODULE)/internal/common.CommitHash=$(COMMIT_HASH) -X '$(MODULE)/internal/common.BuildDate=$(BUILD_DATE)'" $(MAIN_DIR)
ifeq ($(GITHUB_ACTIONS),true)
	echo "ARTIFACT=$(BIN_TARGET)" >> $$GITHUB_ENV
endif

.PHONY: executable
executable: $(PSMODULE_TARGET) $(BIN_TARGET) ## Build Executable (Hyper-V service on Windows, Driver on Linux)

##@ Docker (Linux only)
ifneq ($(detected_OS),Windows)

image: executable ## Make docker image
	docker build -t $(BIN_TARGET) --build-arg EXECUTABLE=$(BIN_TARGET) -f docker/Dockerfile .
endif


##@ Testing

# ------------------ TESTS ------------------
ifeq ($(detected_OS),Windows)

.PHONY: install-module
install-module: powershell ## (Windows) Install the powershell module as current user (for tests)
	@$(POWERSHELL) -NonInteractive -File cmd\khypervprovider\psmodule\install-module.ps1 -Package cmd/khypervprovider/psmodule/khyperv-csi.$(VERSION).nupkg -CurrentUser

endif

.PHONY: test
test: $(TEST_TARGETS) ## Test components approriate for OS this runs on
	@go test -timeout 1m -v ./...


temp/golangci-lint.ok: .golangci.yml $(SOURCE_FILES) $(LOGGING_FILES)
	golangci-lint run --timeout 2m30s ./...
ifeq ($(detected_OS),Windows)
	@cmd /c "if not exist temp mkdir temp"
	@cmd /c "type nul > temp\golangci-lint.ok"
else
	@mkdir -p temp
	@touch temp/golangci-lint.ok
endif

.PHONY: lint-windows
lint-windows: powershell temp/golangci-lint.ok

.PHONY: lint
lint: $(LINT_TARGETS) temp/golangci-lint.ok	## Run linter over source code

