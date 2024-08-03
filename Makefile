CORE_IAM_OPENAPI_SPEC=../astro/apps/core/docs/iam/v1beta1/iam_v1beta1.yaml
CORE_PLATFORM_OPENAPI_SPEC=../astro/apps/core/docs/platform/v1beta1/platform_v1beta1.yaml

DESIRED_OAPI_CODEGEN_VERSION=v2.1.0

## Location to install dependencies to
ENVTEST_ASSETS_DIR=$(shell pwd)/bin
$(ENVTEST_ASSETS_DIR):
	mkdir -p $(ENVTEST_ASSETS_DIR)
OAPI_CODEGEN ?= $(ENVTEST_ASSETS_DIR)/oapi-codegen

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -run TestAcc $(TESTARGS) -timeout 180m

# Run unit tests
.PHONY: test
test:
	go vet ./...
	TF_ACC="" go test ./... -v $(TESTARGS)

.PHONY: fmt
fmt:
	gofmt -w ./
	goimports -w ./
	[ -z "$$CIRCLECI" ] || git diff --exit-code --color=always # In CI: exit if anything changed

.PHONY: validate-fmt
validate-fmt:
	@output=$$(gofmt -l ./); \
	if [ -n "$$output" ]; then \
		echo "$$output"; \
		echo "Please run 'make fmt' to format the code"; \
		exit 1; \
	fi

.PHONY: dep
dep:
	git config core.hooksPath .githooks
	go mod download
	go install golang.org/x/tools/cmd/goimports
	go mod tidy

.PHONY: build
build:
	go build -o ${ENVTEST_ASSETS_DIR}
	go generate ./...

.PHONY: api_client_gen
api_client_gen:
	@echo "Checking oapi-codegen installation..."
	@if ! command -v oapi-codegen >/dev/null 2>&1; then \
		echo "oapi-codegen not found. Installing..."; \
		go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@$(DESIRED_OAPI_CODEGEN_VERSION); \
	elif ! oapi-codegen --version | grep -q $(DESIRED_OAPI_CODEGEN_VERSION); then \
		echo "Updating oapi-codegen to desired version..."; \
		go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@$(DESIRED_OAPI_CODEGEN_VERSION); \
	else \
		echo "Correct version of oapi-codegen is already installed."; \
	fi
	@echo "Generating IAM API client..."
	oapi-codegen -include-tags=User,Invite,Team,ApiToken -generate=types,client -package=iam "$(CORE_IAM_OPENAPI_SPEC)" > ./internal/clients/iam/api.gen.go
	@echo "Generating Platform API client..."
	oapi-codegen -include-tags=Organization,Workspace,Cluster,Options,Deployment,Role -generate=types,client -package=platform "$(CORE_PLATFORM_OPENAPI_SPEC)" > ./internal/clients/platform/api.gen.go