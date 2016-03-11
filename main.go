package main

import (
	"bytes"
	"os"
	"zecure/crypto"
)

func main() {
	buf := new(bytes.Buffer)
	buf.ReadFrom(os.Stdin)

	if buf.Len() == 0 {
		println("No input. Got nothing to do. Exiting")
		return
	}

	_, sender, err := crypto.ImportKeyAndUser(buf.String())
	if err != nil {
		panic(err)
	}

	user, err := crypto.FindUserWithEmail("rajivn@zillow.com")
	if err != nil {
		panic(err)
	}

	if err = user.EncryptMessage("Hello World!\n", "A message", sender.Id()); err != nil {
		panic(err)
	}

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
