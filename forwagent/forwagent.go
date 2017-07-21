package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"forwagent/common"
	"io"
	"net"
	"os"
	"path/filepath"
)

func main() {
	fingerprint, err := common.ReadFingerprintFile(common.GetFilePath("server.pem"))
	if err != nil {
		fmt.Println("Error loading server cert:", err.Error())
		os.Exit(1)
	}
	cert, err := tls.LoadX509KeyPair(
		common.GetFilePath("client.pem"),
		common.GetFilePath("client.key"),
	)
	if err != nil {
		fmt.Println("Error loading cert:", err.Error())
		os.Exit(1)
	}
	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	gpgPath := filepath.Join(common.GetHomePath(), ".gnupg", "S.gpg-agent")
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
				go handleConnection(config, fingerprint, conn, "GPG")
			}
		}
	}()

	sshPath := path + "S.gpg-agent.ssh"
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
			go handleConnection(config, fingerprint, conn, "SSH")
		}
	}
}

func handleConnection(config tls.Config, fingerprint [32]byte, client net.Conn, connType string) {
	defer client.Close()

	server, err := tls.Dial("tcp", "127.0.0.1:4711", &config)
	if err != nil {
		fmt.Println("Error connecting to server:", err.Error())
		return
	}
	state := server.ConnectionState()
	cert := state.PeerCertificates[0]
	pkDer, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		fmt.Println("Error serializing public key:", err.Error())
		return
	}
	fpCompare := sha256.Sum256(pkDer)
	if fingerprint != fpCompare {
		fmt.Println("Server has wrong public key:", fpCompare)
		return
	}

	io.WriteString(server, connType)

	go func() {
		defer server.Close()
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
