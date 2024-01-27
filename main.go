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
	http.HandleFunc("/favicon.ico", func (w http.ResponseWriter, r *http.Request) {
	})
	http.HandleFunc("/add_memo", addMemo)
	http.HandleFunc("/update_memo", updateMemo)
	http.HandleFunc("/list_memos", listMemos)
	http.HandleFunc("/delete_memos", deleteMemos)
	http.ListenAndServe(":8080", nil)
}

// curl localhost:8080/
func showHTML(w http.ResponseWriter, r *http.Request) {
	OutputAccessLog(r.URL)

	// このエンドポイントはJSONを受け付けないので、
	// ValidateHTTPRequest() は使えない。
	if r.Method != http.MethodGet {
		WarningLog(fmt.Sprintf("invalid http method = %s", r.Method))
		errMsgs := []*ErrorMessage{
			NewErrorMessage(
				"INVALID_HTTP_METHOD",
				"無効なリクエストです。",
			),
		}
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

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
			"タイトルの文字数は1文字以上30文字以下にしてください。",
		),
		)
		WarningLog(fmt.Sprintf("title length is invalid, title = %s, length = %d)", m.Title, titleLength))
	}

	bodyLength := len([]rune(m.Body))

	//メモの本文が1文字未満、100文字より長い場合はエラーにする。
	if bodyLength < 1 || bodyLength > 500 {
		errMsgs = append(errMsgs, NewErrorMessage(
			"InvalidBody",
			"本文の文字数は1文字以上100文字以下にしてください。",
		),
		)
		WarningLog(fmt.Sprintf("body length is invalid, body = %s, length = %d)", m.Body, bodyLength))
	}

	return errMsgs
}

var memos *Memos = NewMemos()

type Memos struct {
	Memos []*Memo
}

func NewMemos() *Memos {
	return &Memos{
		Memos: []*Memo{},
	}
}

func (ms *Memos) DeleteMemoByID(id int) {
	//以下のスライスを例にする
	//[id=100（0番目）, id=200（1番目）, id=300（2番目）]
	//削除対象はid=200（1番目）とする。

	//削除対象のメモがMemosの何番目にあるか
	//初期値として -1 を指定する。
	position := -1

	for i, m := range ms.Memos {
		if m.ID == id {
			//削除対象のメモがあったら、
			//positionにi（スライスのインデックス）を代入する。
			//
			//positionには、id=200（1番目）の1という値が入る。
			position = i
			break
		}
	}

	//positionの値に変化がない（メモが存在しない場合）は、
	//削除処理をせずに return する。
	//スライスは0番目から始まるので、
	//positionの初期値が0だと上手く動かないので注意すること。
	if position == -1 {
		return
	}

	//Memosの一番最後の要素を削除対象の要素に上書きする。
	//スライスは以下の状態になる。
	//[id=100（0番目）, id=300（1番目）, id=300（2番目）]
	ms.Memos[position] = ms.Memos[len(ms.Memos)-1]

	//Memosの一番最後の要素以外のスライスをnewMemosに代入する。
	//スライスの":"については補足ブログで取り上げます。
	//newMemosのスライスは先頭の2つの要素のみ保持することになる。
	//[id=100（0番目）, id=300（1番目）]
	newMemos := ms.Memos[:len(ms.Memos)-1]

	//newMemosをMemosに代入する。
	ms.Memos = newMemos
}

func (ms *Memos) GetMemoByID(id int) *Memo {
	for _, m := range ms.Memos {
		if m.ID == id {
			return m
		}
	}
	return nil
}

func (ms *Memos) UpdateMemo(m *Memo) []*ErrorMessage {
	if memo := ms.GetMemoByID(m.ID); memo == nil {
		WarningLog(fmt.Sprintf("memo is not found = %d", m.ID))
		errMsgs := []*ErrorMessage{
			NewErrorMessage(
				"MEMO_ID_FORMAT_IS_INVALID",
				"対象のメモは存在しません。",
			),
		}
		return errMsgs
	}

	errMsg := ms.Validate(m)

	if len(errMsg) > 0 {
		return errMsg
	}

	for i, memo := range ms.Memos {
		if memo.ID == m.ID {
			ms.Memos[i] = m
			break
		}
	}

	return nil
}

func (ms *Memos) Validate(m *Memo) []*ErrorMessage {
	errMsg := m.Validate()

	for _, memo := range ms.Memos {
		if m.ID == memo.ID {
			//更新対象のメモとタイトルが同じでも問題ない
			break
		}
		if m.Title == memo.Title {
			//タイトルが同じなのでエラーにする
			errMsg = append(
				errMsg,
				NewErrorMessage("TItleIsDuplicated", "タイトルが重複しています。"),
			)
			WarningLog(fmt.Sprintf("title is duplicated, title = %s", m.Title))
			break
		}
	}
	return errMsg
}

