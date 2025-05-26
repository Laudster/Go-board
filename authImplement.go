package main

import (
	"errors"
	"net/http"
)

func registrering(r *http.Request, w http.ResponseWriter) error {
	if r.Method != http.MethodPost {
		return errors.New("Feil metode")
	}

	brukernavn := r.FormValue("navn")
	email := r.FormValue("email")
	passord := r.FormValue("passord")

	err := registerate(brukernavn, passord, email)

	if err != nil {
		return err
	}

	return loggeinn(brukernavn, passord, w)
}

func loggeinn(brukernavn string, passord string, w http.ResponseWriter) error {
	sessionToken, _ := generateToken(32)
	csrfToken, _ := generateToken(32)

	return loggingIn(sessionToken, csrfToken, brukernavn, passord, w)
}

func loggeUt(r *http.Request, w http.ResponseWriter) error {
	if r.Method != http.MethodPost {
		return errors.New("Feil metode")
	}

	user, err := getUser(r)

	if err != nil {
		return errors.New("Uautorisert bruker")
	}

	err = csrfCheck(r, user.Csrf)

	if err != nil {
		return errors.New("Uautorisert csrf")
	}

	return loggingOut(w, user.Id)
}
