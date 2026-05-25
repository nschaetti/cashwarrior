package domain

import (
	"fmt"
	"regexp"
	"strconv"
)

// TransactionID is the public transaction identifier split into typed parts.
type TransactionID struct {
	Year  int
	Month int
	Num   int
}

// String returns the compact string form of the transaction ID.
func (t TransactionID) String() string {
	return fmt.Sprintf("%d.%d.%d", t.Year, t.Month, t.Num)
}

// ParseTransactionID parses a transaction ID in the form YYYY.MM.NN.
func ParseTransactionID(s string) (TransactionID, error) {
	var publicIDRegex = regexp.MustCompile(`^\d{4}\.\d{2}\.\d$`)
	if publicIDRegex.MatchString(s) {
		year, month, num := s[:4], s[5:7], s[8:]

		// Parse year.
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			return TransactionID{}, err
		}

		// Parse month.
		monthInt, err := strconv.Atoi(month)
		if err != nil {
			return TransactionID{}, err
		}

		// Parse sequence number.
		numInt, err := strconv.Atoi(num)
		if err != nil {
			return TransactionID{}, err
		}

		return TransactionID{
			Year:  yearInt,
			Month: monthInt,
			Num:   numInt,
		}, nil
	}
	return TransactionID{}, fmt.Errorf("invalid transaction id: %s", s)
}
