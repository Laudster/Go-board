package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var templates = template.Must(template.ParseFiles("Templates/index.html", "Templates/brett.html", "Templates/post.html", "Templates/admin.html"))

var db *sql.DB

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

	data := map[string]any{
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

	rows, err := db.Query("select id, title, body, image, created_by, created_at from posts where board = $1 order by created_at desc", id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var posts []Post

	for rows.Next() {
		post, err := skaffPost(brettet, rows, false)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		posts = append(posts, post)
	}

	sort.Slice(posts, func(x, y int) bool {
		return posts[x].Upvotes > posts[y].Upvotes
	})

	data := map[string]any{
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

func post(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var brett int

	err := db.QueryRow("select board from posts where id = $1", id).Scan(&brett)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var brettet string

	db.QueryRow("select name from boards where id = $1", brett).Scan(&brettet)

	row, err := db.Query("select id, title, body, image, created_by, created_at from posts where id = $1", id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	row.Next()

	post, err := skaffPost(brettet, row, false)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var charge bool

	user, logged := getUser(r)

	if logged == nil {
		err = db.QueryRow("select charge from upvotes where created_by = $1 and post = $2 and comment = false", user.Id, post.Id).Scan(&charge)

		if err == nil {
			if charge {
				post.Stemt = "po"
			} else {
				post.Stemt = "ne"
			}
		}
	}

	rows, err := db.Query("select id, body, image, created_by, created_at from comments where post = $1", post.Id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var kommentarer []Post

	for rows.Next() {
		kommentar, err := skaffPost("", rows, true)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if logged == nil {
			err = db.QueryRow("select charge from upvotes where created_by = $1 and post = $2 and comment = true", user.Id, kommentar.Id).Scan(&charge)

			if err == nil {
				if charge {
					kommentar.Stemt = "po"
				} else {
					kommentar.Stemt = "ne"
				}
			}
		}

		kommentarer = append(kommentarer, kommentar)
	}

	sort.Slice(kommentarer, func(x, y int) bool {
		return kommentarer[x].Upvotes > kommentarer[y].Upvotes
	})

	data := map[string]any{
		"Post":        post,
		"Kommentarer": kommentarer,
	}

	err = templates.ExecuteTemplate(w, "post.html", data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	if !user.Admin {
		return
	}

	brett := r.FormValue("Brett")
	beskrivelse := r.FormValue("beskrivelse")

	db.Exec("insert into boards (name, desc, created_by, created_at) values($1, $2, $3, $4)", brett, beskrivelse, user.Id, time.Now())

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func nyPost(w http.ResponseWriter, r *http.Request) {
	user, err := getUser(r)

	if err != nil {
		http.Error(w, "Du trenger å være innlogget", http.StatusUnauthorized)
		return
	}

	r.ParseMultipartForm(5 << 20)

	tittel := r.FormValue("tittel")
	tekst := r.FormValue("tekst")

	file, header, err := r.FormFile("bilde")

	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Image uplaod error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	id := user.Id

	sender := strings.Split(r.Referer(), "/")[4]

	var board int

	err = db.QueryRow("select id from boards where name = $1", sender).Scan(&board)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if header == nil || header.Filename == "" {
		_, err = db.Exec("insert into posts (title, body, created_by, created_at, board) values($1, $2, $3, $4, $5)", tittel, tekst, id, time.Now(), board)
	} else {
		defer file.Close()

		imageData, _ := io.ReadAll(file)
		_, err = db.Exec("insert into posts (title, body, image, created_by, created_at, board) values($1, $2, $3, $4, $5, $6)", tittel, tekst, imageData, id, time.Now(), board)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func kommenter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method", http.StatusMethodNotAllowed)
		return
	}

	user, err := getUser(r)

	if err != nil {
		http.Error(w, "Du trenger å være innlogget for å komentere", http.StatusUnauthorized)
		return
	}

	r.ParseMultipartForm(5 << 20)

	tekst := r.FormValue("tekst")

	file, header, err := r.FormFile("bilde")

	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Image uplaod error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sender := strings.Split(r.Referer(), "/")[4]

	post, _ := strconv.Atoi(sender)

	if header == nil || header.Filename == "" {
		_, err = db.Exec("insert into comments (body, created_by, created_at, post) values($1, $2, $3, $4)", tekst, user.Id, time.Now(), post)
	} else {
		defer file.Close()

		imageData, _ := io.ReadAll(file)
		_, err = db.Exec("insert into comments (body, image, created_by, created_at, post) values($1, $2, $3, $4, $5)", tekst, imageData, user.Id, time.Now(), post)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

func vote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong method", http.StatusMethodNotAllowed)
		return
	}

	user, err := getUser(r)

	if err != nil {
		http.Error(w, "You need to be logged to vote", http.StatusInternalServerError)
		return
	}

	post := r.FormValue("post")

	var charge bool

	formCharge := r.FormValue("charge")

	var comment bool

	if r.FormValue("type") == "post" {
		comment = false
	} else {
		comment = true
	}

	if formCharge == "↑" {
		charge = true
	} else if formCharge == "↓" {
		charge = false
	} else {
		db.Exec("delete from upvotes where created_by = $1 and post = $2 and comment = $3", user.Id, post, comment)

		http.Redirect(w, r, r.Referer(), http.StatusFound)
		return
	}

	var count int

	_ = db.QueryRow("select count(*) from upvotes where created_by = $1 and post = $2 and comment = $3", user.Id, post, comment).Scan(&count)

	if count > 0 {
		_, err = db.Exec("update upvotes set charge = $1 where created_by = $2 and post = $3 and comment = $4", charge, user.Id, post, comment)
	} else {
		_, err = db.Exec("insert into upvotes (created_by, post, charge, comment) values($1, $2, $3, $4)", user.Id, post, charge, comment)
	}

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
	http.HandleFunc("/kommenter", kommenter)
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/upvote", vote)
	http.HandleFunc("/post/{id}", post)
	http.HandleFunc("/post", nyPost)
	http.HandleFunc("/brett/{brett}", brett)
	http.HandleFunc("/", index)

	fmt.Println("Running on http://localhost:50")

	log.Fatal(http.ListenAndServe(":50", nil))
}
