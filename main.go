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

	data, err := os.ReadFile("index.html")
	if err != nil {
		//ファイルが見つからない場合
		log.Fatal(err)
	}

	htmlStr = string(data)

	//http://localhost:8080/
	http.HandleFunc("/", hanlder)
	http.ListenAndServe(":8080", nil)
}

func hanlder(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, htmlStr)
}
