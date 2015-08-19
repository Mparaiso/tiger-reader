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

package tigerreader

import (
	"encoding/json"

	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/interactiv/monorail"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"gopkg.in/unrolled/render.v1"
)

const NOTE_KIND = "Note"
const NOTE_KEY = "ID"

type NoteRepository struct {
	context context.Context
}

func NewNoteRepository(context context.Context) *NoteRepository {
	return &NoteRepository{context: context}
}

func (nr NoteRepository) Insert(note *Note) (int64, error) {
	key, err := datastore.Put(nr.context, datastore.NewIncompleteKey(nr.context, NOTE_KIND, nil), note)
	return key.IntID(), err
}
func (nr NoteRepository) Find(id int64) (*Note, error) {
	note := &Note{}
	err := datastore.Get(nr.context, datastore.NewKey(nr.context, NOTE_KIND, "", id, nil), note)
	return note, err

}
func (nr NoteRepository) FindAll() ([]*Note, error) {
	notes := []*Note{}
	query := datastore.NewQuery(NOTE_KIND)
	_, err := query.GetAll(nr.context, &notes)
	return notes, err
}
func (nr NoteRepository) Update(id int64, note *Note) error {
	_, err := nr.Find(id)
	if err != nil {
		return err
	}
	_, err = datastore.Put(nr.context, datastore.NewKey(nr.context, NOTE_KIND, "", id, nil), note)
	return err
}
func (nr NoteRepository) Delete(id int64) error {
	return datastore.Delete(nr.context, datastore.NewKey(nr.context, NOTE_KIND, "", id, nil))
}

type Note struct {
	ID        int64
	Content   string
	CreatedAt time.Time
	Version   int
}

type NoteMessage struct {
	Note *Note
}

type NoteCollectionMessage struct {
	Notes []*Note
}

type Message struct {
	Message string
	Link    string
}

// GetApplication returns a monorail application
func GetApplication() *monorail.Monorail {
	var (
		app  *monorail.Monorail
		rndr *render.Render
	)
	app = monorail.New()
	rndr = render.New(render.Options{Extensions: []string{".html"}})
	app.Injector().Register(rndr)
	app.Use("/", AppEngineContextMiddleWare)
	app.Get("/", HomeController)
	notesController := monorail.NewControllerCollection()
	notesController.Use("/", NotesMiddleware)
	notesController.Get("/", NoteIndex)
	notesController.Post("/", NoteCreate)
	notesController.Get("/:noteId", NoteShow)
	notesController.Put("/:noteId", NoteUpdate)
	notesController.Delete("/:noteId", NoteDelete)
	app.Mount("/notes", notesController)
	return app
}

func NotesMiddleware(context context.Context, injector *monorail.Injector, next monorail.Next) {
	noteRepository := NewNoteRepository(context)
	injector.Register(noteRepository)
	next()
}

func AppEngineContextMiddleWare(r *http.Request, next monorail.Next, injector *monorail.Injector) {
	c := appengine.NewContext(r)
	injector.Register(c)
	next()
}

// HomeController handles requests for homepage
func HomeController(writer http.ResponseWriter, render *render.Render) {
	render.HTML(writer, http.StatusOK, "home", struct{ Title string }{Title: "TIGER READER"})
}

// NoteIndex handles the listing of notes
func NoteIndex(w http.ResponseWriter, r *http.Request, render *render.Render, context context.Context, noteRepository *NoteRepository) {
	noteCollectionMessage := NoteCollectionMessage{}
	notes, err := noteRepository.FindAll()
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, Message{Message: err.Error()})
		return
	} else {
		noteCollectionMessage.Notes = notes
		render.JSON(w, http.StatusOK, noteCollectionMessage)
	}
}

// NoteCreate persists a new note in the db
func NoteCreate(w http.ResponseWriter, r *http.Request, render *render.Render, context context.Context, noteRepository *NoteRepository) {
	noteMessage := &NoteMessage{}
	err := json.NewDecoder(r.Body).Decode(noteMessage)
	if err != nil {
		render.JSON(w, http.StatusBadRequest, Message{Message: "Invalid JSON"})
		return
	}
	noteMessage.Note.CreatedAt = time.Now()
	id, err := noteRepository.Insert(noteMessage.Note)
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, Message{Message: err.Error()})
		return
	} else {
		log.Debugf(context, "note_id : %#v", id)
		render.JSON(w, http.StatusCreated, Message{Message: "Created!", Link: fmt.Sprintf("/%s/%d", NOTE_KIND, id)})
	}

}

// NoteUpdate handles note show
func NoteShow(noteRepository *NoteRepository, ctx *monorail.Context, w http.ResponseWriter, render *render.Render) {
	id, err := strconv.ParseInt(ctx.RequestVars["noteId"], 10, 64)
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, Message{Message: "Error invalid note id"})
		return
	}
	note, err := noteRepository.Find(id)
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, Message{Message: err.Error()})
		return
	}
	render.JSON(w, http.StatusOK, NoteMessage{Note: note})
}

// NoteUpdate handles note update
func NoteUpdate(noteRepository *NoteRepository, ctx *monorail.Context, w http.ResponseWriter, render *render.Render) {
	noteMessage := &NoteMessage{}
	if err := json.NewDecoder(ctx.Request.Body).Decode(noteMessage); err != nil {
		render.JSON(w, http.StatusBadRequest, Message{Message: err.Error()})
		return
	}
	if id, err := strconv.ParseInt(ctx.RequestVars["node_id"], 10, 64); err != nil {
		render.JSON(w, http.StatusBadRequest, Message{Message: err.Error()})
	} else if err := noteRepository.Update(id, noteMessage.Note); err != nil {
		render.JSON(w, http.StatusNotFound, Message{Message: err.Error()})
	} else {
		render.JSON(w, http.StatusOK, Message{Message: "Updated", Link: fmt.Sprintf("/%s/%d", NOTE_KIND, id)})
	}
}

// NoteDelete handles note deletion
func NoteDelete(noteRepository *NoteRepository, ctx *monorail.Context, w http.ResponseWriter, render *render.Render) {
	if id, err := strconv.ParseInt(ctx.RequestVars["node_id"], 10, 64); err != nil {
		render.JSON(w, http.StatusBadRequest, Message{Message: err.Error()})
	} else if err := noteRepository.Delete(id); err != nil {
		render.JSON(w, http.StatusNotFound, Message{Message: err.Error()})
	} else {
		render.JSON(w, http.StatusOK, Message{Message: "Deleted"})
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
