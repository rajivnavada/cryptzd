BIN = cryptzd
HOST = 127.0.0.1
PORT = 8000
SCHEMAFILE = ./schema.sql
RUNPATH = /usr/local/var/db/cryptz
SQLFILE = $(RUNPATH)/cryptz.db
KEYFILE = $(RUNPATH)/key.pem
CERTFILE = $(RUNPATH)/cert.pem

format:
	go fmt ./...

test:
	go test ./...

install:
	go install .

cert.pem: key.pem

key.pem:
	openssl req \
		-x509 \
		-new \
		-newkey rsa:4096 \
		-subj "/C=US/ST=Denial/L=California/O=Dis/CN=cryptz.local" \
		-keyout /usr/local/var/db/cryptz/key.pem \
		-out /usr/local/var/db/cryptz/cert.pem \
		-days 3650 \
		-nodes

clean:
	rm -f $(BIN)
	rm -f $(GOBIN)/$(BIN) $(SQLFILE) $(KEYFILE) $(CERTFILE)
	find $(GOPATH)/pkg -maxdepth 2 -type d -name "cryptzd" -exec rm -rf {} \;

cleandb:
	rm $(SQLFILE)

seed:
	mkdir -p `dirname $(SQLFILE)`
	sqlite3 -init $(SCHEMAFILE) $(SQLFILE) -version

web: clean seed cert.pem install
	exec $(BIN) -host $(HOST) -port $(PORT) -db $(SQLFILE) -key $(KEYFILE) -cert $(CERTFILE)

