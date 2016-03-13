CC = clang
BIN = cryptz

all: $(BIN)

format:
	go fmt ./...

test:
	go test ./...

$(BIN):
	CC=$(CC) go build -o $@ main.go

install: $(BIN)
	go install

clean:
	rm -f $(BIN)

web: clean all
	exec ./$(BIN)

