package utils

import (
	"log"
	"time"
)

func ParseMonthYear(s string) (time.Time, error) {
	log.Println("parseMonthYear (handler): called with s=", s)
	return time.Parse("01-2006", s)
}
