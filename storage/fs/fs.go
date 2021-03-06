/*
 *  This program is free software: you can redistribute it and/or
 *  modify it under the terms of the GNU Affero General Public License as
 *  published by the Free Software Foundation, either version 3 of the
 *  License, or (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 *  General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public
 *  License along with this program.  If not, see
 *  <http://www.gnu.org/licenses/>.
 *
 *  (c) Vincenzo "KatolaZ" Nicosia 2017 -- <katolaz@freaknet.org>
 *
 *
 *  This file is part of "binnit", a minimal no-fuss pastebin-like
 *  server written in golang
 *
 */

/*
 *
 * Write/Read functions for FS-based paste storage
 *
 */

package fs

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

//Paste is a struct containing a paste
type Paste struct {
	title   string
	date    time.Time
	lang    string
	content []byte
	dir     string
}

//NewStorage creates a file system storage
func NewStorage(dirname string) (*Paste, error) {
	me := Paste{}
	return &me, nil
}

//Put will write a paste to the file system
func (p Paste) Put(title, date, lang string, content []byte, destDir string) (string, error) {

	safename, errN := p.makePasteName(title, date, lang, content, destDir)
	if errN != nil {
		return "", errN
	}

	meta := fmt.Sprintf("Title: %s\nDate: %s\nLanguage: %s\n", title, date, lang)
	if errM := ioutil.WriteFile(destDir+"/"+safename+".meta", []byte(meta), 0644); errM != nil {
		return "", errM
	}

	if errC := ioutil.WriteFile(destDir+"/"+safename, content, 0644); errC != nil {
		return "", errC
	}

	return safename, nil

}

func (p Paste) getPasteMetadata(URI string) (title string, date string, lang string, err error) {
	meta, err := ioutil.ReadFile(URI + ".meta")
	lines := bytes.Split(meta, []byte("\n"))
	for _, l := range lines {
		switch mod := string(bytes.TrimSpace(bytes.Split(l, []byte(":"))[0])); mod {
		case "Title":
			title = string(bytes.TrimSpace(bytes.Split(l, []byte(":"))[1]))
		case "Date":
			date = string(bytes.TrimSpace(bytes.Split(l, []byte(":"))[1]))
		case "Language":
			lang = string(bytes.TrimSpace(bytes.Split(l, []byte(":"))[1]))

		}
	}
	return
}

func (p Paste) makePasteName(title, date, lang string, content []byte, destDir string) (string, error) {
	var pasteName string
	h := sha256.New()
	h.Write([]byte(title))
	h.Write([]byte(date))
	h.Write([]byte(lang))
	h.Write(content)
	pasteHash := fmt.Sprintf("%x", h.Sum(nil))
	pasteDir := destDir + "/"

	for i := 0; i < len(pasteHash)-16; i++ {
		pasteName = pasteHash[i : i+16]
		if _, err := os.Stat(pasteDir + pasteName); os.IsNotExist(err) {
			if _, errC := os.Create(pasteDir + pasteName); errC != nil {
				return "", errC
			}
			break
		}
	}
	return pasteName, nil
}

//Get will get a paste from the filesystem
func (p Paste) Get(URI string) (title string, date string, lang string, content []byte, err error) {
	title, date, lang, err = p.getPasteMetadata(URI)
	content, err = ioutil.ReadFile(URI)
	return
}

// Flush the storage
func (p Paste) Flush() error {
	fmt.Println("Flush called")
	return nil
}
