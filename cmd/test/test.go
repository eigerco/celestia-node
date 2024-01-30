package main

import (
	"encoding/json"
	"log"
	"reflect"
)

func main() {

	var v interface{}
	if err := json.Unmarshal([]byte("{\"data\": true}"), &v); err != nil {
		log.Fatal(err)
	}

	log.Printf("%v\n", reflect.TypeOf(v))
}
