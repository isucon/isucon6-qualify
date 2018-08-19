package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Songmu/strrand"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
)

const (
	sessionName   = "isuda_session"
	sessionSecret = "tonymoris"
)

var (
	isupamEndpoint string

	baseUrl *url.URL
	db      *sql.DB
	re      *render.Render
	store   *sessions.CookieStore

	errInvalidUser = errors.New("Invalid User")
)

func setName(w http.ResponseWriter, r *http.Request) error {
	session := getSession(w, r)
	userID, ok := session.Values["user_id"]
	if !ok {
		return nil
	}
	setContext(r, "user_id", userID)
	row := db.QueryRow(`SELECT name FROM user WHERE id = ?`, userID)
	user := User{}
	err := row.Scan(&user.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return errInvalidUser
		}
		panicIf(err)
	}
	setContext(r, "user_name", user.Name)
	return nil
}

func authenticate(w http.ResponseWriter, r *http.Request) error {
	if u := getContext(r, "user_id"); u != nil {
		return nil
	}
	return errInvalidUser
}

func initializeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := db.Exec(`DELETE FROM entry WHERE id > 7101`)
	panicIf(err)

	// keywordsの初期化
	//replaceKeyword()

	_, err = db.Exec("TRUNCATE star")
	panicIf(err)

	re.JSON(w, http.StatusOK, map[string]string{"result": "ok"})
}

func topHandler(w http.ResponseWriter, r *http.Request) {
	//　めちゃくちゃ重い
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	perPage := 10 // TODO: 色んなとこで書きそうだからglobalにしよ
	p := r.URL.Query().Get("page")
	if p == "" {
		p = "1"
	}
	page, _ := strconv.Atoi(p)

	// TODO:
	// - TopPageを保存しておくDBを作る
	// - 最新のid 10個が変わってなければ、既存のhtmlify済みのコンテンツを返す
	// - 変わっていれば、新しく作り直してhtmlify済みのコンテンツを保存 & 返す
	// TABLE id, entry_ids, contents

	// TODO:
	// html, keyword
	rows, err := db.Query(fmt.Sprintf(
		// "SELECT * FROM entry ORDER BY updated_at DESC LIMIT %d OFFSET %d",
		"SELECT keyword, description FROM entry ORDER BY updated_at DESC LIMIT %d OFFSET %d",
		perPage, perPage*(page-1),
	))
	if err != nil && err != sql.ErrNoRows {
		panicIf(err)
	}
	entries := make([]*Entry, 0, 10)
	keywords := getKeywords()
	for rows.Next() {
		e := Entry{}
		// err := rows.Scan(&e.ID, &e.AuthorID, &e.Keyword, &e.Description, &e.UpdatedAt, &e.CreatedAt)
		err := rows.Scan(&e.Keyword, &e.Description)
		panicIf(err)
		e.Html = htmlify(w, r, e.Description, keywords)
		e.Stars = loadStars(e.Keyword)
		entries = append(entries, &e)
	}
	rows.Close()

	var totalEntries int
	row := db.QueryRow(`SELECT COUNT(*) FROM entry`)
	err = row.Scan(&totalEntries)
	if err != nil && err != sql.ErrNoRows {
		panicIf(err)
	}

	lastPage := int(math.Ceil(float64(totalEntries) / float64(perPage)))
	pages := make([]int, 0, 10)
	start := int(math.Max(float64(1), float64(page-5)))
	end := int(math.Min(float64(lastPage), float64(page+5)))
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	re.HTML(w, http.StatusOK, "index", struct {
		Context  context.Context
		Entries  []*Entry
		Page     int
		LastPage int
		Pages    []int
	}{
		r.Context(), entries, page, lastPage, pages,
	})
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	notFound(w)
}

