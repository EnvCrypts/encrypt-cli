package main

import (
	"log"

	"github.com/envcrypts/envcrypt_cli/internal/services"
)

func main() {
	err := services.Register("vijay1@gmail.com", "vijayvenkatj")
	if err != nil {
		log.Fatal(err)
	}

	keypair, userId, err := services.Login("vijay1@gmail.com", "vijayvenkatj")
	if err != nil {
		log.Fatal(err)
	}

	err = services.CreateProject("dummt", *userId, keypair.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
}
