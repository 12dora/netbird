package blockeduser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	t.Run("default message when env unset", func(t *testing.T) {
		t.Setenv(MessageEnv, "")
		assert.Equal(t, "user is blocked", Message("user is blocked"))
	})

	t.Run("default message when env is blank", func(t *testing.T) {
		t.Setenv(MessageEnv, "   ")
		assert.Equal(t, "user is blocked", Message("user is blocked"))
	})

	t.Run("env overrides default", func(t *testing.T) {
		t.Setenv(MessageEnv, "无 VPN 权限，请前往 https://iam.example.com 申请")
		assert.Equal(t, "无 VPN 权限，请前往 https://iam.example.com 申请", Message("user is blocked"))
	})
}
