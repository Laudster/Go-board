package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	return string(bytes), err
}

func checkPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

func formatDate(tid time.Time) string {
	siden := time.Since(tid)

	if siden.Seconds() < 60 {
		return fmt.Sprintf("%.0f sekunder", siden.Seconds())
	} else if siden.Minutes() < 60 {
		if siden.Minutes() < 2 {
			return "Ett minutt"
		}

		return fmt.Sprintf("%.0f minutter", siden.Minutes())
	} else if siden.Hours() < 24 {
		if siden.Hours() < 2 {
			return "1 time"
		}

		return fmt.Sprintf("%.0f timer", siden.Hours())
	} else {
		if siden.Hours() < 48.0 {
			return "En dag"
		}

		return fmt.Sprintf("%.0f dager", siden.Hours()/24)
	}
}
