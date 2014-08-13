package views

import "github.com/hoisie/mustache"

var (
	Add, _ = mustache.ParseString(add)
	Feed, _ = mustache.ParseString(feed)
	Layout, _ = mustache.ParseString(layout)
	List, _ = mustache.ParseString(list)
)
