CC = clang
BINS = zecure

all: $(BINS)

format:
	go fmt ./...

test:
	go test ./...

zecure:
	CC=$(CC) go build -o $@ main.go

install: $(BINS)
	go install

clean:
	rm -f $(BINS)

web: clean all
	exec ./zecure

