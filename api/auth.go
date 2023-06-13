package api

import (
	"sync"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/who/db"
	"github.com/shadiestgoat/who/snownode"
)

type atomicMap struct {
	m map[string]bool
	sync.Mutex
}

// returns false if failed to insert
func (a *atomicMap) Insert(uname string) bool {
	a.Lock()
	defer a.Unlock()

	if a.m[uname] {
		return false
	}

	a.m[uname] = true

	return true
}

func (a *atomicMap) Free(uname string) {
	a.Lock()
	defer a.Unlock()
	
	delete(a.m, uname)
}

var unameLock = &atomicMap{
	m:     map[string]bool{},
	Mutex: sync.Mutex{},
}

func NewUser(uname, password string) (id, token string, err error) {
	if err := cleanString(&uname, 7, 33, "username"); err != nil {
		return "", "", err
	}
	if err := cleanString(&password, 7, 33, "password"); err != nil {
		return "", "", err
	}
	
	if db.Exists(`ppl`, `username = $1`, uname) || !unameLock.Insert(uname) {
		return "", "", ErrUniqueUname
	}

	passwordHash, err := generateFromPassword(password)
	
	if log.ErrorIfErr(err, "generating hash") {
		return "", "", ErrServerErr
	}

	id = snownode.Generate()

	for {
		token = randGoodString(128)
		if !db.Exists(`ppl`, `token = $1`, token) && unameLock.Insert(token) {
			break
		}
	}

	_, err = db.InsertOne(`ppl`, []string{`id`, `token`, `username`, `password`}, id, token, uname, passwordHash)
	if err != nil {
		return "", "", ErrDBHandle(err)
	}
	
	unameLock.Free(uname)
	unameLock.Free(token)

	return id, token, nil
}

func EditPassword(id string, oldPassword string, newPassword string) (string, error) {
	password := ""

	err := db.QueryRowID(`SELECT password FROM ppl WHERE id = $1`, id, &password)
	
	if err != nil {
		return "", ErrDBHandle(err)
	}

	match, err := comparePasswordAndHash(password, oldPassword)

	if err != nil {
		return "", ErrServerErr
	}

	if !match {
		return "", ErrNoAuth
	}

	hash, err := generateFromPassword(newPassword)

	if log.ErrorIfErr(err, "generating hash") {
		return "", ErrServerErr
	}

	token := ""

	for {
		token = randGoodString(128)
		if !db.Exists(`ppl`, `token = $1`, token) && unameLock.Insert(token) {
			break
		}
	}
	
	_, err = db.Exec(`UPDATE ppl SET password = $1, token = $2 WHERE id = $3`, hash, token, id)

	if err != nil {
		return "", ErrDBHandle(err)
	}

	unameLock.Free(token)

	return token, nil
}

func Exchange(uname, password string) (id, token string, err error) {
	dbPass := ""

	err = db.QueryRowID(`SELECT password, id, token FROM ppl WHERE username = $1`, uname, &dbPass, &id, &token)
	if err != nil {
		return "", "", ErrDBHandle(err)
	}

	match, err := comparePasswordAndHash(password, dbPass)
	if log.ErrorIfErr(err, "comparePasswordAndHash") {
		return "", "", ErrServerErr
	}
	if !match {
		return "", "", ErrNoAuth
	}

	return
}

func AuthTokenToID(token string) (string, error) {
	id := ""

	err := db.QueryRowID(`SELECT id FROM ppl WHERE token = $1`, token, &id)

	if err != nil {
		return "", ErrDBHandle(err)
	}

	return id, nil
}
