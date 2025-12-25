package cryptutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/argon2"
)

type Argon2idParams struct {
	Time        uint32 `json:"time"`
	Memory      uint32 `json:"memory"`
	Parallelism uint8  `json:"parallelism"`
	KeyLength   uint32 `json:"key_length"`
}

var DefaultArgon2Params = Argon2idParams{
	Time:        3,
	Memory:      64 * 1024, // 64 MB
	Parallelism: 1,
	KeyLength:   32,
}

type KeyPair struct {
	PublicKey  []byte              `json:"public_key"`
	PrivateKey []byte              `json:"private_key"`
	EncKey     EncryptedPrivateKey `json:"encrypted_private_key"`
}
type EncryptedPrivateKey struct {
	EncryptedUserPrivateKey []byte `json:"encrypted_user_private_key"`
	PrivateKeySalt          []byte `json:"private_key_salt"`
	PrivateKeyNonce         []byte `json:"private_key_nonce"`
}

func encryptPrivateKey(priv *ecdh.PrivateKey, password string, params *Argon2idParams) (*EncryptedPrivateKey, error) {

	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	encryptionKey := argon2.IDKey(
		[]byte(password),
		salt,
		params.Time,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	encryptedPrivateKey := gcm.Seal(nil, nonce, priv.Bytes(), nil)
	return &EncryptedPrivateKey{
		EncryptedUserPrivateKey: encryptedPrivateKey,
		PrivateKeySalt:          salt,
		PrivateKeyNonce:         nonce,
	}, nil
}

func DecryptPrivateKey(
	enc *EncryptedPrivateKey,
	password string,
	params *Argon2idParams,
) ([]byte, error) {

	// 1. Derive the same encryption key using Argon2id
	encryptionKey := argon2.IDKey(
		[]byte(password),
		enc.PrivateKeySalt,
		params.Time,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// 2. Create AES block cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// 3. Create GCM instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 4. Decrypt (Open = authenticated decrypt)
	plaintextPrivKey, err := gcm.Open(
		nil,
		enc.PrivateKeyNonce,
		enc.EncryptedUserPrivateKey,
		nil,
	)
	if err != nil {
		// This error covers:
		// - wrong password
		// - corrupted ciphertext
		// - wrong nonce
		// - tampering
		return nil, err
	}

	// 5. Sanity check (X25519 private keys are 32 bytes)
	if len(plaintextPrivKey) != 32 {
		return nil, errors.New("invalid private key length")
	}

	return plaintextPrivKey, nil
}

func GenerateKeyPair(password string) (*KeyPair, error) {
	curve := ecdh.X25519()

	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	encryptedKey, err := encryptPrivateKey(priv, password, &DefaultArgon2Params)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PrivateKey: priv.Bytes(),
		PublicKey:  priv.PublicKey().Bytes(),
		EncKey:     *encryptedKey,
	}, nil
}
