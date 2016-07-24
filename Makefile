BIN = cryptz
HOST = 127.0.0.1
PORT = 8000
SCHEMAFILE = ./schema.sql
SQLFILE = /usr/local/var/db/cryptz/cryptz.db

format:
	go fmt ./...

test:
	go test ./...

install:
	go install .

cert.pem: key.pem

key.pem:
	openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3650 -nodes

clean:
	rm -f $(BIN)
	rm -f $(GOPATH)/bin/$(BIN)
	find $(GOPATH)/pkg -maxdepth 2 -type d -name "cryptz" -exec rm -rf {} \;

cleandb:
	rm $(SQLFILE)

seed:
	sqlite3 -init $(SCHEMAFILE) $(SQLFILE) -version

web: clean all cert.pem
	exec ./$(BIN) -host $(HOST) -port $(PORT) -db $(SQLFILE)

