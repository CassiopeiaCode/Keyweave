package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var ErrInvalidMasterKey = errors.New("invalid master key")

type SealedValue struct {
	Ciphertext []byte
	Nonce      []byte
	KeyID      string
}

func Seal(masterKey []byte, keyID string, plaintext []byte) (SealedValue, error) {
	aead, err := newAEAD(masterKey)
	if err != nil {
		return SealedValue{}, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return SealedValue{}, err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return SealedValue{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		KeyID:      keyID,
	}, nil
}

func Open(masterKey []byte, ciphertext []byte, nonce []byte) ([]byte, error) {
	aead, err := newAEAD(masterKey)
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, ciphertext, nil)
}

func newAEAD(masterKey []byte) (cipher.AEAD, error) {
	switch len(masterKey) {
	case 16, 24, 32:
	default:
		return nil, ErrInvalidMasterKey
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