func keywordPostHandler(w http.ResponseWriter, r *http.Request) {
	// スコアに一番重要な関数
	// 3000ms以内に必ず返すこと

	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}
	if err := authenticate(w, r); err != nil {
		forbidden(w)
		return
	}

	keyword := r.FormValue("keyword")
	if keyword == "" {
		badRequest(w)
		return
	}
	userID := getContext(r, "user_id").(int)
	description := r.FormValue("description")

	if isSpamContents(description) || isSpamContents(keyword) {
		http.Error(w, "SPAM!", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`
		INSERT INTO entry (author_id, keyword, description, created_at, updated_at, keyword_length)
		VALUES (?, ?, ?, NOW(), NOW(), CHARACTER_LENGTH(keyword))
		ON DUPLICATE KEY UPDATE
		author_id = ?, keyword = ?, description = ?, updated_at = NOW(), keyword_length = CHARACTER_LENGTH(keyword)
	`, userID, keyword, description, userID, keyword, description)
	panicIf(err)

	// Update keyword
	//entry := Entry{}
	//row := db.QueryRow(`SELECT keyword, keyword_length FROM entry WHERE keyword = ?`, keyword)
	//row.Scan(entry.Keyword, entry.KeywordLength)
	//updateKeyword(&entry)

	http.Redirect(w, r, "/", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	re.HTML(w, http.StatusOK, "authenticate", struct {
		Context context.Context
		Action  string
	}{
		r.Context(), "login",
	})
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	// TODO: cocodrips nameにindex貼ってあるか確認
	// TODO: cocodrips Id, Password, Saltのみ取得すれば良いので、Name/CreatedAtは取得しないでも良い
	row := db.QueryRow(`SELECT id, name, salt, password FROM user WHERE name = ?`, name)
	user := User{}
	err := row.Scan(&user.ID, &user.Name, &user.Salt, &user.Password)
	if err == sql.ErrNoRows || user.Password != fmt.Sprintf("%x", sha1.Sum([]byte(user.Salt+r.FormValue("password")))) {
		forbidden(w)
		return
	}
	panicIf(err)
	session := getSession(w, r)
	session.Values["user_id"] = user.ID
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(w, r)
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	re.HTML(w, http.StatusOK, "authenticate", struct {
		Context context.Context
		Action  string
	}{
		r.Context(), "register",
	})
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pw := r.FormValue("password")
	if name == "" || pw == "" {
		badRequest(w)
		return
	}
	userID := register(name, pw)
	session := getSession(w, r)
	session.Values["user_id"] = userID
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func register(user string, pass string) int64 {
	salt, err := strrand.RandomString(`....................`)
	panicIf(err)
	res, err := db.Exec(`INSERT INTO user (name, salt, password, created_at) VALUES (?, ?, ?, NOW())`,
		user, salt, fmt.Sprintf("%x", sha1.Sum([]byte(salt+pass))))
	panicIf(err)
	lastInsertID, _ := res.LastInsertId()
	return lastInsertID
}

func keywordByKeywordHandler(w http.ResponseWriter, r *http.Request) {
	// 個別ページ
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	keyword, err := url.QueryUnescape(mux.Vars(r)["keyword"])
	if err != nil {
		return
	}

	row := db.QueryRow(`SELECT keyword, description FROM entry WHERE keyword = ?`, keyword)
	e := Entry{}
	err = row.Scan(&e.Keyword, &e.Description)
	if err == sql.ErrNoRows {
		notFound(w)
		return
	}

	keywords := getKeywords()
	e.Html = htmlify(w, r, e.Description, keywords)
	e.Stars = loadStars(e.Keyword)

	// Html, Keyword, StarsのみでOK
	re.HTML(w, http.StatusOK, "keyword", struct {
		Context context.Context
		Entry   Entry
	}{
		r.Context(), e,
	})
}

func keywordByKeywordDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}
	if err := authenticate(w, r); err != nil {
		forbidden(w)
		return
	}

	keyword := mux.Vars(r)["keyword"]
	if keyword == "" {
		badRequest(w)
		return
	}
	if r.FormValue("delete") == "" {
		badRequest(w)
		return
	}
	row := db.QueryRow(`SELECT id FROM entry WHERE keyword = ?`, keyword)
	e := Entry{}
	err := row.Scan(&e.ID)
	if err == sql.ErrNoRows {
		notFound(w)
		return
	}

	// TODO ここにも書く
	_, err = db.Exec(`DELETE FROM entry WHERE keyword = ?`, keyword)
	panicIf(err)

	http.Redirect(w, r, "/", http.StatusFound)
}

func getKeywords() []string {
	keywords := make([]string, 0, 500)
	rows, err := db.Query(`SELECT keyword FROM entry ORDER BY keyword_length DESC`)
	panicIf(err)

	for rows.Next() {
		s := ""
		err := rows.Scan(&s)
		panicIf(err)
		keywords = append(keywords, s)
	}

	return keywords
}

func htmlify(w http.ResponseWriter, r *http.Request, content string, keywords []string) string {
	if content == "" {
		return ""
	}

	kw2sha := make(map[string]string)
	n := len(keywords)
	for i := 0; i < n; i++ {
		kw := keywords[n - i - 1]

		hasKeyword := strings.Contains(content, kw)

		if !hasKeyword {
			continue
		}

		fmt.Println(kw, hasKeyword)

		kw2sha[kw] = "isuda_" + fmt.Sprintf("%x", sha1.Sum([]byte(kw)))
		content = strings.Replace(content, kw, kw2sha[kw], -1)
	}

	for i := 0; i < n; i++ {
		kw := keywords[n - i - 1]
		if hash, ok := kw2sha[kw]; ok {
			content = strings.Replace(content, kw, kw2sha[kw], -1)
			u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(kw))
			panicIf(err)
			link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(kw))
			content = strings.Replace(content, hash, link, -1)
		}

	}

	return strings.Replace(content, "\n", "<br />\n", -1)
}

