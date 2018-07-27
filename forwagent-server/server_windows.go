package main

import (
	"bytes"
	"fmt"
	"github.com/dainnilsson/forwagent/common"
	"github.com/davidmz/go-pageant"
	"golang.org/x/crypto/ssh/agent"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func runServer(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
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

func dialAssuan(path string) (conn net.Conn, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = startGpgAgent(); err != nil {
			return nil, err
		}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	i := bytes.Index(data, []byte("\n"))

	port, err := strconv.Atoi(string(data[:i]))
	if err != nil {
		return
	}
	nonce := data[i+1:]

	conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return
	}

	_, err = conn.Write(nonce)
	return
}

func handleGPGRequest(conn net.Conn) {
	assuanPath := filepath.Join(os.Getenv("AppData"), "gnupg", "S.gpg-agent")
	assuanConn, err := dialAssuan(assuanPath)
	if err != nil {
		fmt.Println("Error connecting to assuan socket:", err.Error())
		return
	}

	common.ProxyConnections(conn, assuanConn)
}
