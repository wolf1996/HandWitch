package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var errorTemplate = `{
	"error_message": "%s"
}
`

func handler(w http.ResponseWriter, r *http.Request) {
	pathsParts := strings.Split(r.URL.Path, "/")
	if len(pathsParts) < 3 {
		errmsg := fmt.Sprintf("wrong path arguments number %d path should be like /{string}/{int}", len(pathsParts))
		http.Error(w, fmt.Sprintf(errorTemplate, errmsg), http.StatusBadRequest)
		return
	}
	pathsParts = pathsParts[1:]
	responce := make(map[string]interface{})
	responce["string_argument"] = pathsParts[0]
	var err error
	responce["int_argument"], err = strconv.Atoi(pathsParts[1])
	if err != nil {
		errmsg := fmt.Sprintf("can't parse second path argument %s path should be like /{string}/{int}", pathsParts[1])
		http.Error(w, fmt.Sprintf(errorTemplate, errmsg), http.StatusBadRequest)
		return
	}
	for key, val := range r.URL.Query() {
		responce[key] = val
	}

	w.Header().Add("Content-Type", "application/json")
	rsp, err := json.Marshal(responce)
	if err != nil {
		errmsg := fmt.Sprintf("can't build responce %s", err.Error())
		http.Error(w, errmsg, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(rsp)
	log.Printf("Failed to write responce %s", err.Error())
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
