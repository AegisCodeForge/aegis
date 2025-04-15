package gitlib

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"time"
)

type AuthorTime struct {
	AuthorName string
	AuthorEmail string
	Time time.Time
}

func parseTimezoneOffset(s string) (int, error) {
	if s == "Z" { return 0, nil }
	if len(s) != 5 { return 0, errors.New("Invalid timezone offset string") }

	hour := (int(s[1]) - int('0') * 10) + (int(s[2]) - int('0'))
	minute := (int(s[3]) - int('0') * 10) + (int(s[4]) - int('0'))
	total := (hour * 60 + minute) * 60
	if s[0] == '-' { total = -total }
	return total, nil
}

func parseAuthorTime(s string) AuthorTime {
	res := AuthorTime {
		AuthorName: "",
		AuthorEmail: "",
		Time: time.Unix(0, 0),
	}
	re, err := regexp.Compile("([^<>]+)\\s*<([^>]+)>\\s*([^\\s]+)\\s*([^\\s]+)")
	if err != nil { log.Fatal(err) }
	matchres := re.FindSubmatch([]byte(s))
	if len(matchres) <= 0 {
		log.Fatalf("Cannot parse author-time: %s\n", s)
	}
	res.AuthorName = string(matchres[1])
	res.AuthorEmail = string(matchres[2])
	timeStampString := string(matchres[3])
	timezoneOffsetString := string(matchres[4])
	timeStampInt, err := strconv.ParseInt(timeStampString, 10, 64)
	if err != nil { timeStampInt = 0 }
	timePiece := time.Unix(timeStampInt, 0).UTC()
	timezoneOffsetInt, err := parseTimezoneOffset(timezoneOffsetString)
	if err != nil { log.Fatal(err) }
	timezone := time.FixedZone("UTC" + timezoneOffsetString, timezoneOffsetInt)
	res.Time = timePiece.In(timezone)
	return res
}

