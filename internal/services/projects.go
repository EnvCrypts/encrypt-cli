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

	log.Print(pmk)

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
