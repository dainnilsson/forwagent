package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/dainnilsson/forwagent/common"
	"github.com/davidmz/go-pageant"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

func loadClients() map[[32]byte]bool {
	fp, err := common.ReadFingerprintFile(common.GetFilePath("client.pem"))
	if err != nil {
		fmt.Println("Error reading cert fingerprint:", err.Error())
		return nil
	}

	return map[[32]byte]bool{
		fp: true,
	}
}

func authenticateFps(fingerprints map[[32]byte]bool) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) (err error) {
		fp, err := common.ReadFingerprint(rawCerts[0])
		if err != nil {
			return
		}
		if !fingerprints[fp] {
			err = errors.New("Unknown client key")
		}
		return
	}
}

func main() {
	cert, err := tls.LoadX509KeyPair(
		common.GetFilePath("server.pem"),
		common.GetFilePath("server.key"),
	)
	if err != nil {
		fmt.Println("Error loading cert:", err.Error())
		os.Exit(1)
	}
	config := tls.Config{
		Certificates:          []tls.Certificate{cert},
		ClientAuth:            tls.RequireAnyClientCert,
		InsecureSkipVerify:    true,
		VerifyPeerCertificate: authenticateFps(loadClients()),
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
	l, err := tls.Listen("tcp", host, &config)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on:", host)

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
		return
	}
}

func handleSSHRequest(conn net.Conn) {
	avail := pageant.Available()
	if !avail {
		fmt.Println("Pageant not available!")
		return
	}
	agent.ServeAgent(pageant.New(), conn)
}

func readAssuanFile(path string) (port int, nonce []byte, err error) {
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
		os.Exit(1)
	}

	assuanConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Println("Error connecting to assuan socket:", err.Error())
		os.Exit(1)
	}

	_, err = assuanConn.Write(nonce)
	if err != nil {
		fmt.Println("Error writing nonce:", err.Error())
		os.Exit(1)
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
	if err != nil {
		fmt.Println("Error forwarding client -> server:", err.Error())
	}
}
