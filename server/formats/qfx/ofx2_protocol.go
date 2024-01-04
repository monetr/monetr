package qfx

type OFX struct {
	SignonResponse SignonResponseMessageSetV1 `qfx:"SIGNONMSGSRSV1"`
	BankResponse   BankResponseMessageSetV1   `qfx:"BANKMSGSRSV1"`
}

func (q *OFX) Parse(token Token) error {
	_, err := mustArray(q, token)
	if err != nil {
		return err
	}

	return nil
}

type SignonResponseMessageSetV1 struct {
	SignonResponse SignonResponse `qfx:"SONRS"`
}

func (q *SignonResponseMessageSetV1) Parse(token Token) error {
	_, err := mustArray(q, token)
	if err != nil {
		return err
	}

	return nil

}

type BankResponseMessageSetV1 struct {
	StatementTransactionResponse StatementTransactionResponse `qfx:"STMTTRNRS"`
}

type StatementTransactionResponse struct {
	TRNUID string `qfx:"TRNUID"`
	Status Status `qfx:"STATUS"`
}
