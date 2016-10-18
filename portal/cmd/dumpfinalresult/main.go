package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/isucon/isucon6-qualify/bench/score"
)

const (
	// 予選各日の終了時スコアにおける上位3チーム
	generalSeatsForOneDay = 3
	// 以上を除いて、予選両日を通し、1の該当チームを除いた中での上位9チーム
	generalSeatsForWhole = 9
	// 以上を除いて、予選両日を通し学生枠参加チーム内における上位10チーム
	studentsSeatsForWhole = 10
)

var dsn = flag.String("base-dsn", "root:root@tcp(127.0.0.1:3306)", "database `dsn`")
var day = flag.Int("day", 0, "contest `day`")

type teamScore struct {
	ID           int
	Name         string
	InstanceName string
	Category     string
	FinalScore   int64
	Passed       bool
}

func (ts teamScore) PassedString() string {
	if ts.Passed {
		return "PASS"
	} else {
		return "FAIL"
	}
}

func (ts teamScore) CategoryString() string {
	if ts.Category == "general" {
		return "一般"
	} else if ts.Category == "students" {
		return "学生"
	} else {
		return ts.Category
	}
}

func (ts teamScore) EffectiveScore() int64 {
	if ts.Passed {
		return ts.FinalScore
	} else {
		return 0
	}
}

type byScore []*teamScore

func (ss byScore) Len() int { return len(ss) }
func (ss byScore) Less(i, j int) bool {
	if ss[i].EffectiveScore() == ss[j].EffectiveScore() {
		return ss[i].FinalScore > ss[j].FinalScore
	}
	return ss[i].EffectiveScore() > ss[j].EffectiveScore()
}
func (ss byScore) Swap(i, j int) { ss[i], ss[j] = ss[j], ss[i] }

func main() {
	flag.Parse()
	db, err := sql.Open("mysql", *dsn+fmt.Sprintf("/isu6qportal_day%d", *day)+"?parseTime=true&loc=Asia%2FTokyo&time_zone='Asia%2FTokyo'")
	if err != nil {
		log.Fatal(err)
	}

	scores := []*teamScore{}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}

	var deadline time.Time
	if *day == 1 {
		deadline = time.Date(2016, time.September, 17, 18, 0, 0, 0, jst)
	} else {
		deadline = time.Date(2016, time.September, 18, 18, 0, 0, 0, jst)
	}

	log.Printf("deadline: %s", deadline)

	rows, err := db.Query("SELECT id,name,category,IFNULL(instance_name,'') FROM teams")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var s teamScore
		err := rows.Scan(&s.ID, &s.Name, &s.Category, &s.InstanceName)
		if err != nil {
			log.Fatal(err)
		}

		var (
			output     score.Output
			resultJSON []byte
		)
		err = db.QueryRow(`
			SELECT result_json FROM queues
			WHERE team_id = ?
			  AND status = 'done'
			  AND created_at <= ?
			ORDER BY updated_at DESC
			LIMIT 1
		`, s.ID, deadline).Scan(&resultJSON)
		if err == sql.ErrNoRows {
			log.Printf("team %v (%d) has not sent any job", s.Name, s.ID)
			continue
		} else if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(resultJSON, &output)
		if err != nil {
			log.Fatal(err)
		}

		s.FinalScore = output.Score
		s.Passed = output.Pass

		scores = append(scores, &s)
	}

	var (
		generalSeatsForOneDayLeft = generalSeatsForOneDay
		generalSeatsForWholeLeft  = generalSeatsForWhole
		studentsSeatsForWholeLeft = studentsSeatsForWhole
	)
	_ = generalSeatsForWholeLeft
	_ = studentsSeatsForWholeLeft
	sort.Sort(byScore(scores))
	for _, s := range scores {
		var note string
		if generalSeatsForOneDayLeft > 0 {
			generalSeatsForOneDayLeft--
			note = fmt.Sprintf("一般/当日枠/確定(%d)", generalSeatsForOneDay-generalSeatsForOneDayLeft)
		} else if generalSeatsForWholeLeft > 0 {
			generalSeatsForWholeLeft--
			note = fmt.Sprintf("一般/通算枠/候補(%d)", generalSeatsForWhole-generalSeatsForWholeLeft)
		} else if s.Category == "students" && studentsSeatsForWholeLeft > 0 {
			studentsSeatsForWholeLeft--
			note = fmt.Sprintf("学生/通算枠/候補(%d)", studentsSeatsForWhole-studentsSeatsForWholeLeft)
		}

		fmt.Printf("%d\t%s\t%s\t%s\t%d\t%s\t%s\n", s.ID, s.Name, s.CategoryString(), s.PassedString(), s.FinalScore, s.InstanceName, note)
	}
}
