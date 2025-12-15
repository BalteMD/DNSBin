package notify

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestRenderInitialMsg(t *testing.T) {
	msg := &InitialMessage{
		Version:  "1.2.3",
		Interval: 10 * time.Second,
		DNSLog:   "kiki.ringing.com",
		HTTPLog:  []string{"https://cookcu.global.ssl.fastly.net/kiki/{payload}", "https://congcong.cloudfront.net/kiki/{payload}"},
	}
	result := RenderInitialMsg(msg)

	fmt.Println(result)

	assert.Contains(t, result, "1.2.3")
	assert.Contains(t, result, "10s")
	assert.True(t, strings.Contains(result, "DNS&HTTP Log initial successfully"))
}

func TestRenderDNSInfo(t *testing.T) {
	dnsLog := &DNSLogInfo{
		Type:           "HTTP",
		DnsQueryRecord: "config",
		IpAddress:      "1.2.3.4",
		Time:           time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
	result := RenderDNSInfo(dnsLog)

	fmt.Println(result)

	assert.Contains(t, result, "config")
	assert.Contains(t, result, "1.2.3.4")
	assert.Contains(t, result, "#HTTP")
	assert.Contains(t, result, "📝 **Record:**")
	assert.Contains(t, result, "🤖 **IP Address:**")
	assert.Contains(t, result, "⏰ **Time:**")
}

func TestEscapeMarkdown(t *testing.T) {
	testCases := []struct {
		name             string
		inputDescription string
		expected         string
	}{
		{
			name:             "escape underscores",
			inputDescription: "I Doc View. In November 2023, the official released version 13.10.1_20231115, fixing related vulnerabilities.",
			expected:         "I Doc View. In November 2023, the official released version 13.10.1\\_20231115, fixing related vulnerabilities.",
		},
		{
			name:             "escape asterisks",
			inputDescription: "This is not a *bold text",
			expected:         "This is not a \\*bold text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeMarkdown(tc.inputDescription)
			assert.Equal(t, tc.expected, result)
		})
	}
}
