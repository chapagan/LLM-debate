.PHONY: setup test build run package clean

setup:
	./scripts/setup.sh

test:
	./scripts/test.sh

build:
	cd frontend && npm run build
	mkdir -p bin
	go build -o bin/llm-debate ./cmd/server

run:
	./scripts/run.sh

package:
	./scripts/package.sh

clean:
	rm -rf bin release frontend/dist
