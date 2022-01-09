package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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
	http.HandleFunc("/", showScreen)
	http.HandleFunc("/add_memo", addMemo)
	http.ListenAndServe(":8080", nil)
}

func showScreen(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, htmlStr)
}

type Memo struct {
	ID        string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var memos map[string]Memo

func addMemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, htmlStr)
}
