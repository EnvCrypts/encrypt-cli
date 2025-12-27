package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	cryptutils "github.com/envcrypts/envcrypt_cli/internal/crypto"
	"github.com/google/uuid"
)

type AddEnvRequest struct {
	ProjectId uuid.UUID `json:"project_id"`
	Email     string    `json:"user_email"`

	EnvName    string `json:"env_name"`
	CipherText []byte `json:"cipher_text"`
	Nonce      []byte `json:"nonce"`
}

func PushEnv(projectId uuid.UUID, email string, privateKey []byte, wrappedKey *cryptutils.WrappedKey) error {

	// compress the file
	data := []byte("Hello World This is a wiordasdfjkh")

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
		CipherText: encryptedData,
		Nonce:      nonce,
	}
	requestBody, err := json.Marshal(AddEnvRequest)
	if err != nil {
		return err
	}

	// send to server
	resp, err := http.Post("http://localhost:8080/env/create", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
		return nil
	}

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
		Version:   2,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/env/search", "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseBody GetEnvResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		log.Println("Error pushing env", err.Error())
	}

	pmk, err := cryptutils.UnwrapPMK(wrappedKey, privateKey)
	if err != nil {
		log.Println("Pmk not unwrapped")
		return err
	}

	decryptedData, err := cryptutils.DecryptENV(pmk, responseBody.CipherText, responseBody.Nonce)
	if err != nil {
		log.Println("not decrypt")
	}

	log.Println(string(decryptedData))

	return nil
}

type UpdateEnvRequest struct {
	ProjectId uuid.UUID `json:"project_id"`
	Email     string    `json:"user_email"`

	EnvName    string `json:"env_name"`
	CipherText []byte `json:"cipher_text"`
	Nonce      []byte `json:"nonce"`
}

type UpdateEnvResponse struct {
	Message string `json:"message"`
}

func UpdateEnv(projectId uuid.UUID, email string, privateKey []byte, wrappedKey *cryptutils.WrappedKey) error {
	data := []byte("Hello World This is a wiordasdfjkh")

	pmk, err := cryptutils.UnwrapPMK(wrappedKey, privateKey)
	if err != nil {
		log.Println("Pmk not unwrapped")
		return err
	}

	// encrypt using pmk and store the nonce, ciphertext
	encryptedData, nonce, err := cryptutils.EncryptENV(pmk, data)
	if err != nil {
		log.Println("not encrypt")
		return err
	}

	var updateEnvRequest UpdateEnvRequest = UpdateEnvRequest{
		ProjectId:  projectId,
		Email:      email,
		EnvName:    "Testing",
		CipherText: encryptedData,
		Nonce:      nonce,
	}
	requestBody, err := json.Marshal(updateEnvRequest)
	if err != nil {
		return err
	}

	// send to server
	resp, err := http.Post("http://localhost:8080/env/update", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
		return nil
	}

	return nil
}

type GetEnvVersionsRequest struct {
	ProjectId uuid.UUID `json:"project_id"`
	Email     string    `json:"user_email"`

	EnvName string `json:"env_name"`
}

type EnvResponse struct {
	CipherText []byte `json:"cipher_text"`
	Nonce      []byte `json:"nonce"`
	Version    int32  `json:"version"`
}
type GetEnvVersionsResponse struct {
	EnvVersions []EnvResponse `json:"env_versions"`
}

func GetEnvVersions(projectId uuid.UUID, email string, privateKey []byte, wrappedKey *cryptutils.WrappedKey) error {

	var requestBody GetEnvVersionsRequest = GetEnvVersionsRequest{
		ProjectId: projectId,
		Email:     email,
		EnvName:   "Testing",
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/env/search/all", "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))
		return nil
	}

	var responseBody GetEnvVersionsResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		log.Println("Error pushing env", err.Error())
	}

	pmk, err := cryptutils.UnwrapPMK(wrappedKey, privateKey)
	if err != nil {
		log.Println("Pmk not unwrapped")
		return err
	}

	for _, envVersion := range responseBody.EnvVersions {
		decryptedData, err := cryptutils.DecryptENV(pmk, envVersion.CipherText, envVersion.Nonce)
		if err != nil {
			log.Println("not decrypt")
		}

		log.Printf("%d : %s", envVersion.Version, string(decryptedData))
	}

	return nil

}
