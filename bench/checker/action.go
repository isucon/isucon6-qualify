package checker

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/isucon/isucon6-qualify/bench/cache"
	"github.com/isucon/isucon6-qualify/bench/score"
)

type Action struct {
	Method string
	Path   string

	PostData           map[string]string
	Headers            map[string]string
	ExpectedStatusCode int
	ExpectedLocation   *regexp.Regexp
	ExpectedHeaders    map[string]string
	ExpectedHTML       map[string]string

	Description string

	CheckFunc func(body io.Reader) error
}

const (
	successAssetScore = 1
	successGetScore   = 5
	successPostScore  = 10

	failErrorScore     = 50
	failExceptionScore = 100
	failDelayPostScore = 200
)

type Asset struct {
	Path string
	MD5  string
	Type string
}

func NewAction(method, path string) *Action {
	return &Action{
		Method:             method,
		Path:               path,
		ExpectedStatusCode: http.StatusOK,
	}
}

func (a *Action) Play(s *Session) error {
	formData := url.Values{}
	for key, val := range a.PostData {
		formData.Set(key, val)
	}

	buf := bytes.NewBufferString(formData.Encode())
	req, err := s.NewRequest(a.Method, a.Path, buf)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return s.Fail(failExceptionScore, req, fmt.Errorf("リクエストに失敗しました (主催者に連絡してください)"))
	}

	for key, val := range a.Headers {
		req.Header.Add(key, val)
	}

	if req.Method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	res, err := s.SendRequest(req)

	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Timeout() {
				var failScore int64 = failExceptionScore
				if a.Method == http.MethodPost {
					failScore = failDelayPostScore
				}
				score.GetInstance().IncrTimeouts()
				return s.Fail(failScore, req, fmt.Errorf("リクエストがタイムアウトしました"))
			}
		case *url.Error:
			if e.Err == RedirectAttemptedError {
				break
			}
			// fallthroughしようとしたら type switchではできないとのこと
			fmt.Fprintln(os.Stderr, err)
			return s.Fail(failExceptionScore, req, fmt.Errorf("リクエストに失敗しました"))
		default:
			fmt.Fprintln(os.Stderr, err)
			return s.Fail(failExceptionScore, req, fmt.Errorf("リクエストに失敗しました"))
		}
	}
	if res == nil {
		return s.Fail(failErrorScore, req, fmt.Errorf("レスポンスが不正です"))
	}
	defer res.Body.Close()

	if res.StatusCode != a.ExpectedStatusCode {
		return s.Fail(failErrorScore, res.Request, fmt.Errorf("Response code should be %d, got %d, data: %s", a.ExpectedStatusCode, res.StatusCode, a.PostData["keyword"]))
	}

	if a.ExpectedLocation != nil {
		l := res.Header["Location"]
		if len(l) != 1 {
			return s.Fail(
				failErrorScore,
				res.Request,
				fmt.Errorf("リダイレクトURLが適切に設定されていません"))
		}
		u, err := url.Parse(l[0])
		if err != nil || !a.ExpectedLocation.MatchString(u.Path) {
			return s.Fail(
				failErrorScore,
				res.Request,
				fmt.Errorf(
					"リダイレクト先URLが正しくありません: expected '%s', got '%s'",
					a.ExpectedLocation, l[0],
				))
		}
	}

	if a.CheckFunc != nil {
		err := a.CheckFunc(res.Body)
		if err != nil {
			return s.Fail(
				failErrorScore,
				res.Request,
				err,
			)
		}
	}

	var successScore int64 = successGetScore
	if a.Method == http.MethodPost {
		successScore = successPostScore
	}
	s.Success(successScore)

	return nil
}

type AssetAction struct {
	*Action
	Asset *Asset
}

func NewAssetAction(path string, asset *Asset) *AssetAction {
	return &AssetAction{
		Asset: asset,
		Action: &Action{
			Method:             http.MethodGet,
			Path:               path,
			ExpectedStatusCode: http.StatusOK,
		},
	}
}

func (a *AssetAction) Play(s *Session) error {
	formData := url.Values{}
	for key, val := range a.PostData {
		formData.Set(key, val)
	}

	buf := bytes.NewBufferString(formData.Encode())
	req, err := s.NewRequest(a.Method, a.Path, buf)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return s.Fail(failExceptionScore, req, fmt.Errorf("リクエストに失敗しました (主催者に連絡してください)"))
	}

	for key, val := range a.Headers {
		req.Header.Add(key, val)
	}

	urlCache, cacheFound := cache.GetInstance().Get(a.Path)
	if cacheFound {
		urlCache.Apply(req)
	}

	res, err := s.SendRequest(req)

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return s.Fail(failExceptionScore, req, fmt.Errorf("リクエストがタイムアウトしました"))
		}
		fmt.Fprintln(os.Stderr, err)
		return s.Fail(failExceptionScore, req, fmt.Errorf("リクエストに失敗しました"))
	}

	// 2回ioutil.ReadAllを呼ぶとおかしくなる
	uc, md5 := cache.NewURLCache(res)
	if uc != nil {
		cache.GetInstance().Set(a.Path, uc)
		if res.StatusCode == http.StatusOK && a.Asset.MD5 == "" {
			a.Asset.MD5 = md5
		}
	} else if a.Asset.MD5 == "" {
		a.Asset.MD5 = md5
	}

	success := false

	// キャッシュが有効でかつStatusNotModifiedのときは成功
	if cacheFound && res.StatusCode == http.StatusNotModified {
		success = true
	}

	if res.StatusCode == http.StatusOK &&
		((uc == nil && md5 == a.Asset.MD5) || uc != nil) {
		success = true
	}

	defer res.Body.Close()

	if !success {
		return s.Fail(
			failErrorScore,
			res.Request,
			fmt.Errorf("静的ファイルが正しくありません"),
		)
	}

	s.Success(successAssetScore)

	return nil
}
