package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dainnilsson/forwagent/common"
	"github.com/flynn/noise"
	"github.com/go-noisesocket/noisesocket"
	"io"
	"net"
	"os"
	"path/filepath"
)

func main() {
	priv, pub, err := common.GetKeyPair("client")
	if err != nil {
		fmt.Println("Couldn't read or generate key pair!", err.Error())
		os.Exit(1)
	}

	clientKeys := noise.DHKey{
		Public:  pub,
		Private: priv,
	}

	var host string
	if len(os.Args) > 2 {
		fmt.Println("Invalid command line usage!")
		os.Exit(1)
	} else if len(os.Args) == 2 {
		host = os.Args[1]
	} else {
		host = "127.0.0.1:4711"
	}
	fmt.Println("Using server:", host)
	fmt.Println("Client key:", base64.StdEncoding.EncodeToString(pub))

	config := noisesocket.ConnectionConfig{
		StaticKey:      clientKeys,
		VerifyCallback: verifyCallback,
	}

	gpgPath := filepath.Join(common.GetHomeDir(), ".gnupg", "S.gpg-agent")
	os.Remove(gpgPath)
	gpgSock, err := net.Listen("unix", gpgPath)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	go func() {
		defer gpgSock.Close()
		for {
			conn, err := gpgSock.Accept()
			if err != nil {
				fmt.Println("Error accepting:", err.Error())
			} else {
				go handleConnection(host, config, conn, "GPG")
			}
		}
	}()

	sshPath := filepath.Join(common.GetHomeDir(), ".gnupg", "S.gpg-agent.ssh")
	os.Remove(sshPath)
	sshSock, err := net.Listen("unix", sshPath)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer sshSock.Close()
	for {
		conn, err := sshSock.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
		} else {
			go handleConnection(host, config, conn, "SSH")
		}
	}
}

func verifyCallback(publicKey []byte, data []byte) error {
	keys, err := common.ReadKeyList("servers")
	if err != nil {
		return err
	}
	for _, key := range keys {
		if bytes.Equal(key, publicKey) {
			return nil
		}
	}

	publicB64 := base64.StdEncoding.EncodeToString(publicKey)
	fmt.Println("Unknown server key:", publicB64)
	fmt.Println("To allow:")
	fmt.Println("\necho '" + publicB64 + "' >> ~/.forwagent/servers.allowed\n")
	return errors.New("Connection closed, unknown public key.")
}

func handleConnection(host string, config noisesocket.ConnectionConfig, client net.Conn, connType string) {
	defer client.Close()

	server, err := noisesocket.Dial(host, &config)
	if err != nil {
		fmt.Println("Error connecting to server:", err.Error())
		return
	}
	defer server.Close()

	io.WriteString(server, connType)

	go func() {
		_, err := io.Copy(server, client)
		if err != nil {
			fmt.Println("Error forwarding server -> client:", err.Error())
		}
	}()

	_, err = io.Copy(client, server)
	if err != nil {
		fmt.Println("Error forwarding client -> server:", err.Error())
	}
}
