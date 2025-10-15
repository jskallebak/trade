package handlers

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func decimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var pgNum pgtype.Numeric
	err := pgNum.Scan(d.String())
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("error converting decimal to numeric: %v", err)
	}
	return pgNum, nil
}
