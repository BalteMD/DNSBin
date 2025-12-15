package notify

import (
	"strings"
	"text/template"
	"time"
)

const vulnInfoMsg = "📝 *Record:* `{{ .DnsQueryRecord }}`\n" +
	"🤖 *IP Address:* `{{ .IpAddress }}`\n" +
	"📍 *Location:* `{{ .Location }}`\n" +
	"⏰ *Time:* `{{ .Time }}`\n\n" +
	"*#{{ .Type }}*\n"

const initialMsg = `
*🔥 DNS&HTTP Log initial successfully*

*⚙️ Version: {{ .Version }}*

*⏲️ Interval: {{ .Interval }}*

🌐 *HTTP log*
{{ range .HTTPLog }}
` + "`{{ . }}/httplog/{payload}`" + `
{{ end }}
 🌐 *Exchange log*
{{ range .HTTPLog }}
` + "`{{ . }}/users/login?username={username}&password={password}`" + `
{{ end }}
🕸 *DNS log*

` + "`{payload}.{{ .DNSLog }}`" + `

*@dunghm19*
`

var (
	funcMap = template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
	}

	dnsLogInfoMsgTpl = template.Must(template.New("markdown").Funcs(funcMap).Parse(vulnInfoMsg))
	initialMsgTpl    = template.Must(template.New("markdown").Funcs(funcMap).Parse(initialMsg))
)

// DNSLogInfo represents a notification message
type DNSLogInfo struct {
	Type           string    `json:"type"`
	DnsQueryRecord string    `json:"description"`
	IpAddress      string    `json:"ip_address"`
	Location       string    `json:"location"`
	Time           time.Time `json:"time"`
}

const (
	maxDnsQueryRecordLength = 500
)

func RenderDNSInfo(d *DNSLogInfo) string {
	var builder strings.Builder
	runeDescription := []rune(d.DnsQueryRecord)
	if len(runeDescription) > maxDnsQueryRecordLength {
		d.DnsQueryRecord = string(runeDescription[:maxDnsQueryRecordLength]) + "..."
	}

	d.DnsQueryRecord = escapeMarkdown(d.DnsQueryRecord)

	if err := dnsLogInfoMsgTpl.Execute(&builder, d); err != nil {
		return err.Error()
	}
	return builder.String()
}

func RenderInitialMsg(d *InitialMessage) string {
	var builder strings.Builder
	if err := initialMsgTpl.Execute(&builder, d); err != nil {
		return err.Error()
	}
	return builder.String()
}

type InitialMessage struct {
	Version  string        `json:"version"`
	Interval time.Duration `json:"interval"`
	HTTPLog  []string      `json:"http_log"`
	DNSLog   string        `json:"dns_log"`
}

type TextMessage struct {
	Message string `json:"message"`
}

const (
	RawMessageTypeInitial  = "dnslog-initial"
	RawMessageTypeText     = "dnslog-text"
	RawMessageTypeVulnInfo = "dnslog-monitor"
)

type RawMessage struct {
	Content any    `json:"content"`
	Type    string `json:"type"`
}

func NewRawInitialMessage(m *InitialMessage) *RawMessage {
	return &RawMessage{
		Content: m,
		Type:    RawMessageTypeInitial,
	}
}

func NewRawTextMessage(m string) *RawMessage {
	return &RawMessage{
		Content: &TextMessage{Message: m},
		Type:    RawMessageTypeText,
	}
}

func NewRawVulnInfoMessage(m *DNSLogInfo) *RawMessage {
	return &RawMessage{
		Content: m,
		Type:    RawMessageTypeVulnInfo,
	}
}

// escapeMarkdown escapes the special characters in the Markdown text.
func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
