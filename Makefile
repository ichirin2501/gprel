BINNAME=gprel

.PHONY: all clean

all: test build

build:
	go build -ldflags='-w -s' -o $(BINNAME) ./cmd/main.go

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINNAME)
