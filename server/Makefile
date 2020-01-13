.PHONY: generate build

generate:
	@echo "\nGenerating assets...\n"
	go generate ./...
	@echo "\nAssets generated\n"

build: generate
	@echo "\nBuilding...\n"
	go build
	@echo "\nBuild completed. Run the server by ./unshort.link\n"