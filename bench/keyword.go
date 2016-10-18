package main

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
)

type keyword struct {
	Keyword     string   `json:"k"`
	Description string   `json:"v"`
	Links       []string `json:"links"`
	Backlinks   []string `json:"back_links"`
	Length      int64    `json:"length"`
}

func prepareKeywords(datadir string) (keywords []*keyword, err error) {
	files := []string{"init.json", "year.json"}
	for _, f := range files {
		ks, err := loadKeywordFromFile(filepath.Join(datadir, f))
		if err != nil {
			return nil, err
		}
		keywords = append(keywords, ks...)
	}
	return
}

func prepareNewKeywords(datadir string) ([]*keyword, error) {
	return loadKeywordFromFile(filepath.Join(datadir, "ok.json"))
}

func prepareNGWords(datadir string) ([]*keyword, error) {
	return loadKeywordFromFile(filepath.Join(datadir, "ng.json"))
}

func loadKeywordFromFile(f string) (keywords []*keyword, err error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scr := bufio.NewScanner(file)
	buf := make([]byte, 0, 65536)
	scr.Buffer(buf, 16777215)
	for scr.Scan() {
		b := scr.Bytes()
		var k keyword
		err = json.Unmarshal(b, &k)
		if err != nil {
			return nil, err
		}
		keywords = append(keywords, &k)
	}
	if err := scr.Err(); err != nil {
		return nil, err
	}
	return
}

func kwShuffle(data []*keyword) {
	n := len(data)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		data[i], data[j] = data[j], data[i]
	}
}
