CORE_IAM_OPENAPI_SPEC=../astro-monorepo/apps/core/docs/iam/v1beta1/iam_v1beta1.yaml
CORE_PLATFORM_OPENAPI_SPEC=../astro-monorepo/apps/core/docs/platform/v1beta1/platform_v1beta1.yaml

DESIRED_OAPI_CODEGEN_VERSION=v2.1.0
DESIRED_MOCKERY_VERSION=v2.40.2

## Location to install dependencies to
ENVTEST_ASSETS_DIR=$(shell pwd)/bin
$(ENVTEST_ASSETS_DIR):
	mkdir -p $(ENVTEST_ASSETS_DIR)
MOCKERY ?= $(ENVTEST_ASSETS_DIR)/mockery
OAPI_CODEGEN ?= $(ENVTEST_ASSETS_DIR)/oapi-codegen

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run unit tests
.PHONY: test
test:
	go vet ./...
	go run github.com/onsi/ginkgo/v2/ginkgo run -r -v --skip-package internal/mocks --cover --covermode atomic --junit-report=report.xml --output-dir=test_results $(ARGS)

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
	go mod download
	go install golang.org/x/tools/cmd/goimports
	go mod tidy

.PHONY: build
build:
	go build -o ${ENVTEST_ASSETS_DIR}/terraform-provider-astronomer
	go generate ./...
#
#.PHONY: mock
#mock: $(ENVTEST_ASSETS_DIR)
#	# Install correct mockery version if not installed
#	(test -s $(MOCKERY) && $(MOCKERY) --version | grep -i $(DESIRED_MOCKERY_VERSION)) || GOBIN=$(ENVTEST_ASSETS_DIR) go install github.com/vektra/mockery/v2@$(DESIRED_MOCKERY_VERSION)
#	rm -rf internal/mocks
#	$(MOCKERY) --config .mockery.yaml

.PHONY: api_client_gen
api_client_gen: $(ENVTEST_ASSETS_DIR)
	# Install correct oapi-codegen version if not installed
	@{ $(OAPI_CODEGEN) --version | grep $(DESIRED_OAPI_CODEGEN_VERSION) > /dev/null; } || go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@$(DESIRED_OAPI_CODEGEN_VERSION)
	$(OAPI_CODEGEN) -include-tags=User,Invite,Team,ApiToken -generate=types,client -package=iam "${CORE_IAM_OPENAPI_SPEC}" > ./internal/clients/iam/api.gen.go
	$(OAPI_CODEGEN) -include-tags=Organization,Workspace,Cluster,Options,Deployment,Role -generate=types,client -package=platform "${CORE_PLATFORM_OPENAPI_SPEC}" > ./internal/clients/platform/api.gen.go
