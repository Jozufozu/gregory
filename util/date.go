package util

import (
	"math"
	"time"
)

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func (d *Date) NewerThan(o *Date) bool {
	return d.Year > o.Year || int(d.Month) > int(o.Month) || d.Day > o.Day
}

func (d *Date) DaysSince(o *Date) int {
	date := time.Date(o.Year, o.Month, o.Day, 0, 0, 0, 0, time.Local)
	since := date.Sub(time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.Local))
	return int(math.Floor(since.Hours() / 24.0))
}

func NewDate() *Date {
	year, month, day := time.Now().Date()
	return &Date{year, month, day}
}
