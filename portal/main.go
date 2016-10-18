package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/braintree/manners"
	"github.com/golang/gddo/httputil"
	"github.com/lestrrat/go-server-starter/listener"
)

var (
	addr         = flag.String("listen", "localhost:3333", "`address` to listen to")
	startsAtHour = flag.Int("starts-at", 10, "`hour` the content starts at (JST), no limits when negative")
	endsAtHour   = flag.Int("ends-at", 18, "`hour` the contest finishes at (JST), no limits when negative")
)

var (
	appVersion   = "undefined"
	appStartedAt = time.Now()
)

const (
	pathPrefixInternal = "top4aew4fe9yeehu/"
)

type handler func(http.ResponseWriter, *http.Request) error

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func (fn handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var rb httputil.ResponseBuffer
	rw := responseWriter{&rb, http.StatusOK}

	defer func() {
		if rv := recover(); rv != nil {
			var buf [4096]byte
			n := runtime.Stack(buf[:], true)
			log.Printf("panic: [%s %s] %+v", req.Method, req.URL.Path, rv)
			log.Printf("%s", string(buf[:n]))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		log.Printf("method:%s\tpath:%s\tstatus:%d\tremote:%s", req.Method, req.URL.RequestURI(), rw.status, req.RemoteAddr)
	}()

	if getContestStatus() == contestStatusNotStarted && !strings.HasPrefix(req.URL.Path, "/"+pathPrefixInternal) {
		http.Error(w, "Qualifier has not started yet", http.StatusForbidden)
		return
	}

	err := fn(&rw, req)
	if err == nil {
		rb.Header().Set("X-Isu6QPortal-Version", appVersion)
		rb.WriteTo(w)
	} else {
		if he, ok := err.(httpError); ok {
			rw.status = he.httpStatus()
			http.Error(w, he.Error(), he.httpStatus())
			return
		}

		log.Printf("error: [%s %s] %s", req.Method, req.URL.Path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type contestStatus int

const (
	contestStatusNotStarted contestStatus = iota
	contestStatusStarted
	contestStatusEnded
)

func getContestStatus() contestStatus {
	now := time.Now()
	y, m, d := now.Date()

	if *startsAtHour >= 0 {
		startsAt := time.Date(y, m, d, *startsAtHour, 0, 0, 0, locJST)
		if now.Before(startsAt) {
			return contestStatusNotStarted
		}
	}
	if *endsAtHour >= 0 {
		endsAt := time.Date(y, m, d, *endsAtHour, 0, 0, 0, locJST)
		if now.After(endsAt) {
			return contestStatusEnded
		}
	}

	return contestStatusStarted
}

func buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", handler(serveIndex))
	mux.Handle("/favicon.ico", http.NotFoundHandler())
	mux.Handle("/login", handler(serveLogin))
	mux.Handle("/static/", handler(serveStatic))
	mux.Handle("/queue", handler(serveQueueJob))
	mux.Handle("/team", handler(serveUpdateTeam))

	mux.Handle("/"+pathPrefixInternal+"job/new", handler(serveNewJob))
	mux.Handle("/"+pathPrefixInternal+"job/result", handler(servePostResult))
	mux.Handle("/"+pathPrefixInternal+"debug/vars", handler(expvarHandler))
	mux.Handle("/"+pathPrefixInternal+"debug/queue", handler(serveDebugQueue))
	mux.Handle("/"+pathPrefixInternal+"debug/leaderboard", handler(serveDebugLeaderboard))

	return mux
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix("[isucon6q-portal] ")

	flag.Parse()
	if *addr == "" {
		flag.Usage()
		log.Fatal("-listen required")
	}

	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGTERM)
	go func() {
		for {
			s := <-sigc
			if s == syscall.SIGTERM {
				log.Println("got SIGTERM; shutting down...")
				manners.Close()
			}
		}
	}()

	log.Print("initializing...")

	err := initWeb()
	if err != nil {
		log.Fatal(err)
	}

	mux := buildMux()

	var l net.Listener
	ll, err := listener.ListenAll()
	if err != nil {
		log.Printf("go-server-starter: %s", err)
		log.Printf("fallback to standalone; server starting at %s", *addr)

		l, err = net.Listen("tcp", *addr)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("running under server-starter; server starting at %s", ll[0].Addr())
		l = ll[0]
	}

	err = manners.Serve(l, mux)
	if err != nil {
		log.Fatal(err)
	}
}
