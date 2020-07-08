BINNAME=gprel

.PHONY: dep
dep:
	go mod download
	go mod tidy

.PHONY: build
build:
	go build -ldflags='-w -s' -o $(BINNAME) ./cmd/$(BINNAME)/main.go

.PHONY: test
test:
	go test -race -v ./...

.PHONY: test-coverage
test-coverage:
	go test -race -v -coverprofile coverage.out -covermode atomic ./...

.PHONY: clean
clean:
	go clean
	go clean -testcache
	rm -f $(BINNAME)
