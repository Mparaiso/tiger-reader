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
