package main

import (
	"fmt"
	"github.com/dainnilsson/forwagent/common"
	"io"
	"net"
	"os"
	"path/filepath"
)

type dialServer = func() (net.Conn, error)

func createListener(name string) (net.Listener, error) {
	filePath := filepath.Join(common.GetHomeDir(), ".gnupg", name)
	os.Remove(filePath)
	return net.Listen("unix", filePath)
}

func handleConnections(listener net.Listener, kind string, dial dialServer) {
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
		} else {
			go handleConnection(dial, conn, kind)
		}
	}

}

func handleConnection(dial dialServer, conn net.Conn, connType string) {
	defer conn.Close()

	serverConn, err := dial()
	if err != nil {
		fmt.Println("Error connecting to server:", err.Error())
		return
	}

	io.WriteString(serverConn, connType)

	common.ProxyConnections(conn, serverConn)
}

func runClient(dial dialServer) error {
	gpgSocket, err := createListener("S.gpg-agent")
	if err != nil {
		return err
	}
	sshSocket, err := createListener("S.gpg-agent.ssh")
	if err != nil {
		return err
	}

	go handleConnections(gpgSocket, "GPG", dial)
	handleConnections(sshSocket, "SSH", dial)

	return nil
}
