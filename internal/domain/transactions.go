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

func CheckTransactionIDRange(id1 TransactionID, id2 TransactionID) error {
	// id1 is before id2
	if id1.Year < id2.Year {
		return nil
	} else if id2.Year < id1.Year {
		return fmt.Errorf("transaction id %s is before %s", id1, id2)
	}
	// year is the same

	if id1.Month < id2.Month {
		return nil
	} else if id2.Month < id1.Month {
		return fmt.Errorf("transaction id %s is before %s", id1, id2)
	}
	// year and month are the same

	if id1.Num < id2.Num {
		return nil
	} else if id2.Num < id1.Num {
		return fmt.Errorf("transaction id %s is before %s", id1, id2)
	}
	// id1 and id2 are the same

	return fmt.Errorf("transaction id1 must be before id2 (got %s and %s)", id1, id2)
}

// ParseTransactionID parses a transaction ID in the form YYYY.MM.NN.
func ParseTransactionID(s string) (TransactionID, error) {
	var publicIDRegex = regexp.MustCompile(`^\d{4}\.\d{2}\.\d+$`)
	if publicIDRegex.MatchString(s) {
		year, month, num := s[0:4], s[5:7], s[8:]

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
