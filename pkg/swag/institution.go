package swag

type InstitutionStatus uint8

const (
	UnknownInstitutionStatus InstitutionStatus = 0
	HealthyInstitutionStatus InstitutionStatus = iota
	DegradedInstitutionStatus
	DownInstitutionStatus
)

type InstitutionResponse struct {
	Name         string                    `json:"name"`
	URL          string                    `json:"url,omitempty"`
	PrimaryColor *string                   `json:"primaryColor,omitempty"`
	Logo         *string                   `json:"logo,omitempty"`
	Status       InstitutionStatusResponse `json:"status"`
}

type InstitutionStatusResponse struct {
	Transactions
}
