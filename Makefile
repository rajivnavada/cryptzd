CC = clang
BIN = cryptz
HOST = 127.0.0.1
PORT = 8000
SCHEMAFILE = ./schema.sql
SQLFILE = /usr/local/var/db/cryptz/cryptz.db

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

cleandb:
	rm $(SQLFILE)

seed:
	sqlite3 -init $(SCHEMAFILE) $(SQLFILE) -version

web: clean all cert.pem
	exec ./$(BIN) -host $(HOST) -port $(PORT) -db $(SQLFILE)

