package models

import (
	"github.com/hawx/phemera/markdown"
	"strconv"
	"time"
)

type Entries []Entry

func (l Entries) Len() int {
	return len(l)
}

func (l Entries) Less(i, j int) bool {
	return l[i].Time < l[j].Time
}

func (l Entries) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type Entry struct {
	Time string
	Body string
}

func (e Entry) Rendered() string {
	return markdown.Render(e.Body)
}

func (e Entry) RssTime() string {
	secs, _ := strconv.ParseInt(e.Time, 10, 0)
	return time.Unix(secs, 0).Format(time.RFC822)
}
