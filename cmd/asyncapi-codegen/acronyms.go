package main

import (
	tpl "github.com/dimonoff/asyncapi-codegen/pkg/utils/template"
	"github.com/iancoleman/strcase"
)

// AcronymRegistry contains known acronyms/initialisms that should be preserved
// in generated Go identifiers.
type AcronymRegistry struct {
	Items []string
}

// DefaultAcronymRegistry returns the default set of known Go-style
// acronyms/initialisms that should be preserved when using strcase helpers.
func DefaultAcronymRegistry() AcronymRegistry {
	return AcronymRegistry{
		Items: []string{
			"ACL",
			"AES",
			"API",
			"ASCII",
			"AWS",
			"CLI",
			"CPU",
			"CRC",
			"CSS",
			"CSV",
			"DB",
			"DNS",
			"EOF",
			"FTP",
			"GIF",
			"GPU",
			"GUID",
			"HTML",
			"HTTP",
			"HTTPS",
			"HMAC",
			"IAM",
			"ICMP",
			"ID",
			"IP",
			"ISBN",
			"ISO",
			"JSON",
			"JWT",
			"KB",
			"LDAP",
			"LHS",
			"LLM",
			"MAC",
			"MB",
			"MFA",
			"MIME",
			"MQTT",
			"NAT",
			"NTP",
			"OAuth",
			"OIDC",
			"OTP",
			"PDF",
			"PEM",
			"PID",
			"PNG",
			"POP3",
			"PSK",
			"QPS",
			"QR",
			"RAM",
			"RHS",
			"RPC",
			"RSA",
			"S3",
			"SDK",
			"SHA",
			"SLA",
			"SMTP",
			"SMS",
			"SOA",
			"SQL",
			"SSH",
			"SSO",
			"SVG",
			"TCP",
			"TLS",
			"TTL",
			"TOTP",
			"UDP",
			"UI",
			"UID",
			"URI",
			"URL",
			"USB",
			"UTF8",
			"UUID",
			"VM",
			"VPC",
			"VPN",
			"WAF",
			"XML",
			"XMPP",
			"XSRF",
			"XSS",
			"YAML",
		},
	}
}

// Configure registers all configured acronyms globally in strcase.
func (r AcronymRegistry) Configure() {
	tpl.SetKnownAcronyms(r.Items)

	for _, acronym := range r.Items {
		strcase.ConfigureAcronym(acronym, acronym)
	}
}

