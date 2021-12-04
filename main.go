package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var htmlStr string

func main() {
	fmt.Println("start")

	data, err := os.ReadFile("indexxx.html")
	if err != nil {
		//ファイルが見つからない場合
		log.Fatal(err)
	}

	htmlStr = string(data)

	http.HandleFunc("/", hanlder)
	http.ListenAndServe(":8080", nil)
}

func hanlder(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, htmlStr)
}
