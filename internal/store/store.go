package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Store interface {
	Open(db string) error
	RecordUnsubscribe(messageID, listID, recipient string)
	Unsubscribed(listID, recipient string) bool
	MarkSeen(messageID, recipient string) string
	Seen(messageID, recipient string) bool
}

type SQLStore struct {
	db *sql.DB
}

func (ss *SQLStore) Open(db string) error {
	var err error
	ss.db, err = sql.Open("sqlite", db)
	if err != nil {
		return err
	}

	ss.createAll()
	return ss.db.Ping()
}

func (ss *SQLStore) createAll() {
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
create table if not exists inboxes (
	id integer primary key autoincrement,
	addr text not null,
	provider text not null,
	oauth2_access_token text,
	oauth2_refresh_token text
);
create table if not exists config (
	inbox_id integer not null,
	key text not null,
	value text,
	primary key(inbox_id, key),
	foreign key(inbox_id) references inboxes(id)
);
	`
	_, err := ss.db.Exec(ddl)
	if err != nil {
		panic("failed to create tables: " + err.Error())
	}
}

func (ss *SQLStore) RecordUnsubscribe(messageID, listID, recipient string) {
	stmt, err := ss.db.Prepare("insert into unsubscribes (message_id, list_id, recipient) values (?, ?, ?)")
	if err != nil {
		panic("store: RecordUnsubscribe prepare stmt failed: " + err.Error())
	}
	stmt.Exec(messageID, listID, recipient)
}

func (ss *SQLStore) Unsubscribed(listID, recipient string) bool {
	stmt, err := ss.db.Prepare("select count(*) from unsubscribes where list_id = ? and recipient = ?")
	if err != nil {
		panic("store: Unsubscribed prepare stmt failed: " + err.Error())
	}

	var count uint8
	err = stmt.QueryRow(listID, recipient).Scan(&count)
	if err != nil {
		panic("store: Unsubscribed query failed: " + err.Error())
	}

	return count >= 1
}

func (ss *SQLStore) MarkSeen(messageID, recipient string) string {
	stmt, err := ss.db.Prepare("insert into seen (message_id, recipient) values (?, ?)")
	if err != nil {
		panic("store: MarkSeen prepare stmt failed: " + err.Error())
	}

	_, err = stmt.Exec(messageID, recipient)
	if err != nil {
		panic("store: MarkSeen exec stmt failed: " + err.Error())
	}

	return messageID
}

func (ss *SQLStore) Seen(messageID, recipient string) bool {
	stmt, err := ss.db.Prepare("select count(*) from seen where message_id = ? and recipient = ?")
	if err != nil {
		panic("store: Seen prepare stmt failed: " + err.Error())
	}

	var count uint8
	err = stmt.QueryRow(messageID, recipient).Scan(&count)
	if err != nil {
		panic("store: Seen query failed: " + err.Error())
	}

	return count >= 1
}

type Inbox struct {
	ID       int
	Addr     string
	Provider string
}

func (ss *SQLStore) ListInboxes() []Inbox {
	stmt, err := ss.db.Prepare("select id, addr, provider from inboxes")
	if err != nil {
		panic("store: ListInboxes prepare stmt failed: " + err.Error())
	}

	rows, err := stmt.Query()
	if err != nil {
		panic("store: ListInboxes query failed: " + err.Error())
	}

	inboxes := make([]Inbox, 0)
	for rows.Next() {
		var inbox Inbox
		rows.Scan(&inbox.ID, &inbox.Addr, &inbox.Provider)
		inboxes = append(inboxes, inbox)
	}

	return inboxes
}

func (ss *SQLStore) ConfigSet(inboxID int, key, value string) {
	stmt, err := ss.db.Prepare("insert into config (inbox_id, key, value) values (?, ?, ?) on conflict (inbox_id, key) do update set value = ?")
	if err != nil {
		panic("store: ConfigSet prepare stmt failed: " + err.Error())
	}
	_, err = stmt.Exec(inboxID, key, value, value)
	if err != nil {
		panic("store: ConfigSet exec stmt failed: " + err.Error())
	}
}

func (ss *SQLStore) ConfigGetString(inboxID int, key string) string {
	q, err := ss.db.Prepare("select value from config where inbox_id = ? and key = ?")
	if err != nil {
		panic("store: ConfigGet prepare stmt failed: " + err.Error())
	}

	var value string
	err = q.QueryRow(inboxID, key).Scan(&value)
	if err != nil && err != sql.ErrNoRows {
		panic("store: ConfigGet query failed: " + err.Error())
	}

	return value
}

func (ss *SQLStore) ConfigIsSet(inboxID int, key string) bool {
	q, err := ss.db.Prepare("select 1 from config where inbox_id = ? and key = ?")
	if err != nil {
		panic("store: ConfigIsSet prepare stmt failed: " + err.Error())
	}

	var count uint8
	err = q.QueryRow(inboxID, key).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		panic("store: ConfigIsSet query failed: " + err.Error())
	}

	return count == 1
}

type InboxConfig struct {
	inboxID int
	store   *SQLStore
}

func NewInboxConfig(inboxID int, store *SQLStore) *InboxConfig {
	return &InboxConfig{inboxID: inboxID, store: store}
}

func (ic *InboxConfig) Set(key, value string) {
	ic.store.ConfigSet(ic.inboxID, key, value)
}

func (ic *InboxConfig) GetString(key string) string {
	return ic.store.ConfigGetString(ic.inboxID, key)
}

func (ic *InboxConfig) IsSet(key string) bool {
	return ic.store.ConfigIsSet(ic.inboxID, key)
}
