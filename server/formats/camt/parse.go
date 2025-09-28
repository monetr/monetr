package camt

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/pkg/errors"
)

type CamtDocument struct {
	Statement BankToCustomerStatementV13 `xml:"BkToCstmrStmt"`
}

func Parse(reader io.Reader) (*CamtDocument, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read camt.053 buffer")
	}

	var camt CamtDocument
	if err := xml.Unmarshal(data, &camt); err != nil {
		return nil, errors.Wrap(err, "failed to parse camt.053 file")
	}

	return &camt, nil
}

// ParseTransactionAmount takes a CAMT.053 transaction entry and extracts the
// amount field as well as the currency code for the amount (if present?) and
// parses the amount in that field. Even though CAMT.053 files deserialize as
// float64, this function will safely and properly convert that value into an
// int64 for monetr's own storage; as well as changing the sign of the number
// depending on the direction of the transaction.
func ParseTransactionAmount(transaction ReportEntry15) (int64, error) {
	amount, err := currency.ParseFloatToAmount(
		transaction.Amt.Value,
		transaction.Amt.CcyAttr,
	)
	if err != nil {
		return 0, err
	}

	switch {
	case strings.EqualFold(transaction.CdtDbtInd, "DBIT"):
		return myownsanity.Abs(amount), nil
	case strings.EqualFold(transaction.CdtDbtInd, "CRDT"):
		// Credits are represented as negative in monetr.
		return myownsanity.Abs(amount) * -1, nil
	default:
		return amount, nil
	}
}
