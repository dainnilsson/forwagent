package common

import "crypto/sha256"
import "crypto/x509"
import "encoding/pem"
import "io/ioutil"

func ReadFingerprintFile(path string) (fp [32]byte, err error) {
	certPem, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	blk, _ := pem.Decode(certPem)
	return ReadFingerprint(blk.Bytes)
}

func ReadFingerprint(derBytes []byte) (fp [32]byte, err error) {
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return
	}
	key, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return
	}
	fp = sha256.Sum256(key)
	return
}
