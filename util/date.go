package util

import "time"

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func (d *Date) NewerThan(o *Date) bool {
	return d.Year > o.Year || int(d.Month) > int(o.Month) || d.Day > o.Day
}

func NewDate() *Date {
	year, month, day := time.Now().Date()
	return &Date{year, month, day}
}
