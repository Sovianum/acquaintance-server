package model

import (
	"fmt"
	"time"
)

const (
	layout = "2006-01-02T15:04:05"
)

type QuotedTime time.Time

func (t QuotedTime) String() string {
	return time.Time(t).Format(layout)
}

func (t QuotedTime) MarshalJSON() ([]byte, error) {
	ts := time.Time(t).Format(layout)
	stamp := fmt.Sprintf("\"%vZ\"", ts)

	return []byte(stamp), nil
}

func (t *QuotedTime) UnmarshalJSON(b []byte) error {
	var inputS = string(b)
	var ts, err = time.Parse(layout, inputS[1:len(inputS)-2]) // slicing removes quotes and timezone symbol

	if err != nil {
		return err
	}

	*t = QuotedTime(ts)
	return nil
}
