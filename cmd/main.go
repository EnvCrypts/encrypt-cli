package main

import (
	"log"

	"github.com/envcrypts/envcrypt_cli/internal/services"
)

func main() {
	err := services.Register("vijay213@gmail.com", "vijayvenkatj")
	if err != nil {
		log.Println(err.Error())
	}

	keypair, userId, err := services.Login("vijay213@gmail.com", "vijayvenkatj")
	if err != nil {

		log.Fatal("Error Login:", err)
	}

	err = services.CreateProject("dummy5", *userId, keypair.PublicKey)
	if err != nil {
		log.Println("Error project creation", err.Error())
	}

	log.Println("User created", *userId)

	wrappedKeys, projectId, err := services.GetProject("dummy5", *userId)
	if err != nil {
		log.Println("Error wrappedKey get", err.Error())
	}
	log.Println("Wrapped keys:", wrappedKeys)

	err = services.PushEnv(*projectId, "vijay213@gmail.com", keypair.PrivateKey, wrappedKeys)
	if err != nil {
		log.Println("Error pushing env", err.Error())
	}

	err = services.PullEnv(*projectId, "vijay213@gmail.com", keypair.PrivateKey, wrappedKeys)
	if err != nil {
		log.Println("Error pulling env", err.Error())
	}

}
