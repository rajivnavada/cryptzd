package main

import (
	"bytes"
	"gibberz/crypto"
	"os"
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

	_, err = key.EncryptMessage("Hello World!", "A message", user)
	if err != nil {
		panic(err)
	}

	//io.Copy(os.Stdout, strings.NewReader(output.Text()))

	kc := user.Keys()
	for k := kc.Next(); k != nil; k = kc.Next() {
		println("Printing messages for key with fingerprint = ", k.Fingerprint())
		println("---------------------------------------------------------------------------------------")
		println("")

		mc := k.Messages()
		for m := mc.Next(); m != nil; m = mc.Next() {
			println(m.Text())
			println("")
		}
	}

}
