package main

import (
	"log"

	"github.com/envcrypts/envcrypt_cli/internal/services"
)

func main() {
	err := services.Register("vijayvenka@gmail.com", "vijayvenkatj")
	if err != nil {
		log.Println(err.Error())
	}

	keypair, userId, err := services.Login("vijayvenka@gmail.com", "vijayvenkatj")
	if err != nil {
		log.Fatal("Error Login:", err)
	}

	err = services.CreateProject("dummyx", *userId, keypair.PublicKey)
	if err != nil {
		log.Println("Error project creation", err.Error())
	}

	wrappedKeys, projectId, err := services.GetProject("dummyx", *userId)
	if err != nil {
		log.Println("Error wrappedKey get", err.Error())
	}

	err = services.PushEnv(*projectId, "vijayvenka@gmail.com", keypair.PrivateKey, wrappedKeys)
	if err != nil {
		log.Println("Error pushing env", err.Error())
	}

	err = services.UpdateEnv(*projectId, "vijayvenka@gmail.com", keypair.PrivateKey, wrappedKeys)
	if err != nil {
		log.Println("Error updating env", err.Error())
	}

	err = services.PullEnv(*projectId, "vijayvenka@gmail.com", keypair.PrivateKey, wrappedKeys)
	if err != nil {
		log.Println("Error pulling env", err.Error())
	}

	err = services.GetEnvVersions(*projectId, "vijayvenka@gmail.com", keypair.PrivateKey, wrappedKeys)
	if err != nil {
		log.Println("Error getting env versions", err.Error())
	}

}
