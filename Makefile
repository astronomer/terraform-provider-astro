ifneq (,$(wildcard .env))
    include .env
    export
endif

CORE_IAM_OPENAPI_SPEC=../astro/apps/core/docs/iam/v1beta1/iam_v1beta1.yaml
CORE_PLATFORM_OPENAPI_SPEC=../astro/apps/core/docs/platform/v1beta1/platform_v1beta1.yaml
CORE_LABS_OPENAPI_SPEC=../astro/apps/core/docs/labs/v1.1/labsapi_v1.1.yaml
CORE_PLATFORM_V1_OPENAPI_SPEC=../astro/apps/core/docs/versioned/v1.0/versionedapi_v1.0.yaml

DESIRED_OAPI_CODEGEN_VERSION=v2.1.0

## Location to install dependencies to
ENVTEST_ASSETS_DIR=$(shell pwd)/bin
$(ENVTEST_ASSETS_DIR):
	mkdir -p $(ENVTEST_ASSETS_DIR)
OAPI_CODEGEN ?= $(ENVTEST_ASSETS_DIR)/oapi-codegen

MOCKERY_VERSION := 2.42.0

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -run TestAcc $(TESTARGS) -timeout 180m

# Run unit tests
.PHONY: test
test:
	go vet ./...
	TF_ACC="" RUN_IMPORT_SCRIPT_TEST="" go test ./... -v $(TESTARGS)

# Run script tests
.PHONY: test-import-script
test-import-script:
	go test ./import/... -v $(TESTARGS)

.PHONY: fmt
fmt:
	gofmt -w ./
	goimports -w ./
	[ -z "$$CIRCLECI" ] || git diff --exit-code --color=always # In CI: exit if anything changed

.PHONY: mock
mock: ensure-mockery generate-mocks

.PHONY: ensure-mockery
ensure-mockery:
	@if ! mockery --version | grep -q $(MOCKERY_VERSION); then \
		echo "Installing Mockery $(MOCKERY_VERSION)"; \
		go install $(MOCKERY_PACKAGE); \
	fi

.PHONY: generate-mocks
generate-mocks:
	@echo "Generating mocks..."
	@rm -rf mocks
	@mockery --name=ClientWithResponsesInterface \
		--dir=./internal/clients/iam \
		--output=./internal/mocks/iam \
		--outpkg=mocks_iam
	@mockery --name=ClientWithResponsesInterface \
		--dir=./internal/clients/platform \
		--output=./internal/mocks/platform \
		--outpkg=mocks_platform
	@mockery --name=ClientWithResponsesInterface \
		--dir=./internal/clients/platform_v1 \
		--output=./internal/mocks/platform_v1 \
		--outpkg=mocks_platform_v1
	@echo "Mocks generated successfully."

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
	# All three clients below use the forked union template (-templates ./internal/clients/oapi-templates,
	# see union.tmpl) which injects discriminator values into marshaled JSON so Core's optional
	# discriminator fields compile. Re-apply that fork when bumping DESIRED_OAPI_CODEGEN_VERSION.
	@echo "Generating IAM API client..."
	# Do not add "AllowedIpAddressRange" here: it renames the unrelated ListUsersParamsSorts
	# constants (oapi-codegen v2.1.0 constant-naming collision with
	# ListAllowedIpAddressRangesParamsSorts's overlapping values), breaking
	# data_source_users_list.go. The endpoint is generated into platform_v1 instead, where
	# that collision does not occur; the hand-authored binding below stays only because
	# api.gen.go's ClientInterface still references it.
	oapi-codegen -templates ./internal/clients/oapi-templates -include-tags=User,Invite,Team,ApiToken,AgentToken,Role -generate=types,client -package=iam "$(CORE_IAM_OPENAPI_SPEC)" > ./internal/clients/iam/api.gen.go
	@echo "Generating Platform API client..."
	oapi-codegen -templates ./internal/clients/oapi-templates -include-tags=Organization,Workspace,Cluster,Options,Deployment,Role,Environment,Alerts,NotificationChannels -generate=types,client -package=platform "$(CORE_PLATFORM_OPENAPI_SPEC)" > ./internal/clients/platform/api.gen.go
	@echo "Generating Labs API client..."
	# Adding "NotificationChannels" prefixes the AlertNotificationChannelType constants
	# (EMAIL -> AlertNotificationChannelTypeEMAIL, etc.) because the bulk notification-channel
	# types reuse the same bare enum values. No provider code references the bare labs constants,
	# so the rename is safe; the bulk NC resource uses the NotificationChannelType enum.
	oapi-codegen -templates ./internal/clients/oapi-templates -include-tags=Alerts,AllowedIpAddressRange,NotificationChannels -generate=types,client -package=labs "$(CORE_LABS_OPENAPI_SPEC)" > ./internal/clients/labs/api.gen.go
	@echo "Generating Platform v1 (unified public API) client..."
	# Alerts and NotificationChannels are deliberately absent: those resources do not exist in
	# the v1 spec at all and stay on platform/v1beta1 and labs. Organization is absent because
	# v1 drops supportPlan, which the astro_organization data source exposes as support_plan.
	# Adding tags here reshuffles oapi-codegen's unprefixed enum constants: with Options
	# included, WorkerQueueRequestAstroMachine* lose their prefix and become bare A5/A10/...,
	# so schemas/deployment.go references the stable UpdateWorkerQueueRequestAstroMachine*
	# spellings instead. Prefer prefixed constants here; a future tag addition renames the
	# unprefixed ones and you want a compile error, not a silent change.
	oapi-codegen -templates ./internal/clients/oapi-templates -include-tags=Environment,Deployment,Cluster,Workspace,Options,User,Invite,Team,ApiToken,AgentToken,Role,AllowedIpAddressRange -generate=types,client -package=platform_v1 "$(CORE_PLATFORM_V1_OPENAPI_SPEC)" > ./internal/clients/platform_v1/api.gen.go
