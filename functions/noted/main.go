package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/apex/go-apex"
	"github.com/apex/go-apex/proxy"

	"github.com/jrnt30/noted-apex/pkg/noted"
)

func main() {
	ls := NewDynamoLinkSaver()

	srv := http.NewServeMux()
	srv.HandleFunc("/", http.NotFound)
	srv.HandleFunc("/links", func(w http.ResponseWriter, r *http.Request) {
		if ls.Enabled() {
			http.Error(w, "error encountered generating a link saver", http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodPost:
			postLink(&ls, w, r)
		default:
			http.NotFound(w, r)
		}
	})

	apex.Handle(proxy.Serve(srv))
}

func postLink(ls noted.LinkProcessor, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	link := &noted.Link{}
	err = json.Unmarshal(body, link)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		http.Error(w, fmt.Sprintf("Unable to unmarshall respose properly: %v", body), http.StatusBadRequest)
		return
	}

	err = ls.ProcessLink(link)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		http.Error(w, "Error persisting link", http.StatusInternalServerError)
		return
	}

	res, _ := json.Marshal(link)
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}
