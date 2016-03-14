CC = clang
BIN = cryptz
HOST = 127.0.0.1
PORT = 8000
MONGO_HOST = 127.0.0.1
MONGO_DB_NAME = cryptz

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
	exec ./$(BIN) -host $(HOST) -port $(PORT) -mongoHost $(MONGO_HOST) -mongoDbName $(MONGO_DB_NAME)

