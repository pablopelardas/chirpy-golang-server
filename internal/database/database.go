package database

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
	RevokedTokens map[string]bool `json:"revokedTokens"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
	Author int `json:"author_id"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
	IsChirpyRed bool `json:"is_chirpy_red"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func DeleteDB(path string) (bool, error){
	err := os.Remove(path)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *DB) CreateChirp(body string, author_id int ) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:   id,
		Body: body,
		Author: author_id,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}
func (db *DB) DeleteChirp(id, author_id int) (error){
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	chirp, ok := dbStructure.Chirps[id]
	if !ok {
		return errors.New("chirp not found")
	}
	if chirp.Author != author_id {
		return errors.New("unauthorized")
	}
	delete(dbStructure.Chirps, id)
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil

}
func (db *DB) CreateUser(email, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id := len(dbStructure.Users) + 1
	user := User{
		ID:   id,
		Email:email,
		Password: password,
		IsChirpyRed: false,
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpdateUser(id string, email, password string) (User, error){
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user, err := db.GetUser(id)
	if err != nil {
		return User{}, err
	}
	user.Password = password
	user.Email = email
	dbStructure.Users[user.ID] = user
	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return user, nil

}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}
func (db *DB) GetUsers() ([]User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(dbStructure.Users))
	for _, user := range dbStructure.Users {
		users = append(users, User{
			ID: user.ID,
			Email: user.Email,
		})
	}

	return users, nil
}

func (db *DB) GetChirp(id string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	intId, err := strconv.Atoi(id)
	if err != nil {
		return Chirp{}, err
	}
	chirp, ok := dbStructure.Chirps[intId]
	if !ok {
		return Chirp{}, errors.New("chirp not found")
	}

	return chirp, nil
}
func (db *DB) GetUser(id string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	intId, err := strconv.Atoi(id)
	if err != nil {
		return User{}, err
	}
	user, ok := dbStructure.Users[intId]
	if !ok {
		return User{}, errors.New("chirp not found")
	}

	return User{
		ID: user.ID,
		Email: user.Email,
	}, nil
}

func (db *DB) GetUserByEmail(email string) (User, error){
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")

}

func (db *DB) UpgradeUserToChirpyRed(id int) ( error){
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	user, ok := dbStructure.Users[id]
	if !ok {
		return errors.New("user not found")
	}
	user.IsChirpyRed = true
	dbStructure.Users[id] = user
	err = db.writeDB(dbStructure)
	if err != nil {
		return  err
	}
	return nil
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
		RevokedTokens: map[string]bool{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) RevokeToken(token string)(bool, error){
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}
	dbStructure.RevokedTokens[token] = true
	err = db.writeDB(dbStructure)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *DB) CheckIfTokenRevoked(token string) (bool, error){
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}
	_, ok := dbStructure.RevokedTokens[token]
	return ok, nil

}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}