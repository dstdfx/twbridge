package domain

import "strings"

// ExtractMsgJid returns remote jid from the message if it has one,
// otherwise - empty string.
func ExtractMsgJid(message string) string {
	jidStart := strings.Index(message, "jid:")
	jidEnd := strings.Index(message, "]")
	if jidStart == -1 || jidEnd == -1 || jidStart+5 >= jidEnd {
		return ""
	}

	return message[jidStart+5 : jidEnd] // jid: + space
}
