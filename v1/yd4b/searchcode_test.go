package yd4b_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/aethiopicuschan/yd4b-go/v1/yd4b"
	"github.com/stretchr/testify/assert"
)

func TestExportNewSearchcodeRequest(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		opts      []yd4b.SearchcodeOption
		wantPage  int
		wantLimit int
		wantCT    int
		wantST    int
	}{
		{
			name:     "no options",
			code:     "ABC",
			opts:     nil,
			wantPage: 0, wantLimit: 0, wantCT: 0, wantST: 0,
		},
		{
			name:     "page and limit",
			code:     "XYZ",
			opts:     []yd4b.SearchcodeOption{yd4b.WithSCPage(2), yd4b.WithSCLimit(10)},
			wantPage: 2, wantLimit: 10, wantCT: 0, wantST: 0,
		},
		{
			name:     "choikitype and searchtype",
			code:     "123",
			opts:     []yd4b.SearchcodeOption{yd4b.WithChoikitype(1), yd4b.WithSearchtype(2)},
			wantPage: 0, wantLimit: 0, wantCT: 1, wantST: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := yd4b.NewSearchcodeRequest(tt.code, tt.opts...)
			assert.Equal(t, tt.code, r.SearchCode)
			assert.Equal(t, tt.wantPage, r.Page)
			assert.Equal(t, tt.wantLimit, r.Limit)
			assert.Equal(t, tt.wantCT, r.Choikitype)
			assert.Equal(t, tt.wantST, r.Searchtype)
		})
	}
}

func TestClient_Searchcode(t *testing.T) {
	tests := []struct {
		name          string
		ecuid         string
		opts          []yd4b.SearchcodeOption
		doFunc        func(req *http.Request) (*http.Response, error)
		wantErrSubstr string
		wantQuery     map[string]string
		wantResp      yd4b.SearchcodeResponse
	}{
		{
			name: "success without params",
			doFunc: func(req *http.Request) (*http.Response, error) {
				// assert no extra query params
				assert.Empty(t, req.URL.Query().Get("page"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"page":0,"limit":0,"count":0,"searchtype":"zipcode","addresses":[]}`)),
				}, nil
			},
			wantResp: yd4b.SearchcodeResponse{Page: 0, Limit: 0, Count: 0, Searchtype: "zipcode", Addresses: []yd4b.SearchcodeAddressItem{}},
		},
		{
			name:  "with all params",
			ecuid: "EC1",
			opts:  []yd4b.SearchcodeOption{yd4b.WithSCPage(3), yd4b.WithSCLimit(5), yd4b.WithChoikitype(2), yd4b.WithSearchtype(1)},
			doFunc: func(req *http.Request) (*http.Response, error) {
				q := req.URL.Query()
				assert.Equal(t, "EC1", q.Get("ec_uid"))
				assert.Equal(t, "3", q.Get("page"))
				assert.Equal(t, "5", q.Get("limit"))
				assert.Equal(t, "2", q.Get("choikitype"))
				assert.Equal(t, "1", q.Get("searchtype"))
				// return dummy valid JSON
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"page":3,"limit":5,"count":1,"searchtype":"dgacode","addresses":[{"dgacode":null,"zip_code":"0000000","pref_code":"01","pref_name":"Hokkaido","city_code":"100","city_name":"Sapporo","town_name":"Chuo","searchtype":""}]}`)),
				}, nil
			},
			wantResp: yd4b.SearchcodeResponse{Page: 3, Limit: 5, Count: 1, Searchtype: "dgacode", Addresses: []yd4b.SearchcodeAddressItem{{ZipCode: "0000000", PrefCode: "01", PrefName: "Hokkaido", CityCode: "100", CityName: "Sapporo", TownName: "Chuo"}}},
		},
		{
			name: "client do error",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("fail")
			},
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
			wantErrSubstr: "json decoding error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := yd4b.NewClient("https://api.test", "id", "secret", "1.2.3.4")
			if tt.ecuid != "" {
				client.SetECUID(tt.ecuid)
			}
			client.SetDoFunc(tt.doFunc)

			resp, err := client.Searchcode("CODE", tt.opts...)

			if tt.wantErrSubstr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrSubstr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantResp.Page, resp.Page)
			assert.Equal(t, tt.wantResp.Limit, resp.Limit)
			assert.Equal(t, tt.wantResp.Count, resp.Count)
			assert.Equal(t, tt.wantResp.Searchtype, resp.Searchtype)
			assert.Equal(t, len(tt.wantResp.Addresses), len(resp.Addresses))
		})
	}
}
