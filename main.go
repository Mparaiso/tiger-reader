//    TIGER READER version 0.1
//    TIGER READER  is a rss reader server app build in Go
//    Copyright (C) 2015  mparaiso <mparaiso@online.fr>
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.

//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.

//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/interactiv/monorail"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/unrolled/render.v1"
	"gopkg.in/yaml.v1"
)

const NOTE_COLLECTION = "TIGER_READER_notes"

type Configuration struct {
	TIGER_READER_MONGODB_URL string `yaml:"TIGER_READER_MONGODB_URL"`
	TIGER_READER_MONGODB_DB  string `yaml:"TIGER_READER_MONGODB_DB"`
}

type Note struct {
	Content   string
	CreatedAt time.Time
	Version   int
}

type NoteMessage struct {
	Note Note
}

type NoteCollectionMessage struct {
	Notes []Note
}

type Message struct {
	Message string
}

// GetConfiguration return the config for the app
func GetConfiguration() *Configuration {

	config := &Configuration{}
	filename := os.Getenv("CONFIG_FILE")
	if filename == "" {
		filename = "./.secret.yaml"
	}
	secretFile, err := os.Open(filename)
	FatalOnError(err)
	secretFileInfo, err := ioutil.ReadAll(secretFile)
	FatalOnError(err)
	err = yaml.Unmarshal(secretFileInfo, &config)
	FatalOnError(err)
	return config
}

// GetApplication returns a monorail application
func GetApplication() *monorail.Monorail {
	var (
		app     *monorail.Monorail
		rndr    *render.Render
		session *mgo.Session
		config  *Configuration
		db      *mgo.Database
		err     error
	)
	config = GetConfiguration()
	app = monorail.New()
	rndr = render.New(render.Options{Extensions: []string{".html"}})
	session, err = mgo.Dial(config.TIGER_READER_MONGODB_URL)
	db = session.DB(config.TIGER_READER_MONGODB_DB)
	FatalOnError(err)
	app.Injector().Register(rndr)
	app.Injector().Register(db)
	app.Get("/", HomeController)
	notesController := monorail.NewControllerCollection()
	notesController.Get("/", NoteIndex)
	notesController.Post("/", NoteCreate)
	app.Mount("/notes", notesController)
	return app
}

// HomeController handles requests for homepage
func HomeController(writer http.ResponseWriter, render *render.Render) {

	render.HTML(writer, http.StatusOK, "home", struct{ Title string }{Title: "TIGER READER"})
}

// NoteIndex handles the listing of notes
func NoteIndex(w http.ResponseWriter, r *http.Request, render *render.Render, db *mgo.Database) {
	noteCollectionMessage := NoteCollectionMessage{}
	err := db.C(NOTE_COLLECTION).Find(bson.M{}).All(&noteCollectionMessage.Notes)
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, Message{Message: "Database Error"})
		return
	} else {
		render.JSON(w, http.StatusOK, noteCollectionMessage)
	}
}

// NoteCreate persists a new note in the db
func NoteCreate(w http.ResponseWriter, r *http.Request, render *render.Render, db *mgo.Database) {
	noteMessage := &NoteMessage{}
	err := json.NewDecoder(r.Body).Decode(noteMessage)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, Message{Message: "Invalid JSON"})
		return
	}
	noteMessage.Note.CreatedAt = time.Now()
	err = db.C(NOTE_COLLECTION).Insert(noteMessage.Note)
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, Message{Message: "Cannot persist Note"})
		return
	} else {
		render.JSON(w, http.StatusCreated, Message{Message: "Created!"})
	}

}

// FatalOnError logs a fatal error
func FatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// init appengine app
// +build !testing
func init() {
	app := GetApplication()
	http.Handle("/", app)
}

/*
func loginHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	fmt.Fprintf(w, "Hello, %v!", u)
}
*/
