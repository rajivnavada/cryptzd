package main

import (
	"bytes"
	"fmt"
	"gibberz/crypto"
	"os"
)

func main() {
	buf := new(bytes.Buffer)
	buf.ReadFrom(os.Stdin)

	if buf.Len() == 0 {
		fmt.Println("No input. Got nothing to do")
		return
	}

	s := buf.String()
	//fmt.Println("Will send")
	//fmt.Println(s)
	//fmt.Println("")
	key, _, err := crypto.ImportKeyAndUser(s)
	if err != nil {
		panic(err)
	}

	output, err := key.Encrypt("Hello World!")
	if err != nil {
		panic(err)
	}

	fmt.Println(output)
}
