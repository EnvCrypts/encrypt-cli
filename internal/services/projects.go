package services

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"

	cryptutils "github.com/envcrypts/envcrypt_cli/internal/crypto"
	"github.com/google/uuid"
)

type ProjectCreateRequest struct {
	Name               string    `json:"name"`
	UserId             uuid.UUID `json:"user_id"`
	WrappedPMK         []byte    `json:"wrapped_pmk"`
	WrapNonce          []byte    `json:"wrap_nonce"`
	EphemeralPublicKey []byte    `json:"ephemeral_public_key"`
}

func CreateProject(name string, userId uuid.UUID, publicKey []byte) error {

	pmk := make([]byte, 32)
	_, err := rand.Read(pmk)
	if err != nil {
		return err
	}

	wrappedKey, err := cryptutils.WrapPMKForUser(pmk, publicKey)
	if err != nil {
		return err
	}

	projectRequest := ProjectCreateRequest{
		Name:               name,
		UserId:             userId,
		WrappedPMK:         wrappedKey.WrappedPMK,
		WrapNonce:          wrappedKey.WrapNonce,
		EphemeralPublicKey: wrappedKey.WrapEphemeralPub,
	}

	requestBody, err := json.Marshal(projectRequest)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/projects/create", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

type GetUserProjectRequest struct {
	ProjectName string    `json:"project_name"`
	UserId      uuid.UUID `json:"user_id"`
}
type GetUserProjectResponse struct {
	ProjectId          uuid.UUID `json:"project_id"`
	WrappedPMK         []byte    `json:"wrapped_pmk"`
	WrapNonce          []byte    `json:"wrap_nonce"`
	EphemeralPublicKey []byte    `json:"ephemeral_public_key"`
}

func GetProject(projectName string, userId uuid.UUID) (*cryptutils.WrappedKey, *uuid.UUID, error) {
	var requestBody GetUserProjectRequest = GetUserProjectRequest{
		ProjectName: projectName,
		UserId:      userId,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, nil, err
	}

	resp, err := http.Post("http://localhost:8080/projects/keys", "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var responseBody GetUserProjectResponse
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		log.Println("ERROR in JSON decoding")
		return nil, nil, err
	}

	return &cryptutils.WrappedKey{
		WrappedPMK:       responseBody.WrappedPMK,
		WrapNonce:        responseBody.WrapNonce,
		WrapEphemeralPub: responseBody.EphemeralPublicKey,
	}, &responseBody.ProjectId, nil
}
