package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/isucon/isucon6-qualify/bench/checker"
	"github.com/isucon/isucon6-qualify/bench/score"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota

	InitializeTimeout = 5 * time.Second
	BenchmarkTimeout  = 60 * time.Second
	WaitAfterTimeout  = 15 * time.Second

	PostsPerPage = 10
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

type user struct {
	Name     string
	Password string
}

type scenarioFn func(*checker.Session, *benchdata)

type scenarioRunner struct {
	scenarios []scenarioFn
	para      int
	bdata     *benchdata
	done      chan struct{}
}

func newScenarioRunner(bd *benchdata, para int, funcs []scenarioFn) *scenarioRunner {
	return &scenarioRunner{
		bdata:     bd,
		scenarios: funcs,
		done:      make(chan struct{}, para),
	}
}

func (sr *scenarioRunner) start(ch chan *scenarioRunner) {
	para := cap(sr.done)
	for i := 0; i < para; i++ {
		ch <- sr
	}
	for range sr.done {
		ch <- sr
	}
}
func (sr *scenarioRunner) run() {
	for _, fn := range sr.scenarios {
		fn(checker.NewSession(), sr.bdata)
	}
	sr.done <- struct{}{}
}

type scenarioManager struct {
	runners []*scenarioRunner
	bdata   *benchdata
	next    chan *scenarioRunner
}

func newScenarioManager(bdata *benchdata) *scenarioManager {
	return &scenarioManager{
		bdata: bdata,
		next:  make(chan *scenarioRunner),
	}
}

func (sm *scenarioManager) register(para int, funcs []scenarioFn) {
	sr := newScenarioRunner(sm.bdata, para, funcs)
	sm.runners = append(sm.runners, sr)
}
func (sm *scenarioManager) start() {
	for _, sr := range sm.runners {
		go sr.start(sm.next)
	}
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		target  string
		datadir string

		version bool
		debug   bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&target, "target", "http://localhost", "")
	flags.StringVar(&datadir, "datadir", "data", "userdata directory")
	flags.BoolVar(&debug, "debug", false, "Debug mode")
	flags.BoolVar(&version, "version", false, "Print version information and quit.")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	// Show version
	if version {
		fmt.Fprintf(cli.errStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	targetHost, terr := checker.SetTargetHost(target)
	if terr != nil {
		outputNeedToContactUs(terr.Error())
		return ExitCodeError
	}

	initialize := make(chan bool)

	setupInitialize(targetHost, initialize)

	bdata, err := initializeBenchdata(datadir)
	if err != nil {
		outputNeedToContactUs(err.Error())
		return ExitCodeError
	}

	initReq := <-initialize
	if !initReq {
		fmt.Println(outputResultJson(false, []string{"初期化リクエストに失敗しました"}))
		return ExitCodeError
	}

	ok := start(bdata)
	msgs := []string{}
	if !debug {
		msgs = score.GetFailErrorsStringSlice()
	} else {
		msgs = score.GetFailRawErrorsStringSlice()
	}
	fmt.Println(outputResultJson(ok, msgs))

	if !ok {
		return ExitCodeError
	}
	return ExitCodeOK
}

func start(bdata *benchdata) bool {
	log.Println("start pre-checking")

	startAt := time.Now()

	// 最初にDOMチェックなどをやってしまい、通らなければさっさと失敗させる
	initialScenarios := []scenarioFn{
		cannotLoginNonexistentUserScenario,
		cannotLoginWrongPasswordScenario,
		loginScenario,
		ngwordScenario,
		loadIndexWithStarScenario,
		keywordScenario,
		postKeywordScenario,
		updateKeywordScenario,
	}
	var wg sync.WaitGroup
	wg.Add(len(initialScenarios))
	for _, fn := range initialScenarios {
		go func(fn scenarioFn) {
			defer wg.Done()
			fn(checker.NewSession(), bdata)
		}(fn)
	}
	wg.Wait()

	restDur := startAt.Add(BenchmarkTimeout).Sub(time.Now())
	if restDur <= 0 {
		score.GetInstance().SetFails(0)
		score.GetFailErrorsInstance().Append(fmt.Errorf("初期チェックがタイムアウトしました"))
	}
	if (score.GetInstance().GetFails() - score.GetInstance().GetTimeouts()) > 0  {
		return false
	}
	log.Println("pre-check finished and start main benchmarking")

	// シナリオ登録
	sm := newScenarioManager(bdata)
	sm.register(1, []scenarioFn{indexMoreAndMoreScenario})
	sm.register(2, []scenarioFn{loadIndexScenario})
	sm.register(1, []scenarioFn{
		loginScenario,
		cannotLoginNonexistentUserScenario,
		cannotLoginWrongPasswordScenario,
	})
	sm.register(2, []scenarioFn{loadIndexWithStarScenario})
	sm.register(1, []scenarioFn{ngwordScenario})
	sm.register(2, []scenarioFn{keywordScenario})
	sm.register(3, []scenarioFn{postKeywordScenario})
	sm.register(1, []scenarioFn{updateKeywordScenario})

	// ベンチ開始
	sm.start()
	timeoutCh := time.After(restDur)
L:
	for {
		select {
		case <-timeoutCh:
			break L
		case sr := <-sm.next:
			go sr.run()
		}
	}
	log.Println("benchmarking finished")
	time.Sleep(WaitAfterTimeout)
	return true
}

func outputResultJson(pass bool, messages []string) string {
	output := score.Output{
		Pass:     pass,
		Score:    score.GetInstance().GetScore(),
		Suceess:  score.GetInstance().GetSucesses(),
		Fail:     score.GetInstance().GetFails(),
		Messages: messages,
	}

	b, _ := json.Marshal(output)

	return string(b)
}

// 主催者に連絡して欲しいエラー
func outputNeedToContactUs(message string) {
	fmt.Println(outputResultJson(false, []string{"！！！主催者に連絡してください！！！", message}))
}

func setupInitialize(targetHost string, initialize chan bool) {
	go func(targetHost string) {
		client := &http.Client{
			Timeout: InitializeTimeout,
		}

		parsedURL, _ := url.Parse("/initialize")
		parsedURL.Scheme = "http"
		parsedURL.Host = targetHost

		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			return
		}

		req.Header.Set("User-Agent", checker.UserAgent)

		res, err := client.Do(req)

		if err != nil {
			initialize <- false
			return
		}
		defer res.Body.Close()
		initialize <- true
	}(targetHost)
}
