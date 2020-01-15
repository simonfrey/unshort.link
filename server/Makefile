DOCKER_USER?=test
IMAGE_NAME?=unshort

.PHONY: generate build dockerize

generate:
	@echo "Generating assets..."
	@go generate ./... && go generate ./db/db.go
	@echo "Assets generated"

build: generate test
	@echo "Building..."
	@go build -o unshort.link
	@echo "Build completed. Run the server by ./unshort.link"

test:
	@echo "Running tests...."
	@go test ./...
	@echo "Finished tests"
	@echo "Running vet..."
	@go vet ./...
	@echo "Finished vet"

dockerize:
	@echo "Start dockerizing...."
	docker image build -t $(IMAGE_NAME) .
	@echo "Finished dockerizing"