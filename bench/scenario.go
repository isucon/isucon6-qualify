package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/isucon/isucon6-qualify/bench/checker"
	"github.com/isucon/isucon6-qualify/bench/util"
)

func checkHTML(f func(*goquery.Document) error) func(io.Reader) error {
	return func(r io.Reader) error {
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			return fmt.Errorf("ページのHTMLがパースできませんでした")
		}
		return f(doc)
	}
}

func extractArticles(doc *goquery.Document) (keywords []string) {
	doc.Find("article h1 a").Each(func(_ int, selection *goquery.Selection) {
		keywords = append(keywords, selection.Text())
	})
	return
}

var indexReg = regexp.MustCompile(`^/$`)

// インデックスにリクエストしてページャを最大10ページ辿る
// WaitAfterTimeout秒たったら問答無用で打ち切る
func indexMoreAndMoreScenario(s *checker.Session, bd *benchdata) {
	allKw := bd.allWords

	start := time.Now()
	seen := make(map[string]struct{})

	keywordPerPageChecker := func(r io.Reader) error {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("レスポンスbodyの取得に失敗しました")
		}
		contentLength := int64(len(b))
		r2 := bytes.NewReader(b)
		doc, err := goquery.NewDocumentFromReader(r2)
		if err != nil {
			return fmt.Errorf("ページのHTMLがパースできませんでした")
		}

		keywords := extractArticles(doc)
		if len(keywords) < PostsPerPage {
			return fmt.Errorf("1ページに表示されるキーワードが足りません")
		}
		var expectedLength int64 = 3000
		for _, k := range keywords {
			kw, ok := allKw[k]
			if !ok {
				return fmt.Errorf("おかしなキーワード(%s)が存在しています", k)
			}
			for _, kwName := range kw.Links {
				l := doc.Find(fmt.Sprintf(`article div a[href$="/keyword/%s"]`, pathURIEscape(kwName))).Length()
				if l < 1 {
					return fmt.Errorf("%s に %s へのリンクがありません", k, kwName)
				}
			}
			if _, ok := seen[k]; ok {
				return fmt.Errorf("%s は既に表示されています", k)
			}
			seen[k] = struct{}{}
			expectedLength += kw.Length
		}
		if expectedLength > contentLength {
			return fmt.Errorf("ページサイズが小さすぎます")
		}
		return nil
	}

	index := checker.NewAction("GET", "/")
	index.Description = "インデックスページが表示できること"
	index.CheckFunc = keywordPerPageChecker
	err := index.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)

	offset := util.RandomNumber(10) + 5 // 10は適当。URLをバラけさせるため
	for i := 0; i < 10; i++ {           // 10ページ辿る
		posts := checker.NewAction("GET", fmt.Sprintf("/?page=%d", offset+i*10))
		posts.Description = "深いページが表示できること"
		posts.CheckFunc = keywordPerPageChecker
		err := posts.Play(s)
		if err != nil {
			return
		}
		loadAssets(s)

		if time.Now().Sub(start) > WaitAfterTimeout {
			break
		}
	}
}

// インデックスページを5回表示するだけ（負荷かける用）
// WaitAfterTimeout秒たったら問答無用で打ち切る
func loadIndexScenario(s *checker.Session, _ *benchdata) {
	var keywords []string
	start := time.Now()

	keywordPerPageChecker := checkHTML(func(doc *goquery.Document) error {
		keywords = extractArticles(doc)
		if len(keywords) < PostsPerPage {
			return fmt.Errorf("1ページに表示されるキーワードの数が足りません")
		}
		return nil
	})

	index := checker.NewAction("GET", "/")
	index.Description = "インデックスページが表示できること"
	index.CheckFunc = keywordPerPageChecker
	err := index.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)

	for i := 0; i < 4; i++ {
		// あとの4回はDOMをパースしない。トップページをキャッシュして超高速に返されたとき対策
		index := checker.NewAction("GET", "/")
		index.Description = "インデックスページが表示できること"
		err := index.Play(s)
		if err != nil {
			return
		}

		loadAssets(s)
		if time.Now().Sub(start) > WaitAfterTimeout {
			break
		}
	}
}

