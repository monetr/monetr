package qfx

type SignonResponse struct {
	Status        Status `qfx:"STATUS"`
	Language      string `qfx:"LANGUAGE"`
	DateServer    string `qfx:"DTSERVER"`
	DateProfUp    string `qfx:"DTPROFUP"`
	DateAccountUp string `qfx:"DTACCTUP"`
}

type Status struct {
	Code     string   `qfx:"CODE"`
	Severity Severity `qfx:"SEVERITY"`
	Message  string   `qfx:"MESSAGE"`
}
