package api

import (
	"html/template"
	"time"
)

type Quote struct {
	Date    time.Time
	Quote   template.HTML
	Comment template.HTML
	IP      string
	ID      int
	Likes   int
	Votes   int
}
