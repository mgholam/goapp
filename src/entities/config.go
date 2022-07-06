package entities

type Config struct {
	Port              string
	GitClientID       string
	GitSecurity       string
	GoogleClientID    string
	GoogleSecurity    string
	BaseURL           string
	LDAPServerAddress string
	LDAPDomainName    string
	JWTSecret         string
}

func NewConfig() Config {
	return Config{
		Port:              "3000",
		GitClientID:       "",
		GitSecurity:       "",
		GoogleClientID:    "",
		GoogleSecurity:    "",
		LDAPServerAddress: "",
		LDAPDomainName:    "",
		BaseURL:           "http://localhost",
		JWTSecret:         "",
	}
}
