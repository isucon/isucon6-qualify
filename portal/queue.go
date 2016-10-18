package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/isucon/isucon6-qualify/portal/job"
	"github.com/pkg/errors"
)

type errAlreadyQueued int

func (n errAlreadyQueued) Error() string {
	return fmt.Sprintf("job already queued (teamID=%d)", n)
}

func enqueueJob(teamID int, ipAddr string) error {
	var id int
	err := db.QueryRow(`
      SELECT id FROM queues
      WHERE team_id = ? AND status IN ('waiting', 'running')`, teamID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		// 行がない場合はINSERTする
	case err != nil:
		return errors.Wrap(err, "failed to enqueue job when selecting table")
	default:
		return errAlreadyQueued(teamID)
	}
	// XXX: worker nodeが死んだ時のために古くて実行中のジョブがある場合をケアした方が良いかも

	// XXX: ここですり抜けて二重で入る可能性がある
	_, err = db.Exec(`
      INSERT INTO queues (team_id, ip_address) VALUES (?, ?)`, teamID, ipAddr)
	if err != nil {
		return errors.Wrap(err, "enqueue job failed")
	}
	return nil
}

func dequeueJob(benchNode string) (*job.Job, error) {
	var j job.Job
	err := db.QueryRow(`
    SELECT id, team_id, ip_address FROM queues
      WHERE status = 'waiting' ORDER BY id LIMIT 1`).Scan(&j.ID, &j.TeamID, &j.IPAddr)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "dequeue job failed when scanning job")
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to dequeue job when beginning tx")
	}
	ret, err := tx.Exec(`
    UPDATE queues SET status = 'running', bench_node = ?
      WHERE id = ? AND status = 'waiting'`, benchNode, j.ID)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "failed to dequeue job when locking")
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "failed to dequeue job when checking affected rows")
	}
	if affected > 1 {
		tx.Rollback()
		return nil, fmt.Errorf("failed to dequeue job. invalid affected rows: %d", affected)
	}
	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to dequeue job when commiting tx")
	}
	// タッチの差で別のワーカーにジョブを取られたとか
	if affected < 1 {
		return nil, nil
	}
	return &j, nil
}

func doneJob(res *job.Result) error {
	b, _ := json.Marshal(res.Output)
	resultJSON := string(b)

	log.Printf("doneJob: job=%#v output=%#v", res.Job, res.Output)

	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "doneJob failed when beginning tx")
	}
	ret, err := tx.Exec(`
    UPDATE queues SET status = 'done', result_json = ?
    WHERE
      id = ?         AND
      team_id = ?    AND
      ip_address = ? AND status = 'running'`,
		resultJSON,
		res.Job.ID,
		res.Job.TeamID,
		res.Job.IPAddr,
	)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "doneJob failed when locking")
	}
	affected, err := ret.RowsAffected()
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "doneJob failed when checking affected rows")
	}
	if affected != 1 {
		tx.Rollback()
		return fmt.Errorf("doneJob failed. invalid affected rows=%d", affected)
	}

	if res.Output.Pass {
		_, err = tx.Exec("INSERT INTO scores (team_id, score) VALUES (?, ?)", res.Job.TeamID, res.Output.Score)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, "INSERT INTO scores")
		}
		_, err = tx.Exec(`
			INSERT INTO team_scores (team_id, latest_score, best_score)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
			best_score = GREATEST(best_score, VALUES(best_score)),
			latest_score = VALUES(latest_score)
		`,
			res.Job.TeamID, res.Output.Score, res.Output.Score,
		)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, "INSERT INTO team_scores")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "doneJob failed when commiting tx")
	}
	return nil
}
