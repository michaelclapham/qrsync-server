package main

import "os"

func main() {
	if os.Args[1] == "-ts" {
		convertToTS()
	} else {
		app := App{}
		app.Init()
	}
}
