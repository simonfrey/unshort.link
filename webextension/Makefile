DOCKER_USER?=test
IMAGE_NAME?=unshort

.PHONY: generate build dockerize

dev: clean
	@echo "Generating test directories..."
	@mkdir dev
	@echo "Generating Chrome directory..."
	@mkdir dev/chrome
	@cp manifest.chrome.json dev/chrome/manifest.json
	@cp -r options.html js/ icons dev/chrome
	@echo "Chrome generated"
	@echo "Generating Firefox directory..."
	@mkdir dev/firefox
	@cp manifest.firefox.json dev/firefox/manifest.json
	@cp -r options.html js/ icons dev/firefox
	@echo "Firefox generated"
	@echo "Directories generated"

store: dev
	@echo "Building..."
	@mkdir store
	@cd dev/firefox/ && zip ../../store/unshort_link.firefox.zip * -r
	@cd dev/chrome && zip ../../store/unshort_link.chrome.zip * -r
	@echo "Build completed. Get the zips from ./store directory"

clean:
	@echo "Started cleaning...."
	@rm -rf dev store
	@echo "Finished cleaning"
