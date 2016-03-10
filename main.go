package main

import (
	"bytes"
	"gibberz/crypto"
	"io"
	"os"
	"strings"
)

func main() {
	buf := new(bytes.Buffer)
	buf.ReadFrom(os.Stdin)

	if buf.Len() == 0 {
		println("No input. Got nothing to do. Exiting")
		return
	}

	key, user, err := crypto.ImportKeyAndUser(buf.String())
	if err != nil {
		panic(err)
	}

	output, err := key.EncryptMessage("Hello World!", "A message", user)
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, strings.NewReader(output.Text()))
}
