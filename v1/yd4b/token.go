package yd4b

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// TokenRequest は API 利用トークン取得のためのリクエストボディです。
type TokenRequest struct {
	GrantType string `json:"grant_type"` // 認可タイプ（client_credentials 固定）
	ClientID  string `json:"client_id"`  // クライアントID
	SecretKey string `json:"secret_key"` // シークレットキー
}

// ToRequest は TokenRequest を HTTP POST リクエストに変換します。
//
// 引数:
//   - endpoint: トークン取得APIのエンドポイントURL
//   - myip: クライアントのグローバルIPアドレス（x-forwarded-for ヘッダに設定）
//
// 戻り値:
//   - *http.Request: 生成されたHTTPリクエスト
//   - error: 失敗した場合のエラー
func (t *TokenRequest) ToRequest(endpoint string) (req *http.Request, err error) {
	jsonBody, err := json.Marshal(t)
	if err != nil {
		return
	}
	req, err = http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return
	}
	return
}

// TokenResponse はトークン取得APIのレスポンス形式です。
type TokenResponse struct {
	Scope     string `json:"scope"`      // トークンスコープ（例: "J1"）
	TokenType string `json:"token_type"` // トークンタイプ（例: "Bearer"）
	ExpiresIn int64  `json:"expires_in"` // 有効期限（秒数）
	Token     string `json:"token"`      // アクセストークン（"Token: " プレフィックス付きの場合あり）
}

// GetToken はトークン取得APIを呼び出し、アクセストークンを取得します。
//
// 引数:
//   - myip: クライアントのIPアドレス（x-forwarded-for ヘッダとして使用）
//
// 戻り値:
//   - TokenResponse: トークン情報（スコープ、タイプ、有効秒数、トークン）
//   - error: 通信エラー、ステータスコード異常、デコード失敗など
func (c *Client) GetToken() (res TokenResponse, err error) {
	endpoint, err := url.JoinPath(c.origin, "api", c.version, "j", "token")
	if err != nil {
		err = errors.Join(NewError(500, "endpoint error"), err)
		return
	}

	body := &TokenRequest{
		GrantType: "client_credentials",
		ClientID:  c.clientID,
		SecretKey: c.clientSecret,
	}
	req, err := body.ToRequest(endpoint)
	if err != nil {
		err = errors.Join(NewError(500, "request creation error"), err)
		return
	}

	resp, err := c.do(req)
	if err != nil {
		err = errors.Join(NewError(500, "client do error"), err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = NewError(resp.StatusCode, "unexpected status code")
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		err = errors.Join(NewError(500, "json decoding error"), err)
		return
	}

	// "Token: xxx" の形式で返ってきた場合に "Token: " を削除
	res.Token = strings.TrimPrefix(res.Token, "Token: ")

	return
}
