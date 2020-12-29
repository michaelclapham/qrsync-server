package main

import "os"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-ts" {
		convertToTS()
	} else {
		app := App{}
		app.Init()
	}
}
