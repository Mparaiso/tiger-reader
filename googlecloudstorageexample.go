// google cloud storage example
package hello

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine/file"
	"google.golang.org/appengine/log"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/appengine"
)

// local cache of the app's default bucket name
var bucket string = "pipes-1038.appspot.com" //os.Getenv("GCS_BUCKET_NAME")

// demo holds infos needed to run various demo functions
type demo struct {
	c       context.Context
	w       http.ResponseWriter
	ctx     context.Context
	cleanUp []string
	failed  bool
}

func (d *demo) errorf(format string, args ...interface{}) {
	d.failed = true
	log.Errorf(d.c, format, args...)
}

// create file in Google cloud storage
func (d *demo) createFile(filename string) {
	fmt.Fprintf(d.w, "Creating file /%v/%v\n", bucket, filename)
	wc := storage.NewWriter(d.ctx, bucket, filename)
	wc.ContentType = "text/plain"
	wc.Metadata = map[string]string{
		"x-goog-meta-foo": "foo",
	}
	d.cleanUp = append(d.cleanUp, filename)
	if _, err := wc.Write([]byte("abcde\n")); err != nil {
		d.errorf("createFile: unable to write data to bucket %q, file %q, %v", bucket, filename, err)
		return
	}
	if err := wc.Close(); err != nil {
		d.errorf("createFile: unable to close bucket %q, file %q: %v", bucket, filename, err)
		return
	}
}

// readFile reads the named file in Google Cloud Storage.
func (d *demo) readFile(filename string) {
	io.WriteString(d.w, "\nAbbreviated file content (firstline): \n")
	rc, err := storage.NewReader(d.ctx, bucket, filename)
	if err != nil {
		d.errorf("readFile: unable to open file from bucket %q, file %q: %v", bucket, filename, err)
		return
	}
	defer rc.Close()
	slurp, err := ioutil.ReadAll(rc)
	if err != nil {
		d.errorf("readFile: unable to read data from bucket %q, file %q: %v", bucket, filename, err)
		return
	}
	fmt.Fprintf(d.w, "%s\n", bytes.SplitN(slurp, []byte("\n"), 2)[0])
}

// deleteFiles deletes all the temp files from created by this demo
func (d *demo) deleteFiles() {
	io.WriteString(d.w, "\nDeleting files...\n")
	for _, v := range d.cleanUp {
		fmt.Fprintf(d.w, "Deleting file %v\n", v)
		if err := storage.DeleteObject(d.ctx, bucket, v); err != nil {
			d.errorf("deleteFiles: unable to delete bucket %q, file %q: %v", bucket, v, err)
			return
		}
	}
}

// entry point
func storageController(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if bucket == "" {
		var err error
		if bucket, err = file.DefaultBucketName(c); err != nil || bucket == "" {
			log.Errorf(c, "failed to get default GCS bucket name: %v", err)
			return
		}
	}
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(c, storage.ScopeFullControl),
		},
	}
	ctx := cloud.NewContext(appengine.AppID(c), client)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Demo CGS Application running from Version: %v\n", appengine.VersionID(c))
	fmt.Fprintf(w, "Using bucket name: %v\n\n", bucket)
	d := &demo{c: c, w: w, ctx: ctx}
	n := "demo-testfile-go"
	d.createFile(n)
	d.readFile(n)
	d.deleteFiles()
	if d.failed {
		io.WriteString(w, "\n Demo failed.\n")
	} else {
		io.WriteString(w, "\n Demo Succeeded. \n")
	}

}
