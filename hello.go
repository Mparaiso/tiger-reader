package hello

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

type Greeting struct {
	Author  string
	Content string
	Date    time.Time
}

func guestbookKey(c context.Context) *datastore.Key {
	return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/sign", sign)
	http.HandleFunc("/storage", storageController)
	http.HandleFunc("/env", envHandler)
}
func envHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, os.Getenv("TEST"))

}

// root is the root route handler
func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(10)
	greetings := make([]Greeting, 0, 10)
	if _, err := q.GetAll(c, &greetings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := guestbookTemplate.Execute(w, greetings); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var guestbookTemplate = template.Must(template.New("book").Parse(`
  <html>
    <head>
      <title>Go Guestbook</title>
    </head>
    <body>
      {{range .}}
        {{with .Author}}
          <p><b>{{.}}</b> wrote: </p>
        {{else}}
          <p>An anonymous person wrote:</p>
        {{end}}
        <pre>{{.Content}}</pre>
      {{end}}
      <form action="/sign" method="POST">
        <div><textarea id="" name="content" cols="30" rows="10"></textarea></div>
        <div><input type="submit" value="Sign Guestbook"></div>
      </form>
    </body>
  </html>
`))

func sign(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := Greeting{
		Content: r.FormValue("content"),
		Date:    time.Now(),
	}
	if u := user.Current(c); u != nil {
		g.Author = u.String()
	}
	key := datastore.NewIncompleteKey(c, "Greeting", guestbookKey(c))
	_, err := datastore.Put(c, key, &g)
	err = signTemplate.Execute(w, r.FormValue("content"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const guestBookForm = `
  <html>
    <body>
      <form action="/sign" method="POST">
        <div><textarea id="" name="content" cols="30" rows="10"></textarea></div>
        <div><input type="submit" value="Sign Guestbook"></div>
      </form>
    </body>
  </html>
`

var signTemplate = template.Must(template.New("sign").Parse(signTemplateHTML))

const signTemplateHTML = `
  <html>
    <body>
      <p></p>
      <pre>{{.}}</pre>
    </body>
  </html>
`

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
