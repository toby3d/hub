package common

const (
	MIMEApplicationForm            string = "application/x-www-form-urlencoded"
	MIMEApplicationFormCharsetUTF8 string = MIMEApplicationForm + "; " + charsetUTF8
	MIMETextHTML                   string = "text/html"
	MIMETextHTMLCharsetUTF8        string = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                  string = "text/plain"
	MIMETextPlainCharsetUTF8       string = MIMETextPlain + "; " + charsetUTF8
)

const (
	HeaderAcceptLanguage string = "Accept-Language"
	HeaderContentType    string = "Content-Type"
	HeaderLink           string = "Link"
	HeaderXHubSignature  string = "X-Hub-Signature"
)

const (
	HubCallback     string = hub + ".callback"
	HubChallenge    string = hub + ".challenge"
	HubLeaseSeconds string = hub + ".lease_seconds"
	HubMode         string = hub + ".mode"
	HubReason       string = hub + ".reason"
	HubSecret       string = hub + ".secret"
	HubTopic        string = hub + ".topic"
	HubURL          string = hub + ".url"
)

const Und string = "und"

const (
	hub         string = "hub"
	charsetUTF8 string = "charset=UTF-8"
)
