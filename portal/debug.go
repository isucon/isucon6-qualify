package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func expvarHandler(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")

	fmt.Fprintf(w, "%q: ", "db")
	json.NewEncoder(w).Encode(db.Stats())

	fmt.Fprintf(w, ",\n%q: ", "runtime")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"NumGoroutine": runtime.NumGoroutine(),
	})

	rows, err := db.Query("SELECT status,COUNT(*) FROM queues GROUP BY status")
	if err != nil {
		return err
	}
	defer rows.Close()
	queueStats := map[string]int{}
	for rows.Next() {
		var (
			st string
			c  int
		)
		rows.Scan(&st, &c)
		queueStats[st] = c
	}

	fmt.Fprintf(w, ",\n%q: ", "queue")
	json.NewEncoder(w).Encode(queueStats)

	fmt.Fprintf(w, ",\n%q: ", "app")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":   appVersion,
		"startedAt": appStartedAt,
	})

	expvar.Do(func(kv expvar.KeyValue) {
		fmt.Fprintf(w, ",\n")
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")

	return nil
}

func serveDebugQueue(w http.ResponseWriter, req *http.Request) error {
	rows, err := db.Query(`
      SELECT
        queues.id,team_id,name,status,queues.ip_address,IFNULL(bench_node, ''),IFNULL(result_json, ''),created_at
      FROM queues
        LEFT JOIN teams ON queues.team_id = teams.id
      ORDER BY queues.created_at DESC
      LIMIT 50
	`)
	if err != nil {
		return err
	}

	type queueItem struct {
		ID        int
		TeamID    int
		TeamName  string
		Status    string
		IPAddr    string
		BenchNode string
		Result    string
		Time      time.Time
	}

	type viewParamsDebugQueue struct {
		viewParamsLayout
		Items []*queueItem
	}

	items := []*queueItem{}

	defer rows.Close()
	for rows.Next() {
		var item queueItem
		err := rows.Scan(&item.ID, &item.TeamID, &item.TeamName, &item.Status, &item.IPAddr, &item.BenchNode, &item.Result, &item.Time)
		if err != nil {
			return err
		}

		items = append(items, &item)
	}

	return templates["debug-queue.tmpl"].Execute(w, viewParamsDebugQueue{viewParamsLayout{nil, day}, items})
}
