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

cert.pem: key.pem

key.pem:
	openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3650 -nodes

clean:
	rm -f $(BIN)

web: clean all cert.pem
	exec ./$(BIN)

