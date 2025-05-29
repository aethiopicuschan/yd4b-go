// 郵便番号・デジタルアドレス for Bizのクライアントライブラリ
package yd4b

import (
	"net/http"
)

// クライアントの実体
type Client struct {
	version      string                                          // APIのバージョン
	origin       string                                          // APIサーバのオリジン
	clientID     string                                          // クライアントID
	clientSecret string                                          // クライアントシークレット
	token        string                                          // API利用トークン
	myip         string                                          // クライアントのグローバルIPアドレス（x-forwarded-for ヘッダに設定）
	ecuid        string                                          // プロバイダーのユーザーID
	doFunc       func(req *http.Request) (*http.Response, error) // HTTPクライアントのDoメソッドをラップする関数
}

// [Client]のコンストラクタ
func NewClient(origin string, clientID string, clientSecret string, myip string) *Client {
	return &Client{
		version:      "v1",
		origin:       origin,
		clientID:     clientID,
		clientSecret: clientSecret,
		myip:         myip,
		ecuid:        "",
		doFunc:       http.DefaultClient.Do,
	}
}

// APIのバージョンを返す
func (c *Client) Version() string {
	return c.version
}

// Doメソッドを書き換える
func (c *Client) SetDoFunc(do func(req *http.Request) (*http.Response, error)) {
	c.doFunc = do
}

// API利用トークンを設定する
func (c *Client) SetToken(token string) {
	c.token = token
}

// API利用トークンが設定されているかどうかを確認する
func (c *Client) HasToken() bool {
	return c.token != ""
}

// ECUIDを設定する
func (c *Client) SetECUID(ecuid string) {
	c.ecuid = ecuid
}

// 設定されたDoメソッドを実行する
func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-forwarded-for", c.myip)
	if c.HasToken() {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.doFunc(req)
}
