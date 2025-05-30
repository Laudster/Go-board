package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v2"
)

func lagKlient() *resend.Client {
	godotenv.Load()

	client := resend.NewClient(os.Getenv("emailAPI"))

	return client
}
