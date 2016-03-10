package main

import (
	"bytes"
	"gibberz/crypto"
	"os"
)

func main() {
	buf := new(bytes.Buffer)
	buf.ReadFrom(os.Stdin)

	if buf.Len() > 0 {
		s := buf.String()
		//fmt.Println("Will send")
		//fmt.Println(s)
		//fmt.Println("")
		crypto.ImportKeyAndUser(s)
	}
}