func loadStars(keyword string) []Star {
	rows, err := db.Query(`SELECT * FROM star WHERE keyword = ?`, keyword)
	if err != nil && err != sql.ErrNoRows {
		panicIf(err)
	}

	stars := make([]Star, 0, 10)
	for rows.Next() {
		s := Star{}
		err := rows.Scan(&s.ID, &s.Keyword, &s.UserName, &s.CreatedAt)
		panicIf(err)
		stars = append(stars, s)
	}

	rows.Close()

	return stars
}

func starsHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.FormValue("keyword")
	stars := loadStars(keyword)

	re.JSON(w, http.StatusOK, map[string][]Star{
		"result": stars,
	})
}

func starsPostHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.FormValue("keyword")

	var existing int
	err := db.QueryRow(`SELECT 1 FROM entry WHERE keyword = ? LIMIT 1`, keyword).Scan(&existing)
	if err != nil {
		if err == sql.ErrNoRows {
			notFound(w)
			return
		}
		panicIf(err)
	}

	user := r.FormValue("user")
	_, err = db.Exec(`INSERT INTO star (keyword, user_name, created_at) VALUES (?, ?, NOW())`, keyword, user)
	panicIf(err)

	re.JSON(w, http.StatusOK, map[string]string{"result": "ok"})
}

func isSpamContents(content string) bool {
	// TODO: あやしい　要チェック -> ispamはbinaryなのでどうしようもないぽい
	v := url.Values{}
	v.Set("content", content)
	resp, err := http.PostForm(isupamEndpoint, v)
	panicIf(err)
	defer resp.Body.Close()

	var data struct {
		Valid bool `json:valid`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	panicIf(err)
	return !data.Valid
}

func getContext(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

func setContext(r *http.Request, key, val interface{}) {
	if val == nil {
		return
	}

	r2 := r.WithContext(context.WithValue(r.Context(), key, val))
	*r = *r2
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, _ := store.Get(r, sessionName)
	return session
}

func main() {
	host := os.Getenv("ISUDA_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	portstr := os.Getenv("ISUDA_DB_PORT")
	if portstr == "" {
		portstr = "3306"
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		log.Fatalf("Failed to read DB port number from an environment variable ISUDA_DB_PORT.\nError: %s", err.Error())
	}
	user := os.Getenv("ISUDA_DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("ISUDA_DB_PASSWORD")
	dbname := os.Getenv("ISUDA_DB_NAME")
	if dbname == "" {
		dbname = "isuda"
	}

	db, err = sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?loc=Local&parseTime=true&charset=utf8mb4",
		user, password, host, port, dbname,
	))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s.", err.Error())
	}
	db.Exec("SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'")
	db.Exec("SET NAMES utf8mb4")

	isupamEndpoint = os.Getenv("ISUPAM_ORIGIN")
	if isupamEndpoint == "" {
		isupamEndpoint = "http://localhost:5050"
	}

	store = sessions.NewCookieStore([]byte(sessionSecret))

	re = render.New(render.Options{
		Directory: "views",
		Funcs: []template.FuncMap{
			{
				"url_for": func(path string) string {
					return baseUrl.String() + path
				},
				"title": func(s string) string {
					return strings.Title(s)
				},
				"raw": func(text string) template.HTML {
					return template.HTML(text)
				},
				"add": func(a, b int) int { return a + b },
				"sub": func(a, b int) int { return a - b },
				"entry_with_ctx": func(entry Entry, ctx context.Context) *EntryWithCtx {
					return &EntryWithCtx{Context: ctx, Entry: entry}
				},
			},
		},
	})

	r := mux.NewRouter()
	r.UseEncodedPath()
	r.HandleFunc("/", myHandler(topHandler))
	r.HandleFunc("/initialize", myHandler(initializeHandler)).Methods("GET")
	// TODO: nginxで返す
	r.HandleFunc("/robots.txt", myHandler(robotsHandler))
	r.HandleFunc("/keyword", myHandler(keywordPostHandler)).Methods("POST")

	l := r.PathPrefix("/login").Subrouter()
	l.Methods("GET").HandlerFunc(myHandler(loginHandler))
	l.Methods("POST").HandlerFunc(myHandler(loginPostHandler))
	r.HandleFunc("/logout", myHandler(logoutHandler))

	g := r.PathPrefix("/register").Subrouter()
	g.Methods("GET").HandlerFunc(myHandler(registerHandler))
	g.Methods("POST").HandlerFunc(myHandler(registerPostHandler))

	k := r.PathPrefix("/keyword/{keyword}").Subrouter()
	k.Methods("GET").HandlerFunc(myHandler(keywordByKeywordHandler))
	k.Methods("POST").HandlerFunc(myHandler(keywordByKeywordDeleteHandler))

	s := r.PathPrefix("/stars").Subrouter()
	s.Methods("GET").HandlerFunc(myHandler(starsHandler))
	s.Methods("POST").HandlerFunc(myHandler(starsPostHandler))

	// TODO: /publicはnginxで返す
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	log.Fatal(http.ListenAndServe(":5000", r))
}
