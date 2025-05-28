package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func createDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "Db/data.db")

	if err != nil {
		return db, err
	}

	db.Exec("PRAGMA journal_mode=WAL;")

	sqlStatement := `
		create table if not exists users (
			id integer not null primary key autoincrement,
			name text,
			email text,
			hash text,
			admin bool default false,
			session text,
			csrf text
		);

		create table if not exists boards (
			id integer not null primary key autoincrement,
			name text not null,
			desc text not null,
			created_by integer not null,
			created_at datetime,

			foreign key (created_by) references users(id) on delete cascade
		);

		create table if not exists posts (
			id integer not null primary key autoincrement,
			title text not null,
			body text not null,
			image blob,
			created_by integer not null,
			created_at datetime not null,
			board integer not null,

			foreign key (created_by) references users(id) on delete cascade,
			foreign key (board) references boards(id) on delete cascade
		);

		create table if not exists comments (
			id integer not null primary key autoincrement,
			body text not null,
			image blob,
			created_by integer not null,
			created_at datetime not null,
			post integer not null,

			foreign key (created_by) references users(id) on delete cascade,
			foreign key (post) references posts(id) on delete cascade
		);

		create table if not exists upvotes (
			id integer not null primary key autoincrement,
			created_by integer not null,
			post integer not null,
			charge bool not null,
			comment bool not null,

			foreign key (created_by) references users(id) on delete cascade,
			foreign key (post) references posts(id) on delete cascade
		);
	`

	_, err = db.Exec(sqlStatement)

	if err != nil {
		return db, err
	}

	return db, nil
}
