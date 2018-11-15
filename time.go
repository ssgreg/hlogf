package main

import (
	"strconv"
	"time"
)

func parseTimeInt64(v int64) time.Time {
	switch {
	case v > 1e18:
	case v > 1e15:
		v *= 1e3
	case v > 1e12:
		v *= 1e6
	default:
		return time.Unix(v, 0)
	}

	return time.Unix(v/1e9, v%1e9)
}

func timestampToTime(ts []byte) (time.Time, bool) {
	tsI, err := strconv.Atoi(string(ts))
	if err != nil {
		return time.Time{}, false
	}

	return parseTimeInt64(int64(tsI)), true
}

func encodeTime(ts []byte) (time.Time, bool) {
	if len(ts) < 3 {
		return time.Time{}, false
	}
	ts = ts[1 : len(ts)-1]

	t, ok := timestampToTime(ts)
	if ok {
		return t, true
	}

	t, err := time.Parse(time.RFC3339Nano, string(ts))
	if err == nil {
		return t, true
	}

	t, err = time.Parse(time.RFC3339, string(ts))
	if err == nil {
		return t, true
	}

	return time.Time{}, false
}
