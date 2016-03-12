package main

import (
	"bytes"
	"net/http"
	"os"
	"zecure/crypto"
	"zecure/web"
)

func OldMain() {
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
	for k, err := kc.Next(); k != nil && err == nil; k, err = kc.Next() {
		println("Printing messages for key with fingerprint = ", k.Fingerprint())
		println("---------------------------------------------------------------------------------------")
		println("")

		mc := k.Messages()
		for m, err := mc.Next(); m != nil && err == nil; m, err = mc.Next() {
			println(m.Text())
			println("")
		}
	}

}

func main() {
	port := os.Getenv("PORT")
	router := web.Router()
	addr := "127.0.0.1:" + port

	println("Will start http server at:", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		panic(err)
	}
}
