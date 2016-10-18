package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isucon/isucon6-qualify/bench/score"
	"github.com/isucon/isucon6-qualify/portal/job"
)

var s *httptest.Server

func TestMain(m *testing.M) {
	*startsAtHour = -1
	*endsAtHour = -1

	flag.Parse()
	err := initWeb()
	if err != nil {
		log.Fatal(err)
	}

	s = httptest.NewServer(buildMux())
	n := m.Run()
	s.Close()
	os.Exit(n)
}

type testHTTPClient struct {
	*http.Client
	*testing.T
}

func (c *testHTTPClient) Must(resp *http.Response, err error) *http.Response {
	require.NoError(c.T, err)
	return resp
}

func newTestClient(t *testing.T) *testHTTPClient {
	jar, _ := cookiejar.New(nil)
	return &testHTTPClient{
		Client: &http.Client{Jar: jar},
		T:      t,
	}
}

func TestLogin(t *testing.T) {
	resp, err := http.Get(s.URL)
	require.NoError(t, err)
	require.Equal(t, "/login", resp.Request.URL.Path)

	jar, _ := cookiejar.New(nil)
	cli := &http.Client{Jar: jar}

	resp, err = cli.PostForm(s.URL+"/login", url.Values{"team_id": {"1200"}, "password": {"dummy-pass-%d200"}})
	require.NoError(t, err)
	require.Equal(t, "/", resp.Request.URL.Path)
}

func readAll(r io.Reader) string {
	b, _ := ioutil.ReadAll(r)
	return string(b)
}

func benchGetJob(bench *testHTTPClient) *job.Job {
	resp := bench.Must(bench.Post(s.URL+"/top4aew4fe9yeehu/job/new", "", nil))
	if !assert.Equal(bench.T, http.StatusOK, resp.StatusCode) {
		return nil
	}

	var j job.Job
	err := json.NewDecoder(resp.Body).Decode(&j)
	require.NoError(bench.T, err)

	return &j
}

func benchPostResult(bench *testHTTPClient, j *job.Job, output *score.Output) {
	time.Sleep(1 * time.Second)

	result := job.Result{
		Job:    j,
		Output: output,
	}
	resultJSON, err := json.Marshal(result)
	require.NoError(bench.T, err)

	resp := bench.Must(bench.Post(s.URL+"/top4aew4fe9yeehu/job/result", "application/json", bytes.NewBuffer(resultJSON)))
	require.Equal(bench.T, http.StatusOK, resp.StatusCode)
}

func cliLogin(cli *testHTTPClient, teamID int) {
	resp := cli.Must(
		cli.PostForm(
			s.URL+"/login",
			url.Values{
				"team_id":  {fmt.Sprint(teamID)},
				"password": {fmt.Sprint("dummy-pass-%d", teamID-1000)},
			},
		),
	)
	require.Equal(cli.T, "/", resp.Request.URL.Path)
}

