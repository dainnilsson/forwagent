package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dainnilsson/forwagent/common"
	"github.com/davidmz/go-pageant"
	"github.com/flynn/noise"
	"github.com/go-noisesocket/noisesocket"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

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

	err = run(host, keys)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run(host string, keys noise.DHKey) error {
	l, err := noisesocket.Listen(host, &noisesocket.ConnectionConfig{
		StaticKey:      keys,
		VerifyCallback: verifyCallback,
	})
	if err != nil {
		return fmt.Errorf("Error listening: %s", err)
	}
	defer l.Close()
	fmt.Println("Listening on:", host)
	fmt.Println("Server key:", base64.StdEncoding.EncodeToString(keys.Public))

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting:", err.Error())
		} else {
			go handleRequest(conn)
		}
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 16)
	nRead, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading session type:", err.Error())
		return
	}

	r := string(buf[:nRead])

	fmt.Println("Session type:", r)
	if "SSH" == r {
		handleSSHRequest(conn)
	} else if "GPG" == r {
		handleGPGRequest(conn)
	} else {
		fmt.Println("Invalid session type:", r)
	}
}

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

func startGpgAgent() error {
	fmt.Println("Connect to gpg-agent...")
	cmd := exec.Command("gpg-connect-agent.exe", "/bye")
	return cmd.Run()
}

func handleSSHRequest(conn net.Conn) {
	if !pageant.Available() {
		if err := startGpgAgent(); err != nil {
			fmt.Println("Couldn't start gpg-agent:", err.Error())
			return
		}
		if !pageant.Available() {
			fmt.Println("Pageant not available!")
			return
		}
	}

	err := agent.ServeAgent(pageant.New(), conn)
	if err != nil && err.Error() != "EOF" {
		fmt.Println("Error serving SSH agent:", err.Error())
		return
	}
}

func readAssuanFile(path string) (port int, nonce []byte, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = startGpgAgent(); err != nil {
			return 0, nil, err
		}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	i := bytes.Index(data, []byte("\n"))

	port, err = strconv.Atoi(string(data[:i]))
	if err != nil {
		return
	}
	nonce = data[i+1:]

	return
}

func handleGPGRequest(conn net.Conn) {
	assuan := filepath.Join(os.Getenv("AppData"), "gnupg", "S.gpg-agent")
	port, nonce, err := readAssuanFile(assuan)
	if err != nil {
		fmt.Println("Error reading assuan file:", err.Error())
		return
	}

	assuanConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Println("Error connecting to assuan socket:", err.Error())
		return
	}

	_, err = assuanConn.Write(nonce)
	if err != nil {
		fmt.Println("Error writing nonce:", err.Error())
		return
	}

	// Forward between connections
	go func() {
		defer assuanConn.Close()
		_, err := io.Copy(assuanConn, conn)
		if err != nil {
			fmt.Println("Error forwarding server -> client:", err.Error())
		}
	}()

	_, err = io.Copy(conn, assuanConn)
	if err != nil && !strings.HasSuffix(err.Error(), "closed network connection") {
		fmt.Println("Error forwarding client -> server:", err.Error())
	}
}
