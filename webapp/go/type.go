package main

import (
	"context"
	"time"
)

type Entry struct {
	ID            int
	AuthorID      int
	Keyword       string
	Description   string
	UpdatedAt     time.Time
	CreatedAt     time.Time
	KeywordLength int

	Html  string
	Stars []Star
}

type User struct {
	ID        int
	Name      string
	Salt      string
	Password  string
	CreatedAt time.Time
}

type Star struct {
	ID        int       `json:"id"`
	Keyword   string    `json:"keyword"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
}

type EntryWithCtx struct {
	Context context.Context
	Entry   Entry
}
