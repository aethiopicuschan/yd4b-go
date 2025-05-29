# yd4b-go

[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen?style=flat-square)](/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/aethiopicuschan/yd4b-go.svg)](https://pkg.go.dev/github.com/aethiopicuschan/yd4b-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/aethiopicuschan/yd4b-go)](https://goreportcard.com/report/github.com/aethiopicuschan/yd4b-go)
[![CI](https://github.com/aethiopicuschan/yd4b-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/aethiopicuschan/yd4b-go/actions/workflows/ci.yaml)

`yd4b-go`は[郵便番号・デジタルアドレス for Biz](https://guide-biz.da.pf.japanpost.jp/)のGolang向けクライアントライブラリです。

## インストール

```sh
go get -u github.com/aethiopicuschan/yd4b-go
```

## 動作に必要なもの

- Golang 1.24.2以降
- [郵便番号・デジタルアドレス for Biz](https://guide-biz.da.pf.japanpost.jp/)のアカウント
- APIのオリジン
  - APIリファレンスでダウンロードできるOpenAPI specificationに記載されています
- クライアントIDとクライアントシークレット
  - システムを新規登録した際に取得できます
- 送信元IPアドレス
  - 実際にAPIを利用する端末のグローバルIPアドレスを指定します

詳しくは以下のサンプルコードを参照してください。

## 利用例

```go
package main

import (
	"log"

	"github.com/aethiopicuschan/yd4b-go/v1/yd4b"
)

func main() {
	// APIクライアントを作成
	client := yd4b.NewClient("https://example.com", "Your Client ID", "Your Client secret", "Your global ip address")

	// トークンを取得してクライアントにセットする
	res, err := client.GetToken()
	if err != nil {
		log.Fatal(err)
	}
	client.SetToken(res.Token)

	// 郵便番号、事業所個別郵便番号、デジタルアドレスから住所を検索
	res2, err := client.Searchcode("1000001")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res2.Addresses[0].PrefName, res2.Addresses[0].CityName, res2.Addresses[0].TownName)
	// Output: 東京都 千代田区 千代田

	// 住所から郵便番号を検索
	res3, err := client.AddressZip(yd4b.WithPrefName("東京都"), yd4b.WithCityName("千代田区"), yd4b.WithTownName("千代田"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res3.Addresses[0])
	// Output: {1000001 13 東京都 トウキョウト TOKYO 13101 千代田区 チヨダク CHIYODA-KU 千代田 チヨダ CHIYODA}
}
```

## オプションについて

`Searchcode` と `AddressZip` はFunctional Option Patternでオプションを設定できます。「With...」という関数がそれです。上記サンプルコードでも一部利用していますが、詳細は[ドキュメント](https://pkg.go.dev/github.com/aethiopicuschan/yd4b-go)を参照してください。

## カスタムHTTPクライアント

デフォルトではHTTPリクエストに `http.DefaultClient.Do` を使用していますが、必要に応じてカスタムHTTPクライアントを設定できます。以下の型の関数を受け付けます。

```go
func(req *http.Request) (*http.Response, error)
```

用意した関数を `yd4b.Client` の `SetDoFunc` に渡すことで、カスタムHTTPクライアントを設定できます。

## エラーハンドリング

独自の `Error` 型を定義しています。これによって、以下のような実装が可能です。

```go
if err := someFn(); err != nil {
	var yd4berr *yd4b.Error
	if errors.As(err, &yd4berr) {
		http.Error(w, yd4berr.Message, yd4berr.StatusCode)
		return
	}
	http.Error(w, "internal server error", 500)
}
```
