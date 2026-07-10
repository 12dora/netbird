package server

import (
	"os"
	"strings"
)

// blockedUserMessageEnv lets operators replace the default rejection text returned when a
// blocked or pending-approval user is refused peer registration. Clients and the dashboard
// render this server-side text as-is, so it can carry a hint such as the URL of the access
// request portal. When unset, the stock messages are kept.
const blockedUserMessageEnv = "NB_BLOCKED_USER_MESSAGE"

func blockedUserMessage(defaultMessage string) string {
	if msg := strings.TrimSpace(os.Getenv(blockedUserMessageEnv)); msg != "" {
		return msg
	}
	return defaultMessage
}
