package common

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/flynn/noise"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

func generateKeyPair(privateFName string, publicFName string) (keys noise.DHKey, err error) {
	keys, err = noise.DH25519.GenerateKeypair(rand.Reader)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(privateFName, keys.Private, 0600)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(publicFName, keys.Public, 0600)
	return
}

func GetHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Couldn't get the current user!")
		os.Exit(1)
	}
	return usr.HomeDir
}

func getConfigFile(name string) string {
	pth := filepath.Join(GetHomeDir(), ".forwagent")
	os.MkdirAll(pth, os.ModePerm)
	return filepath.Join(pth, name)
}

func ReadKeyList(name string) (keys [][]byte, err error) {
	keysFile := getConfigFile(name + ".allowed")
	file, err := os.Open(keysFile)
	defer file.Close()
	if err != nil {
		return keys, nil
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, err := base64.StdEncoding.DecodeString(scanner.Text())
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return
}

func GetKeyPair(name string) (keys noise.DHKey, err error) {
	privateFile := getConfigFile(name + ".priv")
	publicFile := getConfigFile(name + ".pub")

	private, err := ioutil.ReadFile(privateFile)
	if err != nil {
		fmt.Println("Error loading private key, generating...")
		return generateKeyPair(privateFile, publicFile)
	}

	public, err := ioutil.ReadFile(publicFile)

	keys = noise.DHKey{
		Public:  public,
		Private: private,
	}
	return
}
