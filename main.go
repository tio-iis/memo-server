package main

import (
	"encoding/json"
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

//構造体自体の定義は*を付けることはできません。
type Memo struct {
	ID        string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

//Memo構造体をポインタ型として定義しています。
var memos map[string]*Memo

//curl -X POST -H "Content-Type: application/json" -d '{"ID":"1111","Title":"mytitle","Body":"mybody","CreatedAt":"2022-01-01 10:00:00","UpdatedAt":"2022-01-01 11:00:00"}' localhost:8080/add_memo
func addMemo(w http.ResponseWriter, r *http.Request) {
	//*を付けると、その型をポインタ型として定義できる。
	//ポインタ型の変数を生成するには&を付ける必要がある。
	//var m *Memo = &Memo{}
	m := &Memo{}

	//HTTPリクエストで送信されてきた HTTP Request Body(JSON形式)を
	//Memo構造体にセットしている。
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		fmt.Fprintln(w, "error:"+err.Error())
		return
	}

	//HTTP Response として出力する
	fmt.Fprintln(w, m)
}