func (ms *Memos) AddMemo(m *Memo) []*ErrorMessage {
	errMsg := ms.Validate(m)

	if memo := ms.GetMemoByID(m.ID); memo != nil {
		errMsg = append(
			errMsg,
			NewErrorMessage("IDIsDuplicated", "IDが重複しています。"),
		)
		WarningLog(fmt.Sprintf("id is duplicated, id = %d", m.ID))

	}

	if len(errMsg) > 0 {
		return errMsg
	}

	ms.Memos = append(ms.Memos, m)

	return nil
}

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

func ValidateHTTPRequest(r *http.Request, validHTTPMethod string) []*ErrorMessage {
	errMsgs := make([]*ErrorMessage, 0)

	if c := r.Header.Get("Content-Type"); c != "application/json" {
		WarningLog(fmt.Sprintf("invalid content-type = %s", c))
		errMsgs = append(errMsgs, NewErrorMessage(
			"INVALID_CONTENT_TYPE",
			"無効なContent-Typeです。",
		),
		)
	}

	if r.Method != validHTTPMethod {
		WarningLog(fmt.Sprintf("invalid http method = %s", r.Method))
		errMsgs = append(errMsgs, NewErrorMessage(
			"INVALID_HTTP_METHOD",
			"無効なHTTP Methodです。",
		))
	}

	return errMsgs
}

//メモを登録する。
//curl -X POST -H "Content-Type: application/json" -d '{"ID":1111,"Title":"mytitle","Body":"mybody","CreatedAt":"2022-01-01T10:00:00+09:00","UpdatedAt":"2022-01-01T11:00:00+09:00"}' localhost:8080/add_memo
func addMemo(w http.ResponseWriter, r *http.Request) {
	OutputAccessLog(r.URL)

	if errMsgs := ValidateHTTPRequest(r, http.MethodPost); len(errMsgs) > 0 {
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

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

	//メモを保存する。
	errMsgs := memos.AddMemo(m)

	if len(errMsgs) > 0 {
		// クライアント側から送信されるリクエストに問題があるので、
		// HTTP Status 400 を返す。
		// Go言語では定数としてHTTP Statusが用意されているので、
		// それを利用するのがいいと思います。
		// https://developer.mozilla.org/ja/docs/Web/HTTP/Status/400

		//警告ログはバリデーションするところで出力するので、
		//ここでは出力しなくていい。
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

	//HTTP Response は空にするので、nilを指定する。
	//len()は配列やマップなどの長さを出力することができる関数です。
	fmt.Fprintln(w, len(memos.Memos))
}

//curl -X PUT -H "Content-Type: application/json" -d '{"ID":1111,"Title":"mytitle2","Body":"mybody","CreatedAt":"2022-01-01T10:00:00+09:00","UpdatedAt":"2022-01-01T11:00:00+09:00"}' localhost:8080/update_memo
func updateMemo(w http.ResponseWriter, r *http.Request) {
	OutputAccessLog(r.URL)

	if errMsgs := ValidateHTTPRequest(r, http.MethodPut); len(errMsgs) > 0 {
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

	m := &Memo{}

	//HTTPリクエストで送信されてきた HTTP Request Body(JSON形式)を
	//Memo構造体にセットしている。
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		RespondInternalServerError(w, err.Error())
		return
	}

	//メモを保存する。
	errMsgs := memos.UpdateMemo(m)

	if len(errMsgs) > 0 {
		//警告ログはバリデーションするところで出力するので、
		//ここでは出力しなくていい。
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

	fmt.Fprintln(w, len(memos.Memos))
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

	if errMsgs := ValidateHTTPRequest(r, http.MethodGet); len(errMsgs) > 0 {
		RespondError(w, http.StatusBadRequest, errMsgs)
		return
	}

	b, err := json.Marshal(memos.Memos)
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

	if errMsgs := ValidateHTTPRequest(r, http.MethodDelete); len(errMsgs) > 0 {
		RespondError(w, http.StatusBadRequest, errMsgs)
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

	m := memos.GetMemoByID(idInt)
	if m == nil {
		WarningLog(fmt.Sprintf("memo id = %d is empty", idInt))
		RespondNotFoundError(w)
		return
	}

	memos.DeleteMemoByID(idInt)

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
