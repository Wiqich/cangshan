package sql

import (
	gosql "database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

type NullInt64 gosql.NullInt64
type NullFloat64 gosql.NullFloat64
type NullString gosql.NullString
type NullBool gosql.NullBool

var (
	zeroTime   time.Time
	timeFormat = "2006-01-02 15:04:05.999999"
	// DefaultLocation use as default location for parse datetime value in query result
	DefaultLocation = time.Now().Location()
)

// NullTime is datetime value scanner for sql database
type NullTime struct {
	Time     time.Time
	Valid    bool
	Location *time.Location
}

// Scan raw value from sql driver to get time.Time value
func (nt *NullTime) Scan(value interface{}) (err error) {
	nt.Valid = false
	if value == nil {
		return
	}
	switch v := value.(type) {
	case time.Time:
		nt.Time, nt.Valid = v, true
		return
	case []byte:
		nt.Time, err = parseDateTime(string(v), nt.Location)
		nt.Valid = (err == nil)
		return
	case string:
		nt.Time, err = parseDateTime(v, nt.Location)
		nt.Valid = (err == nil)
		return
	}
	return fmt.Errorf("Can't convert %T to time.Time", value)
}

// Value returns raw driver's value
func (nt *NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func parseDateTime(str string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = DefaultLocation
	}
	base := "0000-00-00 00:00:00.0000000"
	switch len(str) {
	case 10, 19, 21, 22, 23, 24, 25, 26: // up to "YYYY-MM-DD HH:MM:SS.MMMMMM"
		if str == base[:len(str)] {
			return zeroTime, nil
		}
		return time.ParseInLocation(timeFormat[:len(str)], str, loc)
	}
	return zeroTime, fmt.Errorf("Invalid Time-String: %s", str)
}
