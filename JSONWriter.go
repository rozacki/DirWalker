package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONWriter struct {
	First bool
}

func (self JSONWriter) Init() error {
	return nil
}

func (self JSONWriter) SetSessionWriter(w http.ResponseWriter) {

}

func (self *JSONWriter) WriteHeader(w http.ResponseWriter, item Item, error int, msg string) error {
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(fmt.Sprintf("{\"error\":%d, \"msg\":\"%s\", \"header\":", error, msg)))
	j, _ := json.Marshal(item)
	w.Write(j)
	w.Write([]byte(",\"items\":["))
	self.First = false

	return nil
}

func (self *JSONWriter) WriteItem(w http.ResponseWriter, item Item) error {
	if self.First {
		w.Write([]byte(","))
	} else {
		self.First = true
	}
	j, _ := json.Marshal(item)
	w.Write([]byte(j))
	return nil
}

func (self JSONWriter) WriteStartItem(w http.ResponseWriter) {

}

//WriteItem(w http.ResponseWriter, item Item)

func (self JSONWriter) WriteEndItem(w http.ResponseWriter) {

}

func (self JSONWriter) WriteFooter(w http.ResponseWriter, item Item) {
	w.Write([]byte("]}"))
}

//debug
func (self JSONWriter) len() int {
	return 0
}
