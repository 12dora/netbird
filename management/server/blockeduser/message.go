// Package blockeduser centralises the rejection text shown to blocked and
// pending-approval users. It is a leaf package so both the account manager and
// the reverse proxy auth handler can share one implementation without either
// importing the other.
package blockeduser

import (
	"os"
	"strings"
)

// MessageEnv lets operators replace the default rejection text returned when a blocked or
// pending-approval user is refused access. Clients, the dashboard and the reverse proxy
// access-denied page render this server-side text as-is, so it can carry a hint such as the
// URL of the access request portal. When unset, the stock messages are kept.
//
// The override deliberately collapses the distinct "blocked" and "pending approval" wordings
// into one message: an operator pointing users at a request portal wants the same call to
// action in both states.
const MessageEnv = "NB_BLOCKED_USER_MESSAGE"

// Message returns the operator-supplied rejection text, falling back to defaultMessage when
// the override is unset or blank. Callers must only use it for blocked/pending-approval
// denials, never for unrelated failures.
func Message(defaultMessage string) string {
	if msg := strings.TrimSpace(os.Getenv(MessageEnv)); msg != "" {
		return msg
	}
	return defaultMessage
}
