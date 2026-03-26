BINARY   := simple-search
BUILD_FLAGS := --tags fts5
DB_FILE  := example.db

.PHONY: build run test clean fmt vet lint

build:
	go build $(BUILD_FLAGS) -o $(BINARY) .

run: build
	./$(BINARY)

test:
	go test $(BUILD_FLAGS) -v -count=1 ./...

clean:
	rm -f $(BINARY) $(DB_FILE)

fmt:
	gofmt -s -w .

vet:
	go vet $(BUILD_FLAGS) ./...

lint: fmt vet

# Reset DB and run
dev: clean run

# Import data from a file: make import FILE=data/extra.json
import: build
	./$(BINARY) -import $(FILE)
