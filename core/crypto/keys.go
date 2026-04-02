package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
)

// GenerateKey creates a new ECDSA private key.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// SignData hashes the data and signs it with the private key.
func SignData(privKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		return nil, err
	}

	// Encode r and s into a single byte slice (simplified for this build)
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature checks if the signature is valid for the given data and public key.
func VerifySignature(pubKey *ecdsa.PublicKey, data, signature []byte) bool {
	hash := sha256.Sum256(data)

	// Split the signature back into r and s
	sigLen := len(signature)
	if sigLen%2 != 0 {
		return false
	}

	r := new(big.Int).SetBytes(signature[:sigLen/2])
	s := new(big.Int).SetBytes(signature[sigLen/2:])

	return ecdsa.Verify(pubKey, hash[:], r, s)
}

// HexToPrivateKey reconstructs an ECDSA private key from a hex string.
func HexToPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}

	privKey := new(ecdsa.PrivateKey)
	privKey.PublicKey.Curve = elliptic.P256()
	privKey.D = new(big.Int).SetBytes(b)
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(b)

	return privKey, nil
}
