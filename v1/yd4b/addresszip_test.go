package yd4b_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/aethiopicuschan/yd4b-go/v1/yd4b"
	"github.com/stretchr/testify/assert"
)

// Test for newAddressRequest via export
func TestNewAddressRequest(t *testing.T) {
	tests := []struct {
		name         string
		opts         []yd4b.AddressRequestOption
		wantPrefCode string
		wantFreeword string
		wantPage     int
		wantLimit    int
	}{
		{
			name:         "no options",
			opts:         nil,
			wantPrefCode: "",
			wantFreeword: "",
			wantPage:     0,
			wantLimit:    0,
		},
		{
			name: "set fields",
			opts: []yd4b.AddressRequestOption{
				yd4b.WithPrefCode("13"),
				yd4b.WithFreeword("中央区"),
				yd4b.WithAZPage(2),
				yd4b.WithAZLimit(50),
			},
			wantPrefCode: "13",
			wantFreeword: "中央区",
			wantPage:     2,
			wantLimit:    50,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := yd4b.NewAddressRequest(tc.opts...)
			assert.Equal(t, tc.wantPrefCode, r.PrefCode)
			assert.Equal(t, tc.wantFreeword, r.Freeword)
			assert.Equal(t, tc.wantPage, r.Page)
			assert.Equal(t, tc.wantLimit, r.Limit)
		})
	}
}

func TestClient_AddressZip(t *testing.T) {
	tests := []struct {
		name          string
		ecuid         string
		opts          []yd4b.AddressRequestOption
		doFunc        func(req *http.Request) (*http.Response, error)
		wantErrSubstr string
		wantResp      yd4b.AddressResponse
	}{
		{
			name: "success simple",
			doFunc: func(req *http.Request) (*http.Response, error) {
				var b yd4b.AddressRequest
				_ = json.NewDecoder(req.Body).Decode(&b)
				assert.Empty(t, b.PrefCode)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"level":1,"page":0,"limit":0,"count":0,"addresses":[]}`)),
				}, nil
			},
			wantResp: yd4b.AddressResponse{Level: 1, Page: 0, Limit: 0, Count: 0, Addresses: []yd4b.AddressItem{}},
		},
		{
			name:  "with options + ecuid",
			ecuid: "EC42",
			opts: []yd4b.AddressRequestOption{
				yd4b.WithPrefCode("13"),
				yd4b.WithPrefName("東京都"),
				yd4b.WithPrefKana("トウキョウト"),
				yd4b.WithPrefRoma("TOKYO"),
				yd4b.WithCityCode("13101"),
				yd4b.WithCityName("千代田区"),
				yd4b.WithCityKana("チヨダク"),
				yd4b.WithCityRoma("CHIYODA-KU"),
				yd4b.WithTownName("千代田"),
				yd4b.WithTownKana("チヨダイダ"),
				yd4b.WithTownRoma("CHIYODAI-DA"),
				yd4b.WithFreeword("銀座"),
				yd4b.WithFlgGetCity(1),
				yd4b.WithFlgGetPref(1),
				yd4b.WithAZPage(3),
				yd4b.WithAZLimit(20),
			},
			doFunc: func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "EC42", req.URL.Query().Get("ec_uid"))
				var b yd4b.AddressRequest
				_ = json.NewDecoder(req.Body).Decode(&b)
				assert.Equal(t, "13", b.PrefCode)
				assert.Equal(t, 1, b.FlgGetCity)
				assert.Equal(t, 20, b.Limit)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewBufferString(`
                        {"level":2,"page":3,"limit":20,"count":1,
                        "addresses":[{"zip_code":"1000001","pref_code":"13","pref_name":"東京都",
                        "pref_kana":"トウキョウト","pref_roma":"TOKYO","city_code":"13101",
                        "city_name":"千代田区","city_kana":"チヨダク","city_roma":"CHIYODA-KU",
                        "town_name":"千代田","town_kana":"チヨダイダ","town_roma":"CHIYODAI-DA"}]}`)),
				}, nil
			},
			wantResp: yd4b.AddressResponse{
				Level:     2,
				Page:      3,
				Limit:     20,
				Count:     1,
				Addresses: []yd4b.AddressItem{{ZipCode: "1000001", PrefCode: "13", PrefName: "東京都", PrefKana: "トウキョウト", PrefRoma: "TOKYO", CityCode: "13101", CityName: "千代田区", CityKana: "チヨダク", CityRoma: "CHIYODA-KU", TownName: "千代田", TownKana: "チヨダイダ", TownRoma: "CHIYODAI-DA"}},
			},
		},
		{
			name:          "client do error",
			doFunc:        func(req *http.Request) (*http.Response, error) { return nil, errors.New("net fail") },
			wantErrSubstr: "client do error",
		},
		{
			name: "non-200 status",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewBufferString(`err`))}, nil
			},
			wantErrSubstr: "unexpected status code",
		},
		{
			name: "decoding error",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{bad}`))}, nil
			},
			wantErrSubstr: "decoding error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := yd4b.NewClient("https://api.example.com", "id", "secret", "1.2.3.4")
			if tc.ecuid != "" {
				client.SetECUID(tc.ecuid)
			}
			if tc.doFunc != nil {
				client.SetDoFunc(tc.doFunc)
			}

			res, err := client.AddressZip(tc.opts...)
			if tc.wantErrSubstr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSubstr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResp, res)
		})
	}
}
