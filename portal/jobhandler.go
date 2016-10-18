package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/isucon/isucon6-qualify/portal/job"
)

func contestFinishTime() time.Time {
	if *endsAtHour < 0 {
		return time.Date(2038, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	// その日の18時がコンテスト終了日時
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, *endsAtHour, 0, 0, 0, locJST)
}

// serveQueueJob は参加者がベンチマーカのジョブをキューに挿入するエンドポイント。
func serveQueueJob(w http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodPost {
		return errHTTP(http.StatusMethodNotAllowed)
	}

	// 18時になったらコンテスト終了なのでジョブを挿入させない
	switch getContestStatus() {
	case contestStatusNotStarted:
		return errHTTPMessage{http.StatusForbidden, "Qualifier has not started yet"}
	case contestStatusEnded:
		return errHTTPMessage{http.StatusForbidden, "Qualifier has finished"}
	}

	team, err := loadTeamFromSession(req)
	if err != nil {
		return err
	}
	if team == nil {
		return errHTTP(http.StatusForbidden)
	}

	ipAddr := req.FormValue("ip_addr")
	if ipAddr == "" {
		return errHTTP(http.StatusBadRequest)
	}

	err = updateTeamIPAddr(team, ipAddr)
	if err != nil {
		return errors.Wrapf(err, "updateTeamIPAddr(team=%#v, ipAddr=%#v)", team, ipAddr)
	}

	err = enqueueJob(team.ID, ipAddr)
	if err != nil {
		if _, ok := err.(errAlreadyQueued); ok {
			// ユーザに教えてあげる
			return serveIndexWithMessage(w, req, "Job already queued")
		}

		return err
	}

	// TODO(motemen): flash
	http.Redirect(w, req, "/", http.StatusFound)

	return nil
}

func updateTeamIPAddr(team *Team, ipAddr string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec("UPDATE teams SET ip_address = ? WHERE id = ?", ipAddr, team.ID)
	if err == nil {
		var nRows int64
		nRows, err = res.RowsAffected()
		if err == nil && nRows > 1 {
			err = fmt.Errorf("RowsAffected was %#v (> 1)", nRows)
		}
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// 新しいジョブを取り出す。ジョブが無い場合は 204 を返す
// クライアントは定期的(3秒おきくらい)にリクエストしてジョブを確認する
func serveNewJob(w http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodPost {
		return errHTTP(http.StatusMethodNotAllowed)
	}
	benchNode := req.FormValue("bench_node")
	j, err := dequeueJob(benchNode)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil
	}
	if j == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(j)
	w.Write(b)
	return nil
}

func servePostResult(w http.ResponseWriter, req *http.Request) error {
	if req.Method != http.MethodPost {
		http.Error(w, "Method Not Allowd", http.StatusMethodNotAllowed)
		return nil
	}
	var res job.Result
	if req.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return nil
	}
	err := json.NewDecoder(req.Body).Decode(&res)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return nil
	}
	err = doneJob(&res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success":true}`)
	return nil
}
