package fcm

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// fcm
type Fcm struct {
	AppID  string
	BizID  string
	Key    string
	aes    cipher.AEAD
	keys   []string
	client http.Client
}

// check data
type Check struct {
	Ai    string `json:"ai"`
	Name  string `json:"name"`
	IdNum string `json:"idNum"`
}

// query data
type Query struct {
	Ai string `json:"ai"`
}

// login or logout
type Behavior struct {
	No int    `json:"no"`
	Si string `json:"si"`
	Bt int    `json:"bt"`
	Ot int64  `json:"ot"`
	Ct int    `json:"ct"`
	Di string `json:"di"`
	Pi string `json:"pi"`
}

type Result struct {
	ErrCode int         `json:"errcode"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
}

// constructor
func NewFcm(appId, bizId, key string) Fcm {
	fcm := Fcm{
		AppID: appId,
		BizID: bizId,
		Key:   key,
		keys:  []string{"appId", "bizId", "timestamps"},
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
	// cipher
	b, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	block, err := aes.NewCipher(b)
	if err != nil {
		panic(err)
	}
	// aead
	AEAD, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	fcm.aes = AEAD
	return fcm
}

// check api
func (f *Fcm) Check(c *Check) (Result, error) {
	url := "https://api.wlc.nppa.gov.cn/idcard/authentication/check"
	header := f.getHeader()
	header["Content-Type"] = []string{"application/json; charset=utf-8"}
	return f.request("POST", url, c, header)
}

// test check api
func (f *Fcm) TestCheck(c *Check, testCode string) (Result, error) {
	uri := "https://wlc.nppa.gov.cn/test/authentication/check/" + testCode
	header := f.getHeader()
	header["Content-Type"] = []string{"application/json; charset=utf-8"}
	return f.request("POST", uri, c, header)
}

// query api
func (f *Fcm) Query(q *Query) (Result, error) {
	uri := "http://api2.wlc.nppa.gov.cn/idcard/authentication/query"
	header := f.getHeader()
	return f.request("GET", uri, q, header)
}

// test query api
func (f *Fcm) TestQuery(q *Query, testCode string) (Result, error) {
	uri := "https://wlc.nppa.gov.cn/test/authentication/query/" + testCode
	header := f.getHeader()
	return f.request("GET", uri, q, header)
}

// login or logout
func (f *Fcm) LoginOrOut(b *[]Behavior) (Result, error) {
	url := "http://api2.wlc.nppa.gov.cn/behavior/collection/loginout"
	header := f.getHeader()
	header["Content-Type"] = []string{"application/json; charset=utf-8"}
	return f.request("POST", url, b, header)
}

// test login or logout
func (f *Fcm) TestLoginOrOut(q *[]Behavior, testCode string) (Result, error) {
	uri := "https://wlc.nppa.gov.cn/test/collection/loginout/" + testCode
	header := f.getHeader()
	header["Content-Type"] = []string{"application/json; charset=utf-8"}
	return f.request("POST", uri, q, header)
}

// aes-128-gcm + base64
func (f *Fcm) makeBody(body []byte) (string, error) {
	//random bytes
	nonce := make([]byte, f.aes.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	data := append(nonce, f.aes.Seal(nil, nonce, body, nil)...)
	return base64.StdEncoding.EncodeToString(data), nil
}

// sha256
func (f *Fcm) makeSign(header http.Header, body string, query map[string]string) string {
	// except content-type
	header.Del("Content-Type")
	ks := f.keys
	for k, v := range query {
		ks = append(ks, k)
		header[k] = []string{v} //maybe lower case
	}
	sort.Strings(ks)
	raw := ""
	for _, k := range ks {
		raw += k + header[k][0]
	}
	hash := sha256.New()
	d := append(append([]byte(f.Key), raw...), body...)
	hash.Write(d)
	return hex.EncodeToString(hash.Sum(nil))
}

// get the header
func (f *Fcm) getHeader() http.Header {
	return http.Header{
		"appId":      []string{f.AppID},
		"bizId":      []string{f.BizID},
		"timestamps": []string{strconv.FormatInt(time.Now().Unix()*1000, 10)},
	}
}

// set client
func (f *Fcm) SetClient(transport http.RoundTripper, timeout time.Duration) {
	f.client = http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// request
func (f *Fcm) request(method, uri string, b interface{}, header http.Header) (Result, error) {
	jsonBody, err := json.Marshal(b)
	if err != nil {
		return Result{}, err
	}
	var v interface{}
	json.Unmarshal(jsonBody, &v)
	requestData, err := f.makeBody(jsonBody)
	if err != nil {
		return Result{}, err
	}
	var raw string
	switch b.(type) {
	case *[]Behavior:
		raw = `{"data":"` + strings.TrimRight(requestData, "=") + `"}`
		break
	case *Query:
	case *Check:
	default:
		raw = strings.TrimRight(requestData, "=")
	}
	header["sign"] = []string{f.makeSign(header.Clone(), raw, nil)}
	req, err := http.NewRequest(method, uri, bytes.NewReader([]byte(raw)))
	if err != nil {
		return Result{}, err
	}
	req.Header = header
	response, err := f.client.Do(req)

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || err != nil {
		return Result{}, errors.New(response.Status)
	}
	body, _ := ioutil.ReadAll(response.Body)
	responseData := Result{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return Result{}, err
	}
	if responseData.ErrCode != 0 || responseData.ErrMsg != "ok" {
		return responseData, errors.New(responseData.ErrMsg)
	}
	return responseData, nil
}