// インデックスページを表示してStarをつける
func loadIndexWithStarScenario(s *checker.Session, bd *benchdata) {
	me := bd.randomUser()
	login := checker.NewAction(http.MethodPost, "/login")
	login.ExpectedStatusCode = http.StatusFound
	login.ExpectedLocation = indexReg
	login.Description = "ログインできる"
	login.PostData = map[string]string{
		"name":     me.Name,
		"password": me.Password,
	}
	err := login.Play(s)
	if err != nil {
		return
	}

	var keywords []string
	keywordPerPageChecker := checkHTML(func(doc *goquery.Document) error {
		keywords = extractArticles(doc)
		if len(keywords) < PostsPerPage {
			return fmt.Errorf("1ページに表示されるキーワードの数が足りません")
		}
		return nil
	})

	index := checker.NewAction("GET", "/")
	index.Description = "インデックスページが表示できること"
	index.CheckFunc = keywordPerPageChecker
	err = index.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)

	for _, k := range keywords {
		star := checker.NewAction(
			http.MethodPost, fmt.Sprintf("/stars?user=%s&keyword=%s", me.Name, url.QueryEscape(k)))
		star.Description = "starが投稿できる"
		// TODO ExpectedHeaders 未実装…
		star.ExpectedHeaders = map[string]string{
			"Content-Type": "application/json; charset=UTF-8",
		}
		err := star.Play(s)
		if err != nil {
			return
		}
	}

	index2 := checker.NewAction("GET", "/")
	index2.Description = "インデックスページにstarが付いていること"
	index2.CheckFunc = checkHTML(func(doc *goquery.Document) error {
		len := doc.Find(fmt.Sprintf(`img[title="%s"]`, me.Name)).Length()
		if len < 2 {
			return fmt.Errorf("starがついていません")
		}
		return nil
	})
	err = index2.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)

	ngKw := bd.randomNGWord()
	ngStar := checker.NewAction(
		http.MethodPost, fmt.Sprintf("/stars?user=%s&keyword=%s", me.Name, url.QueryEscape(ngKw.Keyword)))
	ngStar.Description = "starが投稿できない"
	ngStar.ExpectedStatusCode = http.StatusNotFound
	err = ngStar.Play(s)
	if err != nil {
		return
	}
}

// 適当なユーザー名でログインしようとする
// ログインできないことをチェック
func cannotLoginNonexistentUserScenario(s *checker.Session, _ *benchdata) {
	fakeName := util.RandomLUNStr(util.RandomNumber(15) + 10)
	fakeUser := map[string]string{
		"name":     fakeName,
		"password": fakeName,
	}

	login := checker.NewAction(http.MethodPost, "/login")
	login.Description = "存在しないユーザー名でログインできないこと"
	login.ExpectedStatusCode = 403
	login.PostData = fakeUser

	login.Play(s)
}

// 誤ったパスワードでログインできない
func cannotLoginWrongPasswordScenario(s *checker.Session, bd *benchdata) {
	me := bd.randomUser()

	fakeUser := map[string]string{
		"name":     me.Name,
		"password": util.RandomLUNStr(util.RandomNumber(15) + 10),
	}

	login := checker.NewAction(http.MethodPost, "/login")
	login.Description = "間違ったパスワードでログインできないこと"
	login.ExpectedStatusCode = 403
	login.PostData = fakeUser

	login.Play(s)
}

