package domain_test

import (
	"fmt"
	"testing"

	"github.com/dstdfx/twbridge/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestExtractMsgJid(t *testing.T) {
	tableTest := []struct{
		input string
		expected string
	}{
		{
			input: "",
			expected: "",
		},
		{
			input: fmt.Sprintf(domain.TextMessageFmt, "test user", "test@jid.net", "hello, world!"),
			expected: "test@jid.net",
		},
		{
			input: "[jid:]",
			expected: "",
		},
		{
			input: "][jid:]",
			expected: "",
		},
		{
			input: "jid: ",
			expected: "",
		},
		{
			input: "[jid: dasd@asdsad.com",
			expected: "",
		},
		{
			input: "ewfwefwef",
			expected: "",
		},
	}

	for _, test := range tableTest {
		assert.Equal(t, test.expected, domain.ExtractMsgJid(test.input))
	}
}
