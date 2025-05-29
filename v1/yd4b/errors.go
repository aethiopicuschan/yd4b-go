package yd4b

import "encoding/json"

// 独自のエラー型
type Error struct {
	StatusCode int    `json:"status_code"` // HTTPステータスコード
	Message    string `json:"message"`     // エラーメッセージ
}

// Errorを生成する
func NewError(statusCode int, message string) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    message,
	}
}

// errorインターフェースを実装する
func (e *Error) Error() string {
	return e.Message
}

// JSONにする
func (e *Error) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
