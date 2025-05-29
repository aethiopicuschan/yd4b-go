package yd4b

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// searchcodeRequest はコード番号検索（郵便番号・事業所個別郵便番号・デジタルアドレス）を行うための内部リクエスト構造体です。
type searchcodeRequest struct {
	SearchCode string // パスパラメータ：検索コード
	Page       int    `json:"page,omitempty"`       // ページ番号
	Limit      int    `json:"limit,omitempty"`      // 取得最大レコード数
	Choikitype int    `json:"choikitype,omitempty"` // 町域フィールドタイプ（1:括弧なし、2:括弧あり）
	Searchtype int    `json:"searchtype,omitempty"` // 検索方法タイプ（1:全対象、2:事業所郵便除外）
}

// searchcodeOption は searchcodeRequest にオプションを適用するためのインターフェースです。
type searchcodeOption interface {
	apply(*searchcodeRequest)
}

type searchcodeOptionFunc func(*searchcodeRequest)

func (f searchcodeOptionFunc) apply(r *searchcodeRequest) { f(r) }

// WithSCPage はsearchcodeにおいてページ番号を指定するオプションです。
func WithSCPage(page int) searchcodeOption {
	return searchcodeOptionFunc(func(r *searchcodeRequest) {
		r.Page = page
	})
}

// WithLimit はsearchcodeにおいて取得最大件数を指定するオプションです。
func WithSCLimit(limit int) searchcodeOption {
	return searchcodeOptionFunc(func(r *searchcodeRequest) {
		r.Limit = limit
	})
}

// WithChoikitype は町域フィールドタイプを指定するオプションです（1:括弧なし、2:括弧あり）。
func WithChoikitype(ct int) searchcodeOption {
	return searchcodeOptionFunc(func(r *searchcodeRequest) {
		r.Choikitype = ct
	})
}

// WithSearchtype は検索方法タイプを指定するオプションです（1:全対象、2:事業所郵便除外）。
func WithSearchtype(st int) searchcodeOption {
	return searchcodeOptionFunc(func(r *searchcodeRequest) {
		r.Searchtype = st
	})
}

// newSearchcodeRequest は必須の search_code とオプションから searchcodeRequest を生成します。
func newSearchcodeRequest(code string, opts ...searchcodeOption) *searchcodeRequest {
	r := &searchcodeRequest{SearchCode: code}
	for _, opt := range opts {
		opt.apply(r)
	}
	return r
}

// SearchcodeResponse はコード番号検索のレスポンスを表す構造体です。
type SearchcodeResponse struct {
	Page       int                     `json:"page"`       // ページ数
	Limit      int                     `json:"limit"`      // 取得最大レコード数
	Count      int                     `json:"count"`      // 該当データ数
	Searchtype string                  `json:"searchtype"` // 検索タイプ（"dgacode" / "zipcode" / "bizzipcode"）
	Addresses  []SearchcodeAddressItem `json:"addresses"`  // 検索結果の住所情報一覧
}

// SearchcodeAddressItem はコード番号検索結果の各アイテムを表す構造体です。
type SearchcodeAddressItem struct {
	DgaCode   *string  `json:"dgacode"`    // デジタルアドレスコード（nullable）
	ZipCode   string   `json:"zip_code"`   // 郵便番号
	PrefCode  string   `json:"pref_code"`  // 都道府県コード
	PrefName  string   `json:"pref_name"`  // 都道府県名
	PrefKana  *string  `json:"pref_kana"`  // 都道府県名カナ（nullable）
	PrefRoma  *string  `json:"pref_roma"`  // 都道府県名ローマ字（nullable）
	CityCode  string   `json:"city_code"`  // 市区町村コード
	CityName  string   `json:"city_name"`  // 市区町村名
	CityKana  *string  `json:"city_kana"`  // 市区町村名カナ（nullable）
	CityRoma  *string  `json:"city_roma"`  // 市区町村名ローマ字（nullable）
	TownName  string   `json:"town_name"`  // 町域名
	TownKana  *string  `json:"town_kana"`  // 町域名カナ（nullable）
	TownRoma  *string  `json:"town_roma"`  // 町域名ローマ字（nullable）
	BizName   *string  `json:"biz_name"`   // 事業所名（nullable）
	BizKana   *string  `json:"biz_kana"`   // 事業所名カナ（nullable）
	BizRoma   *string  `json:"biz_roma"`   // 事業所名ローマ字（nullable）
	BlockName *string  `json:"block_name"` // 町域字等（nullable）
	OtherName *string  `json:"other_name"` // その他名称（nullable）
	Address   *string  `json:"address"`    // 住所（nullable）
	Longitude *float64 `json:"longitude"`  // 経度（nullable）
	Latitude  *float64 `json:"latitude"`   // 緯度（nullable）
}

// Searchcode はコード番号検索エンドポイントを叩き、結果を返します。
// 引数:
//   - code: 検索する郵便番号・事業所個別郵便番号・デジタルアドレス
//   - opts: ページ番号や取得件数、フィールドタイプなどのオプション
func (c *Client) Searchcode(code string, opts ...searchcodeOption) (resp SearchcodeResponse, err error) {
	// リクエスト構築
	reqDTO := newSearchcodeRequest(code, opts...)

	// エンドポイント組み立て
	endpoint, err := url.JoinPath(c.origin, "api", c.version, "searchcode", reqDTO.SearchCode)
	if err != nil {
		err = errors.Join(NewError(500, "endpoint error"), err)
		return
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		err = errors.Join(NewError(500, "url parse error"), err)
		return
	}

	// クエリパラメータ設定
	q := u.Query()
	if c.ecuid != "" {
		q.Set("ec_uid", c.ecuid)
	}
	if reqDTO.Page > 0 {
		q.Set("page", fmt.Sprint(reqDTO.Page))
	}
	if reqDTO.Limit > 0 {
		q.Set("limit", fmt.Sprint(reqDTO.Limit))
	}
	if reqDTO.Choikitype > 0 {
		q.Set("choikitype", fmt.Sprint(reqDTO.Choikitype))
	}
	if reqDTO.Searchtype > 0 {
		q.Set("searchtype", fmt.Sprint(reqDTO.Searchtype))
	}
	u.RawQuery = q.Encode()

	// HTTP リクエスト生成
	httpReq, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		err = errors.Join(NewError(500, "request creation error"), err)
		return
	}

	// 実行
	httpResp, err := c.do(httpReq)
	if err != nil {
		err = errors.Join(NewError(500, "client do error"), err)
		return
	}
	defer httpResp.Body.Close()

	// ステータスコード確認
	if httpResp.StatusCode != http.StatusOK {
		err = NewError(httpResp.StatusCode, "unexpected status code")
		return
	}

	// デコード
	if err = json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		err = errors.Join(NewError(500, "json decoding error"), err)
		return
	}

	return resp, nil
}
