package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-ts" {
		convertToTS()
	} else {
		app := App{}
		app.Init()
		log.Fatal(app.ListenOnPort(4010, false))
	}
}
