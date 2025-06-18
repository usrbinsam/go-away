package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Store interface {
	Open(db string) error
	RecordUnsubscribe(messageId, listId, recipient string)
	Unsubscribed(listId, recipient string) bool
	MarkSeen(messageId, recipient string) string
	Seen(messageId, recipient string) bool
}

type SqlStore struct {
	db *sql.DB
}

func (ss *SqlStore) Open(db string) error {
	var err error
	ss.db, err = sql.Open("sqlite", db)
	if err != nil {
		return err
	}

	ss.createAll()
	return ss.db.Ping()
}

func (ss *SqlStore) createAll() {
	ddl := `
create table if not exists unsubscribes (
	id integer primary key autoincrement,
	ts timestamp default current_timestamp,
	message_id text not null,
	list_id text not null,
	recipient text not null
);
create table if not exists seen (
	id integer primary key autoincrement,
	ts timestamp default current_timestamp,
	message_id text not null,
	recipient text not null
);
	`
	_, err := ss.db.Exec(ddl)
	if err != nil {
		panic("failed to create tables: " + err.Error())
	}
}

func (ss *SqlStore) RecordUnsubscribe(messageId, listId, recipient string) {
	stmt, err := ss.db.Prepare("insert into unsubscribes (message_id, list_id, recipient) values (?, ?, ?)")
	if err != nil {
		panic("store: RecordUnsubscribe prepare stmt failed: " + err.Error())
	}
	stmt.Exec(messageId, listId, recipient)
}

func (ss *SqlStore) Unsubscribed(listId, recipient string) bool {
	stmt, err := ss.db.Prepare("select count(*) from unsubscribes where list_id = ? and recipient = ?")
	if err != nil {
		panic("store: Unsubscribed prepare stmt failed: " + err.Error())
	}

	var count uint8
	err = stmt.QueryRow(listId, recipient).Scan(&count)
	if err != nil {
		panic("store: Unsubscribed query failed: " + err.Error())
	}

	return count >= 1
}

func (ss *SqlStore) MarkSeen(messageId, recipient string) string {
	stmt, err := ss.db.Prepare("insert into seen (message_id, recipient) values (?, ?)")
	if err != nil {
		panic("store: MarkSeen prepare stmt failed: " + err.Error())
	}

	_, err = stmt.Exec(messageId, recipient)
	if err != nil {
		panic("store: MarkSeen exec stmt failed: " + err.Error())
	}

	return messageId
}

func (ss *SqlStore) Seen(messageId, recipient string) bool {
	stmt, err := ss.db.Prepare("select count(*) from seen where message_id = ? and recipient = ?")
	if err != nil {
		panic("store: Seen prepare stmt failed: " + err.Error())
	}

	var count uint8
	err = stmt.QueryRow(messageId, recipient).Scan(&count)
	if err != nil {
		panic("store: Seen query failed: " + err.Error())
	}

	return count >= 1
}
