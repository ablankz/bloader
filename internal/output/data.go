package output

import (
	"fmt"
	"strconv"
)

// Close is a function that closes the output service
type Close func() error

// WriteHTTPData represents the data to be written
type WriteHTTPData struct {
	ID               string
	Success          bool
	SendDatetime     string
	ReceivedDatetime string
	Count            int
	ResponseTime     int
	StatusCode       string
	Data             any
	RawData          any
}

// ToSlice converts the WriteHTTPData to a slice
func (d WriteHTTPData) ToSlice() []string {
	return []string{
		d.ID,
		strconv.FormatBool(d.Success),
		d.SendDatetime,
		d.ReceivedDatetime,
		strconv.Itoa(d.Count),
		strconv.Itoa(d.ResponseTime),
		d.StatusCode,
		fmt.Sprintf("%v", d.Data),
		fmt.Sprintf("%v", d.RawData),
	}
}
