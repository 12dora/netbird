package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netbirdio/netbird/management/server/activity"
	"github.com/netbirdio/netbird/management/server/blockeduser"
	"github.com/netbirdio/netbird/management/server/permissions"
	"github.com/netbirdio/netbird/management/server/permissions/modules"
	"github.com/netbirdio/netbird/management/server/permissions/operations"
	"github.com/netbirdio/netbird/management/server/store"
	"github.com/netbirdio/netbird/management/server/types"
	"github.com/netbirdio/netbird/shared/auth"
)

// Blocked and pending-approval users reach the dashboard through GetCurrentUserInfo and the
// REST permission gate, neither of which goes through peer registration. Both must therefore
// honour NB_BLOCKED_USER_MESSAGE, or an operator who configured an access request hint still
// sees the stock English text in the dashboard.
func TestBlockedUserMessageReachesRESTPaths(t *testing.T) {
	const override = "无 VPN 权限，请前往 https://iam.example.com 申请"

	newManager := func(t *testing.T) DefaultAccountManager {
		t.Helper()

		s, cleanup, err := store.NewTestStoreFromSQL(context.Background(), "", t.TempDir())
		require.NoError(t, err)
		t.Cleanup(cleanup)

		account := newAccountWithId(context.Background(), "acc", "accOwner", "", "", "", false)
		account.Users["blocked-user"] = &types.User{
			Id:        "blocked-user",
			AccountID: account.Id,
			Role:      types.UserRoleUser,
			Blocked:   true,
		}
		account.Users["pending-user"] = &types.User{
			Id:              "pending-user",
			AccountID:       account.Id,
			Role:            types.UserRoleUser,
			Blocked:         true,
			PendingApproval: true,
		}
		require.NoError(t, s.SaveAccount(context.Background(), account))

		return DefaultAccountManager{
			Store:              s,
			eventStore:         &activity.InMemoryEventStore{},
			permissionsManager: permissions.NewManager(s),
		}
	}

	t.Run("GetCurrentUserInfo uses the override", func(t *testing.T) {
		t.Setenv(blockeduser.MessageEnv, override)
		am := newManager(t)

		_, err := am.GetCurrentUserInfo(context.Background(), auth.UserAuth{AccountId: "acc", UserId: "blocked-user"})
		require.Error(t, err)
		assert.Equal(t, override, err.Error())
	})

	t.Run("GetCurrentUserInfo keeps the default when unset", func(t *testing.T) {
		t.Setenv(blockeduser.MessageEnv, "")
		am := newManager(t)

		_, err := am.GetCurrentUserInfo(context.Background(), auth.UserAuth{AccountId: "acc", UserId: "blocked-user"})
		require.Error(t, err)
		assert.Equal(t, "user is blocked", err.Error())
	})

	t.Run("permission gate uses the override for blocked and pending users", func(t *testing.T) {
		t.Setenv(blockeduser.MessageEnv, override)
		am := newManager(t)

		for _, userID := range []string{"blocked-user", "pending-user"} {
			_, _, err := am.permissionsManager.ValidateUserPermissions(
				context.Background(), "acc", userID, modules.Peers, operations.Read,
			)
			require.Error(t, err, userID)
			assert.Equal(t, override, err.Error(), userID)
		}
	})
}
