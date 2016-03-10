CC = clang
BINS = gibberz

all: $(BINS)

format:
	go fmt ./...

test:
	go test ./...

gibberz:
	CC=$(CC) go build -o $@ main.go

install: $(BINS)
	go install

clean:
	rm -f $(BINS)

