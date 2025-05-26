package main

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func registerate(username string, password string, email string) error {
	if len(password) < 8 || len(username) < 2 {
		return errors.New("Password/Username feil")
	}

	userCheck := "select count(*) from users where name = $1"

	var count int

	err := db.QueryRow(userCheck, username).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("Dette brukernavnet er allerde i bruk")
	}

	if len(email) > 4 {
		matched, _ := regexp.Match(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`, []byte(email))

		if !matched {
			email = ""
		}
	} else {
		email = ""
	}

	hash, _ := hashPassword(password)

	_, err = db.Exec("insert into users(name, email, hash) values($1, $2, $3)", username, email, hash)

	if err != nil {
		return err
	}

	return nil
}

func loggingIn(sessionToken string, csrfToken string, username string, password string, w http.ResponseWriter) error {
	usercheck := "select hash from users where name = $1"

	var hash string

	err := db.QueryRow(usercheck, username).Scan(&hash)

	if err != nil || !checkPassword(password, hash) {
		return errors.New("Innloggings informasjon feil")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(168 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  time.Now().Add(168 * time.Hour),
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})

	_, err = db.Exec("update users set session = $1, csrf = $2 where name = $3", sessionToken, csrfToken, username)

	if err != nil {
		return err
	}

	return nil
}

func loggingOut(w http.ResponseWriter, id int) error {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: false,
	})

	_, err := db.Exec("update users set session = '', csrf = '' where id = $1", id)

	if err != nil {
		return err
	}

	return nil
}
