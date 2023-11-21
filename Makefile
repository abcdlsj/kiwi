.DEFAULT_GOAL := help

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

.PHONY: build
build: # Build the project.
	for dir in $(shell ls cmd); do \
		go build -o bin/$$dir cmd/$$dir/main.go; \
	done

.PHONY: run
run: # Run the project.
	go run server/cmd/main.go