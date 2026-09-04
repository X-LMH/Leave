package service

import "time"

type ResRecord struct {
	ID    int       `json:"id"`
	Name  string    `json:"name"`
	Nigao time.Time `json:"nigao"`
}
