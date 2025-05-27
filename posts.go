package main

import (
	"database/sql"
	"time"
)

func skaffPost(brettet string, row *sql.Rows) (Post, error) {
	var post Post

	var tittel, tekst string
	var id, skaper int
	var skapt time.Time

	err := row.Scan(&id, &tittel, &tekst, &skaper, &skapt)

	if err != nil {
		return post, err
	}

	var bruker string

	err = db.QueryRow("select name from users where id = $1", skaper).Scan(&bruker)

	if err != nil {
		return post, err
	}

	post = Post{
		Id:     id,
		Tittel: tittel,
		Tekst:  tekst,
		Brett:  brettet,
		Skapt:  formatDate(skapt),
		Skaper: bruker,
	}

	return post, nil
}
