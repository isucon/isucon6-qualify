package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/isucon/isucon6-qualify/bench/score"
	"github.com/isucon/isucon6-qualify/portal/job"
)

func queueJob(teamID int) error {
	req, err := http.NewRequest("POST", "http://localhost:8000/queue", bytes.NewBufferString(url.Values{"ip_addr": {"0.0.0.0"}}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Cookie", fmt.Sprintf("debug_team=%d", teamID))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	log.Printf("queueJob teamID=%d: %s", teamID, res.Status)
	return nil
}

func dequeueJob() error {
	req, err := http.NewRequest("POST", "http://localhost:8000/top4aew4fe9yeehu/job/new", nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	log.Printf("newJob: %s", res.Status)

	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	var j job.Job
	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	result := job.Result{
		Job: &j,
		Output: &score.Output{
			Pass:     rand.Intn(10) < 7,
			Fail:     int64(math.Min(float64(rand.Intn(100)-50), 0)),
			Score:    int64(rand.Intn(100000)),
			Messages: []string{"m1", "m2"},
		},
	}
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	req, err = http.NewRequest("POST", "http://localhost:8000/top4aew4fe9yeehu/job/result", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	log.Printf("postResult %s", res.Status)

	return nil
}

func main() {
	http.DefaultClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }

	var wg sync.WaitGroup

	cmd := exec.Command("ab", "-l", "-n", "5000", "-c", "32", "-C", "debug_team=1", "http://127.0.0.1:8000/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})

	for i := 0; i < 20; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
				wg.Add(1)
				queueJob(rand.Intn(150) + 1)
				wg.Done()
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}

				wg.Add(1)
				dequeueJob()
				wg.Done()
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}()
	}

	cmd.Wait()
	close(done)
	wg.Wait()
}
