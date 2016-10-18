package main

import (
	"sync"

	"github.com/isucon/isucon6-qualify/bench/util"
)

type benchdata struct {
	users    []*user
	words    []*keyword
	ngWords  []*keyword
	allWords map[string]*keyword

	mu       sync.Mutex
	newWords []*keyword
}

func (bd *benchdata) randomUser() *user {
	return bd.users[util.RandomNumber(len(bd.users))]
}

func (bd *benchdata) randomWord() *keyword {
	return randomKw(bd.words)
}

func (bd *benchdata) randomNGWord() *keyword {
	return randomKw(bd.ngWords)
}

func (bd *benchdata) newWord() *keyword {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	var kw *keyword
	kw, bd.newWords = bd.newWords[0], bd.newWords[1:]
	return kw
}

func randomKw(kws []*keyword) *keyword {
	return kws[util.RandomNumber(len(kws))]
}

func initializeBenchdata(datadir string) (bdata *benchdata, err error) {
	bdata = &benchdata{}
	bdata.users, err = prepareUserdata(datadir)
	if err != nil {
		return
	}
	bdata.words, err = prepareKeywords(datadir)
	if err != nil {
		return
	}
	bdata.newWords, err = prepareNewKeywords(datadir)
	if err != nil {
		return
	}
	kwShuffle(bdata.newWords)

	allKw := make(map[string]*keyword, len(bdata.words)+len(bdata.newWords))
	for _, k := range bdata.words {
		allKw[k.Keyword] = k
	}
	for _, k := range bdata.newWords {
		allKw[k.Keyword] = k
	}
	bdata.allWords = allKw

	bdata.ngWords, err = prepareNGWords(datadir)
	if err != nil {
		return
	}
	return
}
