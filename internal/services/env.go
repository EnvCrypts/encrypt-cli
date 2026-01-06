package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	fileData, err := os.ReadFile("/home/vijay/Projects/encrypt-cli/key.txt")
	if err != nil {
		return err
	}

	data, err := cryptutils.PrepareEnvForStorage(fileData)
	if err != nil {
		return err
	}

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

func PullEnv(projectId uuid.UUID, email string, privateKey []byte, version int32, wrappedKey *cryptutils.WrappedKey) (map[string]string, error) {

	var requestBody GetEnvRequest = GetEnvRequest{
		ProjectId: projectId,
		Email:     email,
		EnvName:   "Testing",
		Version:   version,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:8080/env/search", "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	decryptedData, err := cryptutils.DecryptENV(pmk, responseBody.CipherText, responseBody.Nonce)
	if err != nil {
		log.Println("not decrypt")
	}

	parsedEnv, err := cryptutils.ReadEnvFromStorage(decryptedData)
	if err != nil {
		return nil, err
	}

	for k, v := range parsedEnv {
		log.Println(k, v)
	}

	return parsedEnv, nil
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

	// compress the file
	fileData, err := os.ReadFile("/home/vijay/Projects/encrypt-cli/key.txt")
	if err != nil {
		return err
	}

	data, err := cryptutils.PrepareEnvForStorage(fileData)
	if err != nil {
		return err
	}

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

		readableData, err := cryptutils.ReadEnvFromStorage(decryptedData)
		if err != nil {
			log.Println("not readable")
		}

		log.Printf("%d : %s", envVersion.Version, readableData)
	}

	if len(responseBody.EnvVersions) > 2 {

		type ReadableData struct {
			oldVersion map[string]string
			newVersion map[string]string
		}

		var readableData ReadableData

		for i := len(responseBody.EnvVersions) - 2; i < len(responseBody.EnvVersions); i++ {

			envVersion := responseBody.EnvVersions[i]

			decryptedData, err := cryptutils.DecryptENV(pmk, envVersion.CipherText, envVersion.Nonce)
			if err != nil {
				log.Println("not decrypt")
			}

			data, err := cryptutils.ReadEnvFromStorage(decryptedData)
			if err != nil {
				log.Println("not readable")
			}

			if i == 0 {
				readableData.oldVersion = data
			} else {
				readableData.newVersion = data
			}
		}

		_ = cryptutils.DiffEnvVersions(readableData.oldVersion, readableData.newVersion)
	}

	return nil

}

func DiffENVVersions(projectId uuid.UUID, email string, privateKey []byte, wrappedKey *cryptutils.WrappedKey, oldVersion, newVersion int32) error {

	oldVersionEnv, err := PullEnv(projectId, email, privateKey, oldVersion, wrappedKey)
	if err != nil {
		return err
	}
	newVersionEnv, err := PullEnv(projectId, email, privateKey, newVersion, wrappedKey)
	if err != nil {
		return err
	}

	diffingResult := cryptutils.DiffEnvVersions(oldVersionEnv, newVersionEnv)
	log.Println("ADDED", diffingResult.Added)
	log.Println("Modified", diffingResult.Modified)
	log.Println("Removed", diffingResult.Removed)

	return nil
}
