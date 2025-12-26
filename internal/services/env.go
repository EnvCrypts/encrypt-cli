package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	cryptutils "github.com/envcrypts/envcrypt_cli/internal/crypto"
	"github.com/google/uuid"
)

type AddEnvRequest struct {
	ProjectId uuid.UUID `json:"project_id"`
	Email     string    `json:"user_email"`

	EnvName    string `json:"env_name"`
	Version    int32  `json:"version"`
	CipherText []byte `json:"cipher_text"`
	Nonce      []byte `json:"nonce"`
}

func PushEnv(projectId uuid.UUID, email string, privateKey []byte, wrappedKey *cryptutils.WrappedKey) error {

	// compress the file
	data := []byte("Hello World")

	pmk, err := cryptutils.UnwrapPMK(wrappedKey, privateKey)
	if err != nil {
		log.Println("Pmk not unwrapped")
		return err
	}

	// encrypt using pmk and store the nonce, ciphertext
	encryptedData, nonce, err := cryptutils.EncryptENV(pmk, data)
	if err != nil {
		log.Println("not encruypt")
		return err
	}

	var AddEnvRequest AddEnvRequest = AddEnvRequest{
		ProjectId:  projectId,
		Email:      email,
		EnvName:    "Testing",
		Version:    1,
		CipherText: encryptedData,
		Nonce:      nonce,
	}
	requestBody, err := json.Marshal(AddEnvRequest)
	if err != nil {
		return err
	}

	// send to server
	resp, err := http.Post("http://localhost:8080/env/", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println(resp.Body)

	return nil
}

type GetEnvRequest struct {
	ProjectId uuid.UUID `json:"project_id"`
	Email     string    `json:"user_email"`

	EnvName string `json:"env_name"`
	Version int32  `json:"version"`
}

type GetEnvResponse struct {
	CipherText []byte `json:"cipher_text"`
	Nonce      []byte `json:"nonce"`
}

func PullEnv(projectId uuid.UUID, email string, privateKey []byte, wrappedKey *cryptutils.WrappedKey) error {

	var requestBody GetEnvRequest = GetEnvRequest{
		ProjectId: projectId,
		Email:     email,
		EnvName:   "Testing",
		Version:   1,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/env/", "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseBody GetEnvResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		log.Println("Error pushing env", err.Error())
	}
	log.Println(resp.StatusCode)

	pmk, err := cryptutils.UnwrapPMK(wrappedKey, privateKey)
	if err != nil {
		log.Println("Pmk not unwrapped")
		return err
	}

	log.Println(responseBody)

	decryptedData, err := cryptutils.DecryptENV(pmk, responseBody.CipherText, responseBody.Nonce)
	if err != nil {
		log.Println("not decrypt")
	}

	log.Println(string(decryptedData))

	return nil
}
