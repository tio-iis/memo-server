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

	//http://localhost:8080/
	http.HandleFunc("/", showHTML)
	http.HandleFunc("/add_memo", addMemo)
	http.HandleFunc("/list_memos", listMemos)
	http.ListenAndServe(":8080", nil)
}

// curl localhost:8080/
func showHTML(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("index.html")
	if err != nil {
		//ファイルが見つからない場合
		log.Fatal(err)
	}

	htmlStr = string(data)

	fmt.Fprintln(w, htmlStr)
}

//構造体自体の定義は*を付けることはできません。
type Memo struct {
	ID        int
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

//Memo構造体をポインタ型として定義しています。
var memos map[int]*Memo = map[int]*Memo{}

//メモを登録する。
//curl -X POST -H "Content-Type: application/json" -d '{"ID":"1111","Title":"mytitle","Body":"mybody","CreatedAt":"2022-01-01T10:00:00+09:00","UpdatedAt":"2022-01-01T11:00:00+09:00"}' localhost:8080/add_memo
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

	//メモをmemosに保存する。
	//Memo.IDをキーにしているので、IDが同じメモは上書きされる。
	memos[m.ID] = m

	//HTTP Response は空にするので、nilを指定する。
	//len()は配列やマップなどの長さを出力することができる関数です。
	fmt.Fprintln(w, len(memos))
}

//保存してあるメモの一覧をJSONで出力する。
//curl localhost:8080/list_memos
func listMemos(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(memos)
	if err != nil {
		fmt.Fprintln(w, "error:"+err.Error())
		return

	}

	fmt.Fprintln(w, string(b))
}
