package models

import "time"

// Registration represents a form of verification for a user's sign up. When a
// user registers with the application if we wish to verify that user's email
// and/or phone number we will create a registration record. This record is
// tied to the login record we created (which will not be enabled) and we use
// this registration record to keep track of that user verifying their account.
// The registration record has a uuid for its primary key, upon sign up we will
// send the user a verification email with a link back to this site with that
// key. (Note: This should be on the frontend). Upon the frontend opening that
// link it will send a request to the API with the encoded key. The encoded key
// is decoded and looked up in the registration table. If a record is found the
// associated login is checked and enabled, a success message is returned to
// the user. If the registration record is not found, or the login is already
// enabled then nothing happens. The registration record must also not be
// expired.
type Registration struct {
	tableName string `pg:"registrations"`

	RegistrationId string    `json:"-" pg:"registration_id,notnull,pk,type:'uuid',default:uuid_generate_v4()"`
	LoginId        uint64    `json:"loginId" pg:"login_id,notnull,on_delete:CASCADE"`
	Login          *Login    `json:"login,omitempty" pg:"rel:has-one"`
	IsComplete     bool      `json:"isComplete" pg:"is_complete,notnull,use_zero"`
	DateCreated    time.Time `json:"-" pg:"date_created,notnull"`
	DateExpires    time.Time `json:"-" pg:"date_expires,notnull"`
}
