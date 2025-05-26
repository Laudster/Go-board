package main

import (
	"errors"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Id      int
	Name    string
	Hash    string
	Admin   bool
	Session string
	Csrf    string
}

func getUser(r *http.Request) (User, error) {
	var user User

	sessionToken, err := r.Cookie("session_token")

	if err != nil {
		return user, err
	}

	userCheck := "select id, name, hash, admin, session, csrf from users where session = $1"

	err = db.QueryRow(userCheck, sessionToken.Value).Scan(&user.Id, &user.Name, &user.Hash, &user.Admin, &user.Session, &user.Csrf)

	if err != nil {
		return user, err
	}

	return user, nil
}

func csrfCheck(r *http.Request, csrfToken string) error {
	csrf := r.FormValue("csrf_token")

	if csrfToken != csrf || csrf == "" {
		return errors.New("Unauthorized")
	}

	return nil
}
