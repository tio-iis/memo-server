package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var htmlStr string

func main() {
	fmt.Println("start")

	//http://localhost:8080/
	http.HandleFunc("/", showHTML)
	http.HandleFunc("/add_memo", addMemo)
	http.HandleFunc("/list_memos", listMemos)
	http.HandleFunc("/delete_memos", deleteMemos)
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

//Go言語におけるメソッドである。
func (m *Memo) Validate() []ErrorMessage {
	//エラーメッセージを格納する string の配列を定義する
	errMsgs := make([]ErrorMessage, 0)

	//メモのタイトルが1文字未満、30文字より長い場合はエラーにする。
	if len([]rune(m.Title)) < 1 || len([]rune(m.Title)) > 30 {
		errMsgs = append(errMsgs, ErrorMessage{
			Code:     "InvalidTitle",
			Messsage: "タイトルの文字数は1文字以上30文字以下にしてください",
		})
	}

	//メモの本文が1文字未満、100文字より長い場合はエラーにする。
	if len([]rune(m.Body)) < 1 || len([]rune(m.Body)) > 100 {
		errMsgs = append(errMsgs, ErrorMessage{
			Code:     "InvalidBody",
			Messsage: "本文の文字数は1文字以上100文字以下にしてください",
		})
	}

	return errMsgs
}

//Goではコンストラクタの代わりに関数を利用して、
//構造体のオブジェクトを生成する。
//構造体を生成する関数の名前は New + 構造体名 にするのが一般的です。
//ただ、今回は使いません。
//func NewMemo() *Memo {
//	return &Memo{}
//}

//Memo構造体をポインタ型として定義しています。
// [1111] => {ID:1111, Title:mytitle .... }
// [222] => {ID:1111, Title:mytitle .... }
// [333] => {ID:1111, Title:mytitle .... }
var memos map[int]*Memo = map[int]*Memo{}

//Go言語では構造体や関数などの名前の先頭を大文字にするか小文字にするかで、
//プログラムの挙動が変わります。
//メモ帳アプリでは基本的に大文字にした方がよいです。
//私は忘れて小文字にしてしまっています。
type ErrorMessage struct {
	Code     string `json:"code"`
	Messsage string `json:"message"`
}

type ErrorResponse struct {
	Errors []ErrorMessage `json:"errors"`
}

//メモを登録する。
//curl -X POST -H "Content-Type: application/json" -d '{"ID":1111,"Title":"mytitle","Body":"mybody","CreatedAt":"2022-01-01T10:00:00+09:00","UpdatedAt":"2022-01-01T11:00:00+09:00"}' localhost:8080/add_memo
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

	errMsgs := m.Validate()
	if len(errMsgs) > 0 {
		// クライアント側から送信されるリクエストに問題があるので、
		// HTTP Status 400 を返す。
		// Go言語では定数としてHTTP Statusが用意されているので、
		// それを利用するのがいいと思います。
		// https://developer.mozilla.org/ja/docs/Web/HTTP/Status/400
		ReponseError(w, http.StatusBadRequest, errMsgs)
		return
	}

	//メモをmemosに保存する。
	//Memo.IDをキーにしているので、IDが同じメモは上書きされる。
	memos[m.ID] = m

	//HTTP Response は空にするので、nilを指定する。
	//len()は配列やマップなどの長さを出力することができる関数です。
	fmt.Fprintln(w, len(memos))
}

func ReponseError(w http.ResponseWriter, httpStatus int, e []ErrorMessage) {
	er := &ErrorResponse{
		Errors: e,
	}

	erJSON, err := json.Marshal(er)
	if err != nil {
		//https://developer.mozilla.org/ja/docs/Web/HTTP/Status/500
		//サーバ側が原因でエラーが発生した場合は、500を利用する。
		w.WriteHeader(http.StatusInternalServerError)

		//JSON形式ではないが、
		//↑の json.Marshal() で現実的にエラーが発生することがないので、
		//簡易的なエラーメッセージにしている。
		fmt.Fprintln(w, "error:"+err.Error())
		return
	}

	w.WriteHeader(httpStatus)
	fmt.Fprintln(w, string(erJSON))
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

//メモを削除する
//curl -X DELETE localhost:8080/delete_memos?id=1111,222222,333
func deleteMemos(w http.ResponseWriter, r *http.Request) {
	//メモが存在しない場合は何もせずに終わる
	if len(memos) == 0 {
		fmt.Fprintln(w, "There is not a memo.")
		return
	}

	id := r.URL.Query().Get("id")
	ids := strings.Split(id, ",")
	for _, id := range ids {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			//ID変換でエラーになったら処理を終わる
			fmt.Fprintln(w, err.Error())
			return
		}

		if _, yes := memos[idInt]; !yes {
			//メモIDが存在しない場合は処理を終わる
			fmt.Fprintln(w, fmt.Sprintf("not exist memo_id = %d", idInt))
			return
		}

		delete(memos, idInt)
	}

	fmt.Fprintln(w, "memo_id = "+id+" is deleted")
}
