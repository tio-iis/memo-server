package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var htmlStr string

func main() {
	log.SetFlags(0)
	InfoLog("start")

	//http://localhost:8080/
	http.HandleFunc("/", showHTML)
	http.HandleFunc("/add_memo", addMemo)
	http.HandleFunc("/list_memos", listMemos)
	http.HandleFunc("/delete_memos", deleteMemos)
	http.ListenAndServe(":8080", nil)
}

// curl localhost:8080/
func showHTML(w http.ResponseWriter, r *http.Request) {
	OutputAccessLog(r.URL)

	data, err := os.ReadFile("index.html")
	if err != nil {
		RespondInternalServerError(w, err.Error())
		return
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
func (m *Memo) Validate() []*ErrorMessage {
	//エラーメッセージを格納する string の配列を定義する
	errMsgs := make([]*ErrorMessage, 0)

	titleLength := len([]rune(m.Title))

	//メモのタイトルが1文字未満、30文字より長い場合はエラーにする。
	if titleLength < 1 || titleLength > 30 {
		errMsgs = append(errMsgs, NewErrorMessage(
			"InvalidTitle",
			"タイトルの文字数は1文字以上30文字以下にしてください",
		),
		)
		WarningLog(fmt.Sprintf("title length is invalid, title = %s, length = %d)", m.Title, titleLength))
	}

	bodyLength := len([]rune(m.Body))

	//メモの本文が1文字未満、100文字より長い場合はエラーにする。
	if bodyLength < 1 || bodyLength > 100 {
		errMsgs = append(errMsgs, NewErrorMessage(
			"InvalidBody",
			"本文の文字数は1文字以上100文字以下にしてください",
		),
		)
		WarningLog(fmt.Sprintf("body length is invalid, body = %s, length = %d)", m.Body, bodyLength))
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

func NewErrorMessage(code, message string) *ErrorMessage {
	return &ErrorMessage{
		Code:     code,
		Messsage: message,
	}
}

//500エラー専用の構造体を作成する
func NewErrorMessageForInternalServerError() *ErrorMessage {
	return NewErrorMessage("InternalServerError", "内部エラーが発生しました。")
}

//404エラー専用の構造体を作成する
//HTTP Status 404 は操作対象のデータが存在しない場合に利用する。
func NewErrorMessageForNotFound() *ErrorMessage {
	return NewErrorMessage("NotFound", "対象のメモはありません")
}

type ErrorResponse struct {
	Errors []*ErrorMessage `json:"errors"`
}

func NewErrorResponse(em []*ErrorMessage) *ErrorResponse {
	return &ErrorResponse{
		Errors: em,
	}
}

//メモを登録する。
//curl -X POST -H "Content-Type: application/json" -d '{"ID":1111,"Title":"mytitle","Body":"mybody","CreatedAt":"2022-01-01T10:00:00+09:00","UpdatedAt":"2022-01-01T11:00:00+09:00"}' localhost:8080/add_memo
func addMemo(w http.ResponseWriter, r *http.Request) {
	OutputAccessLog(r.URL)

	//*を付けると、その型をポインタ型として定義できる。
	//ポインタ型の変数を生成するには&を付ける必要がある。
	//var m *Memo = &Memo{}
	m := &Memo{}

	//HTTPリクエストで送信されてきた HTTP Request Body(JSON形式)を
	//Memo構造体にセットしている。
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		RespondInternalServerError(w, err.Error())
		return
	}

	errMsgs := m.Validate()
	if len(errMsgs) > 0 {
		// クライアント側から送信されるリクエストに問題があるので、
		// HTTP Status 400 を返す。
		// Go言語では定数としてHTTP Statusが用意されているので、
		// それを利用するのがいいと思います。
		// https://developer.mozilla.org/ja/docs/Web/HTTP/Status/400
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

	//メモをmemosに保存する。
	//Memo.IDをキーにしているので、IDが同じメモは上書きされる。
	memos[m.ID] = m

	//HTTP Response は空にするので、nilを指定する。
	//len()は配列やマップなどの長さを出力することができる関数です。
	fmt.Fprintln(w, len(memos))
}

func RespondInternalServerError(w http.ResponseWriter, errorLogMessage string) {
	ErrorLog(errorLogMessage)
	RespondError(
		w,
		http.StatusInternalServerError,
		[]*ErrorMessage{
			NewErrorMessageForInternalServerError(),
		},
	)
}

func RespondNotFoundError(w http.ResponseWriter) {
	RespondError(
		w,
		http.StatusNotFound,
		[]*ErrorMessage{
			NewErrorMessageForNotFound(),
		},
	)
}

func RespondError(w http.ResponseWriter, httpStatus int, e []*ErrorMessage) {
	er := NewErrorResponse(e)

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
	OutputAccessLog(r.URL)

	b, err := json.Marshal(memos)
	if err != nil {
		RespondInternalServerError(w, err.Error())
		return

	}

	fmt.Fprintln(w, string(b))
}

//メモを削除する
//curl -X DELETE localhost:8080/delete_memos?id=1111
func deleteMemos(w http.ResponseWriter, r *http.Request) {
	OutputAccessLog(r.URL)

	//登録済みのメモが存在しない場合は何もせずに終わる
	if len(memos) == 0 {
		WarningLog(fmt.Sprintf("memo is empty, length = %d", len(memos)))
		RespondNotFoundError(w)
		return
	}

	id := r.URL.Query().Get("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		//string型のIDをintに変換できない場合は処理を終わる
		WarningLog(fmt.Sprintf("memo id format is invalid, id = %s", id))
		errMsgs := []*ErrorMessage{
			NewErrorMessage(
				"MEMO_ID_FORMAT_IS_INVALID",
				"対象のメモは存在しません。",
			),
		}
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

	//TODO: メモがなかった場合にエラーハンドリングする
	delete(memos, idInt)

	fmt.Fprintln(w, "memo_id = "+id+" is deleted")
}

type Log struct {
	//Go言語の"埋め込み（Embed）"を利用することで、
	//Datetime, Level の定義を重複せずにLog, AccesssLogに定義することができる。
	*BaseLog
	Message string `json:"message"`
}

func NewLogInfo(message string) *Log {
	return &Log{
		BaseLog: NewBaselogInfo(),
		Message: message,
	}
}

func NewLogError(message string) *Log {
	return &Log{
		BaseLog: NewBaselogError(),
		Message: message,
	}
}

func NewLogWarning(message string) *Log {
	return &Log{
		BaseLog: NewBaselogWarning(),
		Message: message,
	}
}

type AccessLog struct {
	*BaseLog
	URL string `json:"url"`
}

func NewAccessLog(u *url.URL) *AccessLog {
	bl := NewBaselogInfo()
	bl.Kind = "access_log"
	return &AccessLog{
		BaseLog: bl,
		URL:     u.String(),
	}
}

type BaseLog struct {
	Datetime string `json:"date_time"`
	Level    string `json:"level"`
	//ログの種類
	Kind string `json:"kind"`
}

func NewBaselog(level string) *BaseLog {
	return &BaseLog{
		Datetime: time.Now().Format("2006-01-02 15:04:05"),
		Level:    level,
		Kind:     "log",
	}
}

func NewBaselogInfo() *BaseLog {
	return NewBaselog("INFO")
}

func NewBaselogError() *BaseLog {
	return NewBaselog("ERROR")
}

func NewBaselogWarning() *BaseLog {
	return NewBaselog("WARNING")
}

//エラーログ
//ログレベル = Error
//用途 = サーバ側でエラーが発生したときに利用する（サーバが悪い）
func ErrorLog(message string) {
	j, _ := json.Marshal(NewLogError(message))
	log.Print(string(j))
}

//警告ログ
//ログレベル = Warning
//用途 = HTTPリクエストに問題があったときに出力するログ（バリデーション）
func WarningLog(message string) {
	j, _ := json.Marshal(NewLogWarning(message))
	log.Print(string(j))
}

//インフォログ
//ログレベル = INFO
//用途 = エラーじゃないけど、ログを出力したいときに利用する
func InfoLog(message string) {
	j, _ := json.Marshal(NewLogInfo(message))
	log.Print(string(j))
}

//アクセスログ
//ログレベル = INFO
//用途 = エンドポイントへのアクセスを記録するために利用する
func OutputAccessLog(u *url.URL) {
	j, _ := json.Marshal(NewAccessLog(u))
	log.Print(string(j))
}
