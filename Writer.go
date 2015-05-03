package main

import(
	"net/http"
	"fmt"
)

type Item map[string] interface{}

type WriterError struct{
	Err error
	Message string
}

func (self WriterError) Error() string{
	return fmt.Sprintf(self.Err.Error())
}

type Writer interface{

	Init() error

	SetSessionWriter(w http.ResponseWriter)

	WriteHeader(w http.ResponseWriter, item Item,error int,msg string ) error

	WriteItem(w http.ResponseWriter, item Item) error

	WriteStartItem(w http.ResponseWriter)

	//WriteItem(w http.ResponseWriter, item Item)

	WriteEndItem(w http.ResponseWriter)

	WriteFooter(w http.ResponseWriter, item Item)


	//debug
	len() int

}
