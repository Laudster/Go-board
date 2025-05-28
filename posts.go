package main

import (
	"database/sql"
	"encoding/base64"
	"time"
)

type Post struct {
	Id      int
	Tittel  string
	Tekst   string
	Bilde   string
	Brett   string
	Skapt   string
	Skaper  string
	Upvotes int
	Stemt   string
}

func skaffPost(brettet string, row *sql.Rows, kommentar bool) (Post, error) {
	var post Post

	var tittel, tekst string
	var id, skaper, upvotes int
	var skapt time.Time
	var bilde []byte

	var err error

	if kommentar {
		err = row.Scan(&id, &tekst, &bilde, &skaper, &skapt)
	} else {
		err = row.Scan(&id, &tittel, &tekst, &bilde, &skaper, &skapt)
	}

	if err != nil {
		return post, err
	}

	var bruker string

	err = db.QueryRow("select name from users where id = $1", skaper).Scan(&bruker)

	if err != nil {
		return post, err
	}

	upvoteQuery := "SELECT SUM(CASE WHEN charge = TRUE THEN 1 ELSE -1 END) AS total_upvotes FROM upvotes WHERE post = $1 AND comment = $2"

	db.QueryRow(upvoteQuery, id, kommentar).Scan(&upvotes)

	post = Post{
		Id:      id,
		Tittel:  tittel,
		Tekst:   tekst,
		Brett:   brettet,
		Bilde:   base64.StdEncoding.EncodeToString(bilde),
		Skapt:   formatDate(skapt),
		Skaper:  bruker,
		Upvotes: upvotes,
		Stemt:   "usikker",
	}

	return post, nil
}
