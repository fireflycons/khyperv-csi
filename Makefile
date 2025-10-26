VERSION = 1.0.0

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
    detected_OS := Windows
	MOCKERY = mockery.exe
	SWAGDIR = internal/windows
	SWAGGERFILES = $(SWAGDIR)/controller/routes.go internal/models/rest/*.go internal/models/get-vhd.go
	POWERSHELL = C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe
	SOURCE_FILES_RAW = $(shell $(POWERSHELL) -ExecutionPolicy Unrestricted -NoProfile -File zbuild/get-windowsdeps.ps1)
	SOURCE_FILES = $(shell echo | set /p="$(SOURCE_FILES_RAW)")
	LOGGING_FILES_RAW = $(shell $(POWERSHELL) -ExecutionPolicy Unrestricted  -NoProfile -File zbuild/get-loggingdeps.ps1)
	LOGGING_FILES = $(shell echo | set /p="$(LOGGING_FILES_RAW)")
	LINT_TARGETS = powershell
	MOCK_TARGETS = internal/windows/powershell/runner.go
else
    detected_OS := $(shell uname)  # same as "uname -s"
	MOCKERY = mockery
	LINT_TARGETS = 
	MOCK_TARGETS = internal/linux/driver/mounter.go internal/linux/kvp/metadata.go
endif

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

PS_FILES := $(shell $(POWERSHELL) -ExecutionPolicy Unrestricted -NoProfile -Command 'Get-ChildItem -Recurse -Filter *.ps* -Path powershell-modules | ForEach-Object { Write-Host -NoNewLine $$_.FullName ""}')

cmd/khypervprovider/psmodule/khyperv-csi.$(VERSION).nupkg: $(PS_FILES)
	@$(POWERSHELL) -ExecutionPolicy Unrestricted -NoProfile -File powershell-modules/build-module.ps1 -Version $(VERSION) -Target "$@"

.PHONY: powershell
powershell: cmd/khypervprovider/psmodule/khyperv-csi.$(VERSION).nupkg ## (Windows) Build khyperv-csi PowerShell Module

endif

# ------------ WINDOWS SERVICE -------------
ifeq ($(detected_OS),Windows)

khypervprovider.exe: swagger $(SOURCE_FILES) $(LOGGING_FILES)
	go build -ldflags "-s -w" ./cmd/khypervprovider

.PHONY: windows-service
windows-service: powershell khypervprovider.exe ## (Windows) Build REST service for Hyper-V

endif

##@ Testing

# ------------ TESTS: WINDOWS SERVICE -------------
ifeq ($(detected_OS),Windows)

.PHONY: install-module
install-module: powershell ## (Windows) Install the powershell module as current user (for tests)
	@$(POWERSHELL) -ExecutionPolicy Unrestricted -NoProfile -NonInteractive -File cmd\khypervprovider\psmodule\install-module.ps1 -Package cmd/khypervprovider/psmodule/khyperv-csi.$(VERSION).nupkg -CurrentUser

.PHONY: test-service
test-service: powershell install-module ## (Windows) Test Windows Service components
	@go test -timeout 1m -v ./...

endif

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

