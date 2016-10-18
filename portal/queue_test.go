package main

import (
	"reflect"
	"testing"

	"github.com/isucon/isucon6-qualify/bench/score"
	"github.com/isucon/isucon6-qualify/portal/job"
)

func TestEnqueueJob(t *testing.T) {
	// 事前に `TRUNCATE queues` しないと動きません…(自動でやるの怖いので)
	// go-test-mysqld使おうとしたけどなんか上手く動かなかった…
	var err error
	initWeb()

	// ジョブを積む
	err = enqueueJob(11, "192.168.0.1")
	if err != nil {
		t.Errorf("failed to enqueue job: %s", err)
	}

	// 同じteamIDはもう積めない
	err = enqueueJob(11, "192.168.0.2")
	if err == nil || err.Error() != "job already queued (teamID=11)" {
		t.Errorf("something went wrong: %v", err)
	}

	// 別のteamIDは積める
	err = enqueueJob(12, "192.168.1.1")
	if err != nil {
		t.Errorf("failed to enqueue job: %s", err)
	}

	// 最初のジョブが返って来る
	j, err := dequeueJob("host1")
	if err != nil {
		t.Errorf("something went wrong: %s", err)
	}
	expect := &job.Job{
		ID:     1,
		TeamID: 11,
		IPAddr: "192.168.0.1",
	}
	if !reflect.DeepEqual(j, expect) {
		t.Errorf("something went wrong: %#v", j)
	}

	// 次のジョブが返って来る
	j2, err := dequeueJob("host2")
	if err != nil {
		t.Errorf("something went wrong: %s", err)
	}
	expect2 := &job.Job{
		ID:     2,
		TeamID: 12,
		IPAddr: "192.168.1.1",
	}
	if !reflect.DeepEqual(j2, expect2) {
		t.Errorf("something went wrong: %#v", j)
	}

	// もうジョブは無いはず
	j3, err := dequeueJob("host3")
	if err != nil {
		t.Errorf("something went wrong: %s", err)
	}
	if j3 != nil {
		t.Errorf("something went wrong: %#v", j3)
	}

	// ジョブ実行中もそのteamIDのジョブは積めない
	err = enqueueJob(11, "192.168.0.3")
	if err == nil || err.Error() != "job already queued (teamID=11)" {
		t.Errorf("something went wrong")
	}

	// ジョブ終了
	res := &job.Result{
		Job: &job.Job{
			ID:     1,
			TeamID: 11,
			IPAddr: "192.168.0.1",
		},
		Output: &score.Output{},
	}
	err = doneJob(res)
	if err != nil {
		t.Errorf("error of doneJob should be nil but: %s", err)
	}

	// 改めてジョブが積めるようになる
	err = enqueueJob(11, "192.168.0.1")
	if err != nil {
		t.Errorf("failed to enqueue job: %s", err)
	}

	// あとかたづけ
	res = &job.Result{
		Job: &job.Job{
			ID:     2,
			TeamID: 12,
			IPAddr: "192.168.1.1",
		},
		Output: &score.Output{},
	}
	err = doneJob(res)
	if err != nil {
		t.Error(err)
	}

	j, err = dequeueJob("host4")
	if err != nil {
		t.Error(err)
	}
	res = &job.Result{
		Job:    j,
		Output: &score.Output{},
	}
	err = doneJob(res)
	if err != nil {
		t.Error(err)
	}
}
