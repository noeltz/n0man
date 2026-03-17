package security

import ()

func RedactValue(v string) string {
	if len(v) <= 8 {
		return "********"
	}
	visible := 4
	if len(v) > 20 {
		visible = 6
	}
	return v[:visible] + "****" + v[len(v)-visible:]
}

func getRiskLevelString(r RiskLevel) string {
	switch r {
	case RiskLevelCritical:
		return "CRITICAL"
	case RiskLevelHigh:
		return "HIGH"
	case RiskLevelMedium:
		return "MEDIUM"
	case RiskLevelLow:
		return "LOW"
	default:
		return "NONE"
	}
}

type SecretType string

const (
	SecretTypeAPIKey        SecretType = "api_key"
	SecretTypeAnthropicKey  SecretType = "anthropic_key"
	SecretTypeGenericAPIKey SecretType = "generic_api_key"
	SecretTypePassword      SecretType = "password"
	SecretTypePrivateKey    SecretType = "private_key"
	SecretTypeToken         SecretType = "token"
	SecretTypeAWSKey        SecretType = "aws_key"
	SecretTypeGitHubToken   SecretType = "github_token"
	SecretTypeJWT           SecretType = "jwt"
	SecretTypeDatabaseURL   SecretType = "database_url"
	SecretTypeGeneric       SecretType = "generic_secret"
	SecretTypePII           SecretType = "pii"
	SecretTypeCreditCard    SecretType = "credit_card"
	SecretTypeSSN           SecretType = "ssn"
	SecretTypeEmail         SecretType = "email"
	SecretTypeIPAddress     SecretType = "ip_address"
)

type Location struct {
	FilePath   string
	LineNumber int
	LineText   string
}

type Finding struct {
	Type       SecretType
	Value      string
	RawValue   string
	Location   Location
	Confidence float64
	Reasons    []string
	RiskLevel  RiskLevel
	Context    string
}

type RiskLevel int

const (
	RiskLevelNone RiskLevel = iota
	RiskLevelLow
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

func (r RiskLevel) String() string {
	return getRiskLevelString(r)
}
