package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockedUserMessage(t *testing.T) {
	t.Run("default message when env unset", func(t *testing.T) {
		t.Setenv(blockedUserMessageEnv, "")
		assert.Equal(t, "user is blocked", blockedUserMessage("user is blocked"))
	})

	t.Run("default message when env is blank", func(t *testing.T) {
		t.Setenv(blockedUserMessageEnv, "   ")
		assert.Equal(t, "user is blocked", blockedUserMessage("user is blocked"))
	})

	t.Run("env overrides default", func(t *testing.T) {
		t.Setenv(blockedUserMessageEnv, "无 VPN 权限，请前往 https://iam.example.com 申请")
		assert.Equal(t, "无 VPN 权限，请前往 https://iam.example.com 申请", blockedUserMessage("user is blocked"))
	})
}
