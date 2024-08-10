.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: run
run: ## Run a controller from your host.
	go run ./cmd/main.go

APP_NAME = clamd_exporter
VERSION = v0.1.0
# 定义目标平台和架构
PLATFORMS = darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64 windows-arm64

# 输出目录
OUTPUT_DIR = bin


# 默认目标
.PHONY: build
build: $(foreach platform,$(PLATFORMS),$(OUTPUT_DIR)/$(APP1)-$(VERSION).$(platform))

$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION).%: GOOS = $(word 1, $(subst -, ,$*))
$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION).%: GOARCH = $(word 2, $(subst -, ,$*))
$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION).%:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ cmd/main.go
	@if [ "$(GOOS)" = "windows" ]; then mv $@ $@.exe; fi

# 清理目标文件
clean:
	rm -rf $(OUTPUT_DIR)

CONTAINER_TOOL ?= docker
IMG ?= $(APP_NAME):latest

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}
