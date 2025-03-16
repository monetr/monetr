package camt

import (
	"encoding/xml"
	"io"

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
