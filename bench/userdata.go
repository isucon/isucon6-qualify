package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func prepareUserdata(datadir string) (users []*user, err error) {
	info, err := os.Stat(datadir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("datadirがディレクトリではありません")
	}

	file, err := os.Open(filepath.Join(datadir, "users.txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scr := bufio.NewScanner(file)
	for scr.Scan() {
		name := scr.Text()
		users = append(users, &user{name, name})
	}
	if err := scr.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
