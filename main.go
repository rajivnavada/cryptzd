package main

import (
	"bytes"
	"gibberz/crypto"
	"io"
	"os"
)

func main() {
	buf := new(bytes.Buffer)
	buf.ReadFrom(os.Stdin)

	if buf.Len() == 0 {
		println("No input. Got nothing to do. Exiting")
		return
	}

	key, _, err := crypto.ImportKeyAndUser(buf.String())
	if err != nil {
		panic(err)
	}

	output, err := key.Encrypt("Hello World!")
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, output)
}
