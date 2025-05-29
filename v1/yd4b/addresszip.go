package yd4b

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

// addressRequest は住所情報をもとに郵便番号を検索するための内部リクエスト構造体です。
// 各フィールドは API リクエストの JSON ボディにマッピングされます。
type addressRequest struct {
	PrefCode   string `json:"pref_code,omitempty"`   // 都道府県コード
	PrefName   string `json:"pref_name,omitempty"`   // 都道府県名
	PrefKana   string `json:"pref_kana,omitempty"`   // 都道府県名（カナ）
	PrefRoma   string `json:"pref_roma,omitempty"`   // 都道府県名（ローマ字）
	CityCode   string `json:"city_code,omitempty"`   // 市区町村コード
	CityName   string `json:"city_name,omitempty"`   // 市区町村名
	CityKana   string `json:"city_kana,omitempty"`   // 市区町村名（カナ）
	CityRoma   string `json:"city_roma,omitempty"`   // 市区町村名（ローマ字）
	TownName   string `json:"town_name,omitempty"`   // 町域名
	TownKana   string `json:"town_kana,omitempty"`   // 町域名（カナ）
	TownRoma   string `json:"town_roma,omitempty"`   // 町域名（ローマ字）
	Freeword   string `json:"freeword,omitempty"`    // フリーワード検索
	FlgGetCity int    `json:"flg_getcity,omitempty"` // 市区町村一覧取得フラグ（1: 有効）
	FlgGetPref int    `json:"flg_getpref,omitempty"` // 都道府県一覧取得フラグ（1: 有効）
	Page       int    `json:"page,omitempty"`        // ページ番号
	Limit      int    `json:"limit,omitempty"`       // 取得件数の上限
}

// addressRequestOption は addressRequest にオプションを適用するためのインターフェースです。
type addressRequestOption interface {
	apply(*addressRequest)
}

// addressRequestOptionFunc は addressRequestOption の関数型実装です。
type addressRequestOptionFunc func(*addressRequest)

// apply は addressRequestOptionFunc を適用し、addressRequest のフィールドを設定します。
func (f addressRequestOptionFunc) apply(r *addressRequest) {
	f(r)
}

// WithPrefCode は都道府県コードを指定するオプションです。
func WithPrefCode(code string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.PrefCode = code
	})
}

// WithPrefName は都道府県名を指定するオプションです。
func WithPrefName(name string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.PrefName = name
	})
}

// WithPrefKana は都道府県名（カナ）を指定するオプションです。
func WithPrefKana(kana string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.PrefKana = kana
	})
}

// WithPrefRoma は都道府県名（ローマ字）を指定するオプションです。
func WithPrefRoma(roma string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.PrefRoma = roma
	})
}

// WithCityCode は市区町村コードを指定するオプションです。
func WithCityCode(code string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.CityCode = code
	})
}

// WithCityName は市区町村名を指定するオプションです。
func WithCityName(name string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.CityName = name
	})
}

// WithCityKana は市区町村名（カナ）を指定するオプションです。
func WithCityKana(kana string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.CityKana = kana
	})
}

// WithCityRoma は市区町村名（ローマ字）を指定するオプションです。
func WithCityRoma(roma string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.CityRoma = roma
	})
}

// WithTownName は町域名を指定するオプションです。
func WithTownName(name string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.TownName = name
	})
}

// WithTownKana は町域名（カナ）を指定するオプションです。
func WithTownKana(kana string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.TownKana = kana
	})
}

// WithTownRoma は町域名（ローマ字）を指定するオプションです。
func WithTownRoma(roma string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.TownRoma = roma
	})
}

// WithFreeword はフリーワード検索語を指定するオプションです。
func WithFreeword(word string) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.Freeword = word
	})
}

// WithFlgGetCity は市区町村一覧取得フラグを指定するオプションです。
func WithFlgGetCity(flag int) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.FlgGetCity = flag
	})
}

// WithFlgGetPref は都道府県一覧取得フラグを指定するオプションです。
func WithFlgGetPref(flag int) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.FlgGetPref = flag
	})
}

// WithPage はページ番号を指定するオプションです。
func WithAZPage(page int) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.Page = page
	})
}

// WithLimit は取得件数の上限を指定するオプションです。
func WithAZLimit(limit int) addressRequestOption {
	return addressRequestOptionFunc(func(r *addressRequest) {
		r.Limit = limit
	})
}

// AddressResponse は住所から検索した郵便番号結果を表す構造体です。
type AddressResponse struct {
	Level     int           `json:"level"`     // 検索レベル
	Page      int           `json:"page"`      // 現在のページ番号
	Limit     int           `json:"limit"`     // １ページあたりの件数
	Count     int           `json:"count"`     // 総件数
	Addresses []AddressItem `json:"addresses"` // 検索結果の住所データ一覧
}

// AddressItem は住所検索結果の各アイテム（郵便番号含む）を表す構造体です。
type AddressItem struct {
	ZipCode  string `json:"zip_code"`  // 郵便番号
	PrefCode string `json:"pref_code"` // 都道府県コード
	PrefName string `json:"pref_name"` // 都道府県名
	PrefKana string `json:"pref_kana"` // 都道府県名（カナ）
	PrefRoma string `json:"pref_roma"` // 都道府県名（ローマ字）
	CityCode string `json:"city_code"` // 市区町村コード
	CityName string `json:"city_name"` // 市区町村名
	CityKana string `json:"city_kana"` // 市区町村名（カナ）
	CityRoma string `json:"city_roma"` // 市区町村名（ローマ字）
	TownName string `json:"town_name"` // 町域名
	TownKana string `json:"town_kana"` // 町域名（カナ）
	TownRoma string `json:"town_roma"` // 町域名（ローマ字）
}

// newAddressRequest はオプションを適用して addressRequest を生成します。
func newAddressRequest(opts ...addressRequestOption) addressRequest {
	var r addressRequest
	for _, opt := range opts {
		opt.apply(&r)
	}
	return r
}

// AddressZip は住所情報をもとに郵便番号を検索し、結果を返します。
// 引数:
//   - opts: 検索条件を指定する addressRequestOption。
//
// 戻り値:
//   - AddressResponse: 住所から取得した郵便番号検索結果
//   - error: 通信エラー、ステータスコード異常、デコード失敗など
func (c *Client) AddressZip(opts ...addressRequestOption) (res AddressResponse, err error) {
	// リクエストボディ用構造体を生成
	reqBody := newAddressRequest(opts...)

	// エンドポイント組み立て
	endpoint, err := url.JoinPath(c.origin, "api", c.version, "addresszip")
	if err != nil {
		err = errors.Join(NewError(500, "endpoint error"), err)
		return
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		err = errors.Join(NewError(500, "url parse error"), err)
		return
	}
	if c.ecuid != "" {
		q := u.Query()
		q.Set("ec_uid", c.ecuid)
		u.RawQuery = q.Encode()
	}

	// JSON エンコーディング
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		err = errors.Join(NewError(500, "json encoding error"), err)
		return
	}

	// HTTP リクエスト生成
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		err = errors.Join(NewError(500, "request creation error"), err)
		return
	}

	// リクエスト送信
	resp, err := c.do(req)
	if err != nil {
		err = errors.Join(NewError(500, "client do error"), err)
		return
	}
	defer resp.Body.Close()

	// ステータスコード確認
	if resp.StatusCode != http.StatusOK {
		err = NewError(resp.StatusCode, "unexpected status code")
		return
	}

	// レスポンス JSON デコード
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		err = errors.Join(NewError(500, "json decoding error"), err)
		return
	}

	return
}