// ログインすると右上にアカウント名が出て、ログインしないとアカウント名が出ない
// 画像のキャッシュにSet-Cookieを含んでいた場合、/にアカウント名が含まれる
func loginScenario(s *checker.Session, bd *benchdata) {
	me := bd.randomUser()

	login := checker.NewAction(http.MethodPost, "/login")
	login.ExpectedStatusCode = http.StatusFound
	login.ExpectedLocation = indexReg
	login.Description = "ログインするとユーザー名が表示されること"
	login.PostData = map[string]string{
		"name":     me.Name,
		"password": me.Password,
	}
	err := login.Play(s)
	if err != nil {
		return
	}

	index := checker.NewAction(http.MethodGet, "/")
	index.CheckFunc = checkHTML(func(doc *goquery.Document) error {
		name := doc.Find(`.isu-account-name`).Text()
		if name == "" {
			return fmt.Errorf("ユーザー名が表示されていません")
		} else if name != me.Name {
			return fmt.Errorf("表示されているユーザー名が正しくありません")
		}
		return nil
	})
	err = index.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)

	logout := checker.NewAction("GET", "/logout")
	logout.ExpectedStatusCode = http.StatusFound
	logout.ExpectedLocation = indexReg
	err = logout.Play(s)
	if err != nil {
		return
	}

	loggedOutIndex := checker.NewAction("GET", "/")
	loggedOutIndex.Description = "ログアウトするとユーザー名が表示されないこと"
	loggedOutIndex.CheckFunc = checkHTML(func(doc *goquery.Document) error {
		name := doc.Find(`.isu-account-name`).Text()
		if name != "" {
			return fmt.Errorf("ログアウトしてもユーザー名が表示されています")
		}
		return nil
	})
	err = loggedOutIndex.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)
}

// NGwワードはスパム判定
func ngwordScenario(s *checker.Session, bd *benchdata) {
	me := bd.randomUser()
	kw := bd.randomNGWord()

	login := checker.NewAction(http.MethodPost, "/login")
	login.ExpectedStatusCode = http.StatusFound
	login.ExpectedLocation = indexReg
	login.Description = "ログインできる"
	login.PostData = map[string]string{
		"name":     me.Name,
		"password": me.Password,
	}
	err := login.Play(s)
	if err != nil {
		return
	}

	kwAct := checker.NewAction(http.MethodPost, "/keyword")
	kwAct.Description = "spam判定される"
	kwAct.ExpectedStatusCode = 400
	kwAct.PostData = map[string]string{
		"keyword":     kw.Keyword,
		"description": kw.Description,
	}
	kwAct.Play(s)
}

// キーワード
func keywordScenario(s *checker.Session, bd *benchdata) {
	kw := bd.randomWord()

	kwAct := checker.NewAction("GET", "/keyword/"+pathURIEscape(kw.Keyword))
	kwAct.Description = "キーワードが正しく表示できる"
	kwAct.CheckFunc = checkHTML(func(doc *goquery.Document) error {
		for _, k := range kw.Links {
			l := doc.Find(fmt.Sprintf(`article div a[href$="/keyword/%s"]`, pathURIEscape(k))).Length()
			if l < 1 {
				return fmt.Errorf("keyword: %q に %q へのリンクがありません", kw.Keyword, k)
			}
		}
		return nil
	})
	err := kwAct.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)
}

func postKeywordScenario(s *checker.Session, bd *benchdata) {
	me := bd.randomUser()
	kw := bd.newWord()
	start := time.Now()

	login := checker.NewAction(http.MethodPost, "/login")
	login.ExpectedStatusCode = http.StatusFound
	login.ExpectedLocation = indexReg
	login.Description = "ログインできる"
	login.PostData = map[string]string{
		"name":     me.Name,
		"password": me.Password,
	}
	err := login.Play(s)
	if err != nil {
		return
	}

	kwAct := checker.NewAction(http.MethodPost, "/keyword")
	kwAct.ExpectedStatusCode = http.StatusFound
	kwAct.ExpectedLocation = indexReg
	kwAct.Description = "キーワード投稿できる"
	kwAct.PostData = map[string]string{
		"keyword":     kw.Keyword,
		"description": kw.Description,
	}
	err = kwAct.Play(s)
	if err != nil {
		return
	}

	kwAct = checker.NewAction("GET", "/keyword/"+pathURIEscape(kw.Keyword))
	kwAct.Description = "キーワードが正しく表示できる"
	kwAct.Play(s)
	err = kwAct.Play(s)
	if err != nil {
		return
	}
	if time.Now().Sub(start) > WaitAfterTimeout {
		return
	}
	loadAssets(s)

	index := checker.NewAction("GET", "/")
	index.Description = "インデックスページにキーワードが反映されている"
	index.CheckFunc = checkHTML(func(doc *goquery.Document) error {
		keywords := extractArticles(doc)
		if len(keywords) < PostsPerPage {
			return fmt.Errorf("1ページに表示されるキーワードの数が足りません")
		}
		for _, v := range keywords {
			if v == kw.Keyword {
				return nil
			}
		}
		return fmt.Errorf("トップページにキーワードが反映されていません: %s", kw.Keyword)
	})
	err = index.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)

	for _, k := range kw.Backlinks {
		kwAct = checker.NewAction("GET", "/keyword/"+pathURIEscape(k))
		kwAct.Description = "バックリンクからちゃんとリンクが張られている"
		kwAct.CheckFunc = checkHTML(func(doc *goquery.Document) error {
			l := doc.Find(fmt.Sprintf(`article div a[href$="/keyword/%s"]`, pathURIEscape(kw.Keyword))).Length()
			if l < 1 {
				return fmt.Errorf("keyword: %q に %q からのリンクがありません", kw.Keyword, k)
			}
			return nil
		})
		err := kwAct.Play(s)
		if err != nil {
			return
		}
		if time.Now().Sub(start) > WaitAfterTimeout {
			return
		}
		loadAssets(s)
	}
}

