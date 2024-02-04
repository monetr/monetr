package models

import "time"

type TellerLinkStatus uint8

//go:generate stringer -type=TellerLinkStatus -output=teller_link.strings.go
const (
	TellerLinkStatusUnknown      TellerLinkStatus = 0
	TellerLinkStatusPending      TellerLinkStatus = 1
	TellerLinkStatusSetup        TellerLinkStatus = 2
	TellerLinkStatusDisconnected TellerLinkStatus = 3
)

type TellerLink struct {
	tableName string `pg:"teller_links"`

	TellerLinkId         uint64           `json:"-" pg:"teller_link_id,notnull,pk,type:'bigserial'"`
	AccountId            uint64           `json:"-" pg:"account_id,notnull,type:'bigint',unique:per_account"`
	Account              *Account         `json:"-" pg:"rel:has-one"`
	SecretId             uint64           `json:"-" pg:"secret_id,type:'bigint'"`
	Secret               *Secret          `json:"-" pg:"rel:has-one"`
	EnrollmentId         string           `json:"-" pg:"enrollment_id,notnull,unique:per_account"`
	UserId               string           `json:"-" pg:"teller_user_id,notnull"`
	Status               TellerLinkStatus `json:"status" pg:"status,notnull,default:0"`
	ErrorCode            *string          `json:"errorCode,omitempty" pg:"error_code"`
	InstitituionName     string           `json:"institutionName" pg:"institution_name,notnull"`
	LastManualSync       *time.Time       `json:"lastManualSync" pg:"last_manual_sync"`
	LastSuccessfulUpdate *time.Time       `json:"lastSuccessfulUpdate" pg:"last_successful_update"`
	LastAttemptedUpdate  *time.Time       `json:"lastAttemptedUpdate" pg:"last_attempted_update"`
	UpdatedAt            time.Time        `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt            time.Time        `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId      uint64           `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser        *User            `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
