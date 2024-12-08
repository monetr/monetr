package config

type Support struct {
	Enabled                    bool   `yaml:"enabled"`
	Email                      string `yaml:"email"`
	ChatwootSDKURL             string `json:"chatwootSdkUrl"`
	ChatwootWebsiteToken       string `yaml:"chatwootWebsiteToken"`
	ChatwootIdentityValidation string `yaml:"chatwootIdentityValidation"`
}

func (s Support) GetSupportEmail() string {
	if s.Enabled {
		return s.Email
	}

	return ""
}

func (s Support) GetChatwootEnabled() bool {
	return s.Enabled &&
		s.ChatwootSDKURL != "" &&
		s.ChatwootIdentityValidation != "" &&
		s.ChatwootWebsiteToken != ""
}
