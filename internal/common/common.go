package common

const (
	MIMEApplicationForm            = "application/x-www-form-urlencoded"
	MIMEApplicationFormCharsetUTF8 = MIMEApplicationForm + "; " + charsetUTF8
	MIMETextHTML                   = "text/html"
	MIMETextHTMLCharsetUTF8        = MIMETextHTML + "; " + charsetUTF8

	charsetUTF8 = "charset=UTF-8"
)

const (
	HeaderContentType    = "Content-Type"
	HeaderLink           = "Link"
	HeaderXHubSignature  = "X-Hub-Signature"
	HeaderAcceptLanguage = "Accept-Language"
)

const (
	HubCallback     = hub + ".callback"
	HubChallenge    = hub + ".challenge"
	HubLeaseSeconds = hub + ".lease_seconds"
	HubMode         = hub + ".mode"
	HubReason       = hub + ".reason"
	HubSecret       = hub + ".secret"
	HubTopic        = hub + ".topic"
	HubURL          = hub + ".url"

	hub = "hub"
)

const Und = "und"
