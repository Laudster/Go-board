package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var templates = template.Must(template.ParseFiles("Templates/index.html", "Templates/brett.html", "Templates/admin.html"))

var db *sql.DB

type Post struct {
	Tittel string
	Tekst  string
	Brett  string
	Skapt  string
	Skaper string
}

func index(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select name from boards")

	defer rows.Close()

	var brett []string

	for rows.Next() {
		var navn string
		rows.Scan(&navn)

		brett = append(brett, navn)
	}

	user, _ := getUser(r)

	data := map[string]interface{}{
		"Navn":  user.Name,
		"Brett": brett,
	}

	err = templates.ExecuteTemplate(w, "index.html", data)

	if err != nil {
		http.Error(w, "Kunne ikke laste template", http.StatusInternalServerError)
		return
	}
}

func admin(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)

	if err != nil {
		http.Error(w, "Du er ikke admin >:", http.StatusUnauthorized)
		return
	}

	if user.Admin != true {
		http.Error(w, "Du er ikke admin", http.StatusUnauthorized)
		return
	}

	err = templates.ExecuteTemplate(w, "admin.html", user.Name)
}

func registrer(w http.ResponseWriter, r *http.Request) {
	err := registrering(r, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func loggInn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Feil metode", http.StatusMethodNotAllowed)
		return
	}

	brukernavn := r.FormValue("navn")
	passord := r.FormValue("passord")

	err := loggeinn(brukernavn, passord, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func loggut(w http.ResponseWriter, r *http.Request) {
	err := loggeUt(r, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func brett(w http.ResponseWriter, r *http.Request) {
	brettet := r.PathValue("brett")

	var beskrivelse string

	var id int

	err := db.QueryRow("select id, desc from boards where name = $1", brettet).Scan(&id, &beskrivelse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("select title, body, created_by, created_at, board from posts where board = $1", id)

	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var tittel, tekst string
		var skaper, brett int
		var skapt time.Time

		rows.Scan(&tittel, &tekst, &skaper, &skapt, &brett)

		var post = Post{
			Tittel: tittel,
			Tekst:  tekst,
			Brett:  strconv.Itoa(brett),
			Skapt:  skapt.Format("2006-01-02 15:04:05"),
			Skaper: strconv.Itoa(skaper),
		}

		posts = append(posts, post)
	}

	data := map[string]interface{}{
		"Brettet":     brettet,
		"Beskrivelse": beskrivelse,
		"Posts":       posts,
	}

	err = templates.ExecuteTemplate(w, "brett.html", data)

	if err != nil {
		http.Error(w, "Kunne ikke laste inn "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func nyttBrett(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)

	if err != nil {
		return
	}

	if csrfCheck(r, user.Csrf) != nil {
		return
	}

	if user.Admin != true {
		return
	}

	brett := r.FormValue("Brett")
	beskrivelse := r.FormValue("beskrivelse")

	db.Exec("insert into boards (name, desc, created_by, created_at) values($1, $2, $3, $4)", brett, beskrivelse, user.Id, time.Now())

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func nyPost(w http.ResponseWriter, r *http.Request) {
	tittel := r.FormValue("tittel")
	tekst := r.FormValue("tekst")

	user, err := getUser(r)

	id := 0

	if err == nil {
		id = user.Id
	}

	sender := strings.Split(r.Referer(), "/")[4]

	var board int

	err = db.QueryRow("select id from boards where name = $1", sender).Scan(&board)

	if err != nil {
		fmt.Print(sender)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("insert into posts (title, body, created_by, created_at, board) values($1, $2, $3, $4, $5)", tittel, tekst, id, time.Now(), board)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func main() {
	fs := http.FileServer(http.Dir("Static"))
	http.Handle("/Static/", http.StripPrefix("/Static/", fs))

	db, _ = createDB()

	defer db.Close()

	http.HandleFunc("/loggut", loggut)
	http.HandleFunc("/logginn", loggInn)
	http.HandleFunc("/registrer", registrer)
	http.HandleFunc("/nytt-brett", nyttBrett)
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/post", nyPost)
	http.HandleFunc("/brett/{brett}", brett)
	http.HandleFunc("/", index)
	log.Fatal(http.ListenAndServe(":50", nil))
}
