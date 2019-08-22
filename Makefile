BINNAME=gprel

.PHONY: all clean

all: dep test build

build:
	go build -ldflags='-w -s' -o $(BINNAME) .

test:
	go test -v ./...

dep:
	dep ensure

clean:
	go clean
	rm -f $(BINNAME)