func updateKeywordScenario(s *checker.Session, bd *benchdata) {
	me := bd.randomUser()
	kw := bd.randomWord()

	if kw == nil {
		time.Sleep(5 * time.Second)
		return
	}
	start := time.Now()

	login := checker.NewAction(http.MethodPost, "/login")
	login.ExpectedStatusCode = http.StatusFound
	login.ExpectedLocation = indexReg
	login.Description = "ログインできる"
	login.PostData = map[string]string{
		"name":     me.Name,
		"password": me.Password,
	}
	err := login.Play(s)
	if err != nil {
		return
	}

	kwAct := checker.NewAction(http.MethodPost, "/keyword")
	kwAct.ExpectedStatusCode = http.StatusFound
	kwAct.ExpectedLocation = indexReg
	kwAct.Description = "キーワード投稿できる"
	kwAct.PostData = map[string]string{
		"keyword":     kw.Keyword,
		"description": kw.Description + util.RandomLUNStr(3),
	}
	err = kwAct.Play(s)
	if err != nil {
		return
	}

	kwAct = checker.NewAction("GET", "/keyword/"+pathURIEscape(kw.Keyword))
	kwAct.Description = "キーワードが正しく表示できる"
	kwAct.Play(s)
	err = kwAct.Play(s)
	if err != nil {
		return
	}
	if time.Now().Sub(start) > WaitAfterTimeout {
		return
	}
	loadAssets(s)

	index := checker.NewAction("GET", "/")
	index.Description = "インデックスページにキーワードが反映されている"
	index.CheckFunc = checkHTML(func(doc *goquery.Document) error {
		keywords := extractArticles(doc)
		if len(keywords) < PostsPerPage {
			return fmt.Errorf("1ページに表示されるキーワードの数が足りません")
		}
		for _, v := range keywords {
			if v == kw.Keyword {
				return nil
			}
		}
		return fmt.Errorf("トップページにキーワード更新が反映されていません: %s", kw.Keyword)
	})
	err = index.Play(s)
	if err != nil {
		return
	}
	loadAssets(s)

	for _, k := range kw.Backlinks {
		kwAct = checker.NewAction("GET", "/keyword/"+pathURIEscape(k))
		kwAct.Description = "バックリンクからちゃんとリンクが張られている"
		kwAct.CheckFunc = checkHTML(func(doc *goquery.Document) error {
			l := doc.Find(fmt.Sprintf(`article div a[href$="/keyword/%s"]`, pathURIEscape(kw.Keyword))).Length()
			if l < 1 {
				return fmt.Errorf("keyword: %q に %q からのリンクがありません", kw.Keyword, k)
			}
			return nil
		})
		err := kwAct.Play(s)
		if err != nil {
			return
		}
		if time.Now().Sub(start) > WaitAfterTimeout {
			return
		}
		loadAssets(s)
	}
}

// url.QueryEscape だと whitespaceが '+' に置換される挙動だったのでパス用に '%20' にするやつ
func pathURIEscape(s string) string {
	return (&url.URL{Path: s}).String()
}
