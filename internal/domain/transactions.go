package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// TransactionID is the public transaction identifier split into typed parts.
type TransactionID struct {
	Year  int
	Month int
	Num   int
}

// String returns the compact string form of the transaction ID.
func (t TransactionID) String() string {
	return fmt.Sprintf("%04d.%02d.%d", t.Year, t.Month, t.Num)
}

func (t TransactionID) Equal(other TransactionID) bool {
	return t.Year == other.Year && t.Month == other.Month && t.Num == other.Num
}

func CurrentTransactionID(num int) TransactionID {
	return TransactionID{
		Year:  time.Now().Year(),
		Month: int(time.Now().Month()),
		Num:   num,
	}
}

// ParseTransactionID parses a transaction ID in the form YYYY.MM.NN.
func ParseTransactionID(s string) (TransactionID, error) {
	var publicIDRegex = regexp.MustCompile(`^T\d{4}\.\d{2}\.\d+$`)
	if publicIDRegex.MatchString(s) {
		year, month, num := s[1:5], s[6:8], s[9:]

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
