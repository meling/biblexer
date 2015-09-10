package biblexer

import (
	"net/url"
	"time"
)

// type bibtype string

//go:generate stringer -type=bibtype
// const (
// 	article bibtype = iota
// 	book
// 	misc
// 	proceedings
// 	inproceedings
// )

// var btypes = [...]string{"article", "book", "misc", "proceedings", "inproceedings"}

var btypes = map[string]bool{
	"article":       true,
	"book":          true,
	"misc":          true,
	"proceedings":   true,
	"inproceedings": true,
}

type bibentry struct {
	bibtype string
	citekey string
	author  string
	title   string
	date    time.Time
	pdf     url.URL
}
