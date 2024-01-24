package teller

type Client interface {
	GetInstitutions() ([]Institution, error)
}
