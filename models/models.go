package models

import (
	"strconv"
	"time"

	"hawx.me/code/phemera/markdown"
)

type Entries []Entry

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
