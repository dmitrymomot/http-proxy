# Environment variables
LATEST_COMMIT := $$(git rev-parse HEAD)

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build application and compile protobuf for golang
	@go clean
	@CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build \
	-a -installsuffix nocgo \
	-ldflags "-X main.buildTag=`date -u +%Y%m%d.%H%M%S`-$(LATEST_COMMIT)" \
	-o http-proxy .

.PHONY: docker
docker: ## Build docker image
	@docker build . -t http-proxy:latest

.PHONY: deploy
deploy: ## Deploy pods to kubernetes
	@kubectl apply -f k8s.yml

.PHONY: down
down: ## Down pods
	@kubectl delete -f k8s.yml

.PHONY: reload
reload: down deploy info ## Reload after app was rebuilt

.PHONY: info
info: ## Get cluster info
	@kubectl get all

.PHONY: logs
log: ## Show logs
	@kubectl logs -lapp=http-proxy --container=http-proxy