func TestPostJob(t *testing.T) {
	var (
		cli   = newTestClient(t)
		cli2  = newTestClient(t)
		bench = newTestClient(t)
	)

	var resp *http.Response

	// cli: ログイン
	resp = cli.Must(cli.PostForm(s.URL+"/login", url.Values{"team_id": {"1200"}, "password": {"dummy-pass-%d200"}}))
	require.Equal(t, "/", resp.Request.URL.Path)

	// bench: ジョブ取る
	resp = bench.Must(bench.Post(s.URL+"/top4aew4fe9yeehu/job/new", "", nil))
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// cli: ジョブいれる
	resp = cli.Must(cli.PostForm(s.URL+"/queue", url.Values{"ip_addr": {"127.0.0.1"}}))
	assert.Contains(t, readAll(resp.Body), `<span class="label label-default">1200*</span>`, "ジョブ入った表示")

	// cli2: ログイン
	resp = cli2.Must(cli2.PostForm(s.URL+"/login", url.Values{"team_id": {"1100"}, "password": {"dummy-pass-%d100"}}))
	require.Equal(t, "/", resp.Request.URL.Path)
	assert.Contains(t, readAll(resp.Body), `<span class="label label-default">1200</span>`, "他人のジョブ入った表示")

	// bench: ジョブ取る
	j := benchGetJob(bench)

	// cli: ジョブいれる (2) → 入らない
	resp = cli.Must(cli.PostForm(s.URL+"/queue", url.Values{"ip_addr": {"127.0.0.1"}}))
	assert.Contains(t, readAll(resp.Body), `Job already queued`)

	// cli2: ジョブ入れる → 入る
	resp = cli2.Must(cli2.PostForm(s.URL+"/queue", url.Values{"ip_addr": {"127.0.0.2"}}))
	assert.NotContains(t, readAll(resp.Body), `Job already queued`)

	// bench: ジョブ取る → 放置
	j2 := benchGetJob(bench)
	_ = j2

	// cli: トップリロード
	resp = cli.Must(cli.Get(s.URL + "/"))
	assert.Contains(t, readAll(resp.Body), `<span class="label label-success">1200*</span>`, "ジョブ実行中の表示")

	// bench: 結果入れる
	benchPostResult(bench, j, &score.Output{Pass: false, Score: 5000})

	// cli: トップリロード
	resp = cli.Must(cli.Get(s.URL + "/"))
	body := readAll(resp.Body)
	require.Contains(t, body, `<th>Status</th><td>FAIL</td>`)
	require.Contains(t, body, `<th>Score</th><td>5000</td>`)
	require.Contains(t, body, `<th>Best</th><td>-</td>`)

	// cli: ジョブいれる (3)
	resp = cli.Must(cli.PostForm(s.URL+"/queue", url.Values{"ip_addr": {"127.0.0.1"}}))
	assert.NotContains(t, readAll(resp.Body), `Job already queued`)

	// bench: ジョブ取る
	j = benchGetJob(bench)

	// bench: 結果入れる
	benchPostResult(bench, j, &score.Output{Pass: true, Score: 3000})

	// cli: トップリロード
	resp = cli.Must(cli.Get(s.URL + "/"))
	body = readAll(resp.Body)
	require.Contains(t, body, `<th>Status</th><td>PASS</td>`)
	require.Contains(t, body, `<th>Score</th><td>3000</td>`)
	require.Contains(t, body, `<th>Best</th><td>3000</td>`)
	require.Regexp(t, `<td>ダミーチーム200</td>\s*<td>3000</td>`, body)
	require.NotContains(t, body, "ダミーチーム100")

	// bench: 結果入れる
	benchPostResult(bench, j2, &score.Output{Pass: true, Score: 4500})

	resp = cli2.Must(cli2.Get(s.URL + "/"))
	body = readAll(resp.Body)
	require.Contains(t, body, `<th>Status</th><td>PASS</td>`)
	require.Contains(t, body, `<th>Score</th><td>4500</td>`)
	require.Contains(t, body, `<th>Best</th><td>4500</td>`)
	require.Regexp(t, `<td>ダミーチーム100</td>\s*<td>4500</td>(?s:.*)<td>ダミーチーム200</td>\s*<td>3000</td>`, body)
}

func TestPostJobNotWithinContestTime(t *testing.T) {
	cli := newTestClient(t)
	cliLogin(cli, 1150)

	var resp *http.Response

	*startsAtHour = 24
	resp = cli.Must(cli.PostForm(s.URL+"/queue", url.Values{"ip_addr": {"127.0.0.1"}}))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "Qualifier has not started yet\n", readAll(resp.Body))
	*startsAtHour = -1

	*endsAtHour = 0
	resp = cli.Must(cli.PostForm(s.URL+"/queue", url.Values{"ip_addr": {"127.0.0.1"}}))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "Qualifier has finished\n", readAll(resp.Body))
	*endsAtHour = -1
}

func TestUpdateTeam(t *testing.T) {
	cli := newTestClient(t)
	cliLogin(cli, 1151)

	resp := cli.Must(cli.PostForm(s.URL+"/team", url.Values{"instance_name": {"xxxxxx"}}))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, readAll(resp.Body), `value="xxxxxx"`)

	resp = cli.Must(cli.Get(s.URL + "/"))
	assert.Contains(t, readAll(resp.Body), `value="xxxxxx"`)
}
