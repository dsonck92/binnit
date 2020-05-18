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

package main

import (
	"flag"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	cfg "github.com/dsonck92/binnit/config"
	store "github.com/dsonck92/binnit/storage"
	"github.com/dsonck92/binnit/storage/fs"
	"github.com/gorilla/mux"
)

var (
	//Version contains the git hashtag injected by make
	Version = "N/A"
	//BuildTime contains the build timestamp injected by make
	BuildTime = "N/A"
)

var (
	confFile = flag.String("c", "binnit.cfg", "Configuration file for binnit")
	v        = flag.Bool("v", false, "print binnit version and build time")
	logger   *log.Logger
	storage  store.StorageBackend
)

var pConf = cfg.Config{
	ServerPrefix: "http://localhost",
	BindAddr:   "0.0.0.0",
	BindPort:   "8080",
	PasteDir:   "paste",
	TemplDir:   "tpl",
	StaticDir:  "static",
	Storage:    "fs",
	MaxSize:    4096,
	LogFile:    "log/binnit.log",
}

type paste struct {
	Title   string
	Lang    string
	Date    string
	Content []byte
	Raw     bool
}

func setLogger() *log.Logger {
	logger = log.New(os.Stderr, "[binnit]: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	return logger
}

func min(a, b int) int {

	if a > b {
		return b
	}
	return a

}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, pConf.TemplDir+"/index.html")
}

func handleGetStatic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	f := vars["file"]
	if _, err := os.Stat("./" + pConf.StaticDir + "/" + f); err == nil {
		http.ServeFile(w, r, pConf.StaticDir+"/"+f)
	} else if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleGetPaste(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var pasteName, origName string

	origName = vars["id"]
	pasteName = pConf.PasteDir + "/" + origName

	origIP := r.RemoteAddr

	logger.Printf("Received GET from %s for  '%s'\n", origIP, pasteName)

	// if the requested paste exists, we serve it...

	title, date, lang, content, err := storage.Get(pasteName)
	title = html.EscapeString(title)
	date = html.EscapeString(date)
	lang = html.EscapeString(lang)
	if err == nil {
		p := paste{Title: title, Lang: lang, Date: date, Content: content, Raw: false}
		t := template.Must(template.ParseFiles(pConf.TemplDir + "/paste.gohtml"))
		errT := t.Execute(w, p)
		if errT != nil {
			panic(errT)
		}
	} // otherwise, we give say we didn't find it
	fmt.Fprintf(w, "%v\n", err)
	return
}

func handleGetRawPaste(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var pasteName, origName string
	origName = vars["id"]
	pasteName = pConf.PasteDir + "/" + origName
	origIP := r.RemoteAddr
	logger.Printf("Received GET from %s for  '%s'\n", origIP, origName)
	// if the requested paste exists, we serve it...
	title, date, lang, content, err := storage.Get(pasteName)
	title = html.EscapeString(title)
	date = html.EscapeString(date)
	lang = html.EscapeString(lang)
	if err == nil {
		p := paste{Title: title, Lang: lang, Date: date, Content: content, Raw: true}
		t := template.Must(template.ParseFiles(pConf.TemplDir + "/paste.gohtml"))
		errT := t.Execute(w, p)
		if errT != nil {
			panic(errT)
		}
	}
	// otherwise, we give say we didn't find it
	fmt.Fprintf(w, "%v\n", err)
	return
}

func handlePutPaste(w http.ResponseWriter, r *http.Request) {
	err1 := r.ParseForm()
	err2 := r.ParseMultipartForm(int64(2 * pConf.MaxSize))

	if err1 != nil && err2 != nil {
		// Invalid POST -- let's serve the default file
		http.ServeFile(w, r, pConf.TemplDir+"/index.html")
	} else {
		reqBody := r.PostForm

		origIP := r.RemoteAddr

		logger.Printf("Received new POST from %s\n", origIP)

		// get title, body, and time
		title := reqBody.Get("title")
		date := time.Now().String()
		lang := reqBody.Get("lang")
		content := []byte(reqBody.Get("paste"))

		content = content[0:min(len(content), int(pConf.MaxSize))]

		ID, err := storage.Put(title, date, lang, content, pConf.PasteDir)

		logger.Printf("   ID: %s; err: %v\n", ID, err)

		if err == nil {
			prefix := pConf.ServerPrefix
			if show := reqBody.Get("show"); show != "1" {
				fmt.Fprintf(w, "%s/%s\n", prefix, ID)
				return
			}
			fmt.Fprintf(w, "<html><body>Link: <a href='%s/%s'>%s/%s</a></body></html>",
				prefix, ID, prefix, ID)
			return

		}
		fmt.Fprintf(w, "%s\n", err)

	}
}

func loadStorage(name, options string) store.StorageBackend {
	var st store.StorageBackend
	var err error
	switch name {
	case "fs":
		st, err = fs.NewStorage(options)
	}
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return st

}

func init() {
	flag.Parse()
	cfg.ParseConfig(*confFile, &pConf)
	setLogger()
	storage = loadStorage(pConf.Storage, pConf.PasteDir)

	logger.Println("Binnit version " + Version + " -- Starting ")
	logger.Printf("  + Config file: %s\n", *confFile)
	logger.Printf("  + Serving pastes on: %s\n", pConf.ServerPrefix)
	logger.Printf("  + listening on: %s:%s\n", pConf.BindAddr, pConf.BindPort)
	logger.Printf("  + paste_dir: %s\n", pConf.PasteDir)
	logger.Printf("  + log_file: %s\n", pConf.LogFile)
	logger.Printf("  + static_dir: %s\n", pConf.StaticDir)
	logger.Printf("  + storage: %s\n", pConf.Storage)
	logger.Printf("  + templ_dir: %s\n", pConf.TemplDir)
	logger.Printf("  + max_size: %d\n", pConf.MaxSize)
}

func main() {

	if *v {
		fmt.Println(Version, BuildTime)
		os.Exit(0)
	}

	// FIXME: create paste_dir if it does not exist

	var r = mux.NewRouter()
	r.StrictSlash(true)

	r.PathPrefix("/favicon.ico").Handler(http.NotFoundHandler()).Methods("GET")
	r.PathPrefix("/robots.txt").Handler(http.NotFoundHandler()).Methods("GET")

	static := "/" + pConf.StaticDir + "/{file}"
	r.HandleFunc("/", handleIndex).Methods("GET")
	r.HandleFunc("/", handlePutPaste).Methods("POST")
	r.HandleFunc("/{id}", handleGetPaste).Methods("GET")
	r.HandleFunc("/{id}/raw", handleGetRawPaste).Methods("GET")
	r.HandleFunc(static, handleGetStatic).Methods("GET")

	logger.Fatal(http.ListenAndServe(pConf.BindAddr+":"+pConf.BindPort, r))
}
