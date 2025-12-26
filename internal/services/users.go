package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	cryptutils "github.com/envcrypts/envcrypt_cli/internal/crypto"
	"github.com/google/uuid"
)

type CreateRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	PublicKey               []byte `json:"public_key"`
	EncryptedUserPrivateKey []byte `json:"encrypted_user_private_key"`
	PrivateKeySalt          []byte `json:"private_key_salt"`
	PrivateKeyNonce         []byte `json:"private_key_nonce"`
}

type UserBody struct {
	Id                      uuid.UUID                 `json:"id"`
	Email                   string                    `json:"email"`
	PublicKey               []byte                    `json:"public_key"`
	EncryptedUserPrivateKey []byte                    `json:"encrypted_user_private_key"`
	PrivateKeySalt          []byte                    `json:"private_key_salt"`
	PrivateKeyNonce         []byte                    `json:"private_key_nonce"`
	ArgonParams             cryptutils.Argon2idParams `json:"argon_params"`
}

type LoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponseBody struct {
	Message string   `json:"message"`
	User    UserBody `json:"user"`
}

func Register(email, password string) error {
	keypair, err := cryptutils.GenerateKeyPair(password)
	if err != nil {
		return err
	}
	
	var RequestBody = CreateRequestBody{
		Email:                   email,
		Password:                password,
		PublicKey:               keypair.PublicKey,
		EncryptedUserPrivateKey: keypair.EncKey.EncryptedUserPrivateKey,
		PrivateKeySalt:          keypair.EncKey.PrivateKeySalt,
		PrivateKeyNonce:         keypair.EncKey.PrivateKeyNonce,
	}

	requestBody, err := json.Marshal(RequestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/users/create", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(responseBody))
	return nil
}

func Login(email, password string) (*cryptutils.KeyPair, *uuid.UUID, error) {

	var RequestBody = LoginRequestBody{
		Email:    email,
		Password: password,
	}
	requestBody, err := json.Marshal(RequestBody)

	resp, err := http.Post("http://localhost:8080/users/login", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var LoginResponse LoginResponseBody

	err = json.NewDecoder(resp.Body).Decode(&LoginResponse)
	if err != nil {
		return nil, nil, err
	}

	encryptedKey := &cryptutils.EncryptedPrivateKey{
		EncryptedUserPrivateKey: LoginResponse.User.EncryptedUserPrivateKey,
		PrivateKeySalt:          LoginResponse.User.PrivateKeySalt,
		PrivateKeyNonce:         LoginResponse.User.PrivateKeyNonce,
	}
	privateKey, err := cryptutils.DecryptPrivateKey(encryptedKey, password, &LoginResponse.User.ArgonParams)
	if err != nil {
		return nil, nil, err
	}

	keyPair := &cryptutils.KeyPair{
		PublicKey:  LoginResponse.User.PublicKey,
		PrivateKey: privateKey,
		EncKey:     *encryptedKey,
	}

	return keyPair, &LoginResponse.User.Id, nil
}
