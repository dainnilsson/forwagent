package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dainnilsson/forwagent/common"
	"github.com/go-noisesocket/noisesocket"
	"os"
)

func verifyCallback(publicKey []byte, data []byte) error {
	if len(publicKey) == 0 {
		return nil
	}
	keys, err := common.ReadKeyList("clients")
	if err != nil {
		return err
	}
	for _, key := range keys {
		if bytes.Equal(key, publicKey) {
			return nil
		}
	}
	publicB64 := base64.StdEncoding.EncodeToString(publicKey)
	fmt.Println("Unknown client key:" + publicB64)
	fmt.Println("To allow:")
	fmt.Println("\necho '" + publicB64 + "' >> ~/.forwagent/clients.allowed\n")
	return errors.New("Connection refused, unauthorized public key: " + publicB64)
}

func main() {
	var host string
	if len(os.Args) > 2 {
		fmt.Println("Invalid command line usage!")
		os.Exit(1)
	} else if len(os.Args) == 2 {
		host = os.Args[1]
	} else {
		host = "127.0.0.1:4711"
	}

	keys, err := common.GetKeyPair("server")
	if err != nil {
		fmt.Println("Couldn't read or generate key pair:", err)
		os.Exit(1)
	}

	listener, err := noisesocket.Listen(host, &noisesocket.ConnectionConfig{
		StaticKey:      keys,
		VerifyCallback: verifyCallback,
	})
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on:", host)
	fmt.Println("Server key:", base64.StdEncoding.EncodeToString(keys.Public))

	err = runServer(listener)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
