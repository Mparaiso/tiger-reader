package main_test

import (
	"testing"

	"github.com/interactiv/expect"
	tiger_reader "github.com/mediactiv/tiger-reader"
)

func TestGetConfiguration(t *testing.T) {
	e := expect.New(t)
	config := tiger_reader.GetConfiguration()
	e.Expect(config.TIGER_READER_MONGODB_URL).Not().ToEqual("")
}
