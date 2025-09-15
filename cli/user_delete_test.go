package cli_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/cryptorand"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/rbac"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestUserDelete(t *testing.T) {
	t.Parallel()
	t.Run("Username", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		client := wirtualdtest.New(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())

		pw, err := cryptorand.String(16)
		require.NoError(t, err)

		_, err = client.CreateUserWithOrgs(ctx, wirtualsdk.CreateUserRequestWithOrgs{
			Email:           "colin5@wirtual.dev",
			Username:        "coolin",
			Password:        pw,
			UserLoginType:   wirtualsdk.LoginTypePassword,
			OrganizationIDs: []uuid.UUID{owner.OrganizationID},
		})
		require.NoError(t, err)

		inv, root := clitest.New(t, "users", "delete", "coolin")
		clitest.SetupConfig(t, userAdmin, root)
		pty := ptytest.New(t).Attach(inv)
		errC := make(chan error)
		go func() {
			errC <- inv.Run()
		}()
		require.NoError(t, <-errC)
		pty.ExpectMatch("coolin")
	})

	t.Run("UserID", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		client := wirtualdtest.New(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())

		pw, err := cryptorand.String(16)
		require.NoError(t, err)

		user, err := client.CreateUserWithOrgs(ctx, wirtualsdk.CreateUserRequestWithOrgs{
			Email:           "colin5@wirtual.dev",
			Username:        "coolin",
			Password:        pw,
			UserLoginType:   wirtualsdk.LoginTypePassword,
			OrganizationIDs: []uuid.UUID{owner.OrganizationID},
		})
		require.NoError(t, err)

		inv, root := clitest.New(t, "users", "delete", user.ID.String())
		clitest.SetupConfig(t, userAdmin, root)
		pty := ptytest.New(t).Attach(inv)
		errC := make(chan error)
		go func() {
			errC <- inv.Run()
		}()
		require.NoError(t, <-errC)
		pty.ExpectMatch("coolin")
	})

	t.Run("UserID", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		client := wirtualdtest.New(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())

		pw, err := cryptorand.String(16)
		require.NoError(t, err)

		user, err := client.CreateUserWithOrgs(ctx, wirtualsdk.CreateUserRequestWithOrgs{
			Email:           "colin5@wirtual.dev",
			Username:        "coolin",
			Password:        pw,
			UserLoginType:   wirtualsdk.LoginTypePassword,
			OrganizationIDs: []uuid.UUID{owner.OrganizationID},
		})
		require.NoError(t, err)

		inv, root := clitest.New(t, "users", "delete", user.ID.String())
		clitest.SetupConfig(t, userAdmin, root)
		pty := ptytest.New(t).Attach(inv)
		errC := make(chan error)
		go func() {
			errC <- inv.Run()
		}()
		require.NoError(t, <-errC)
		pty.ExpectMatch("coolin")
	})

	// TODO: reenable this test case. Fetching users without perms returns a
	// "user "testuser@wirtual.dev" must be a member of at least one organization"
	// error.
	// t.Run("NoPerms", func(t *testing.T) {
	// 	t.Parallel()
	// 	ctx := context.Background()
	// 	client := wirtualdtest.New(t, nil)
	// 	aUser := wirtualdtest.CreateFirstUser(t, client)

	// 	pw, err := cryptorand.String(16)
	// 	require.NoError(t, err)

	// 	toDelete, err := client.CreateUserWithOrgs(ctx, wirtualsdk.CreateUserRequestWithOrgs{
	// 		Email:          "colin5@wirtual.dev",
	// 		Username:       "coolin",
	// 		Password:       pw,
	// 		UserLoginType:  wirtualsdk.LoginTypePassword,
	// 		OrganizationID: aUser.OrganizationID,
	// 	})
	// 	require.NoError(t, err)

	// 	uClient, _ := wirtualdtest.CreateAnotherUser(t, client, aUser.OrganizationID)
	// 	_ = uClient
	// 	_ = toDelete

	// 	inv, root := clitest.New(t, "users", "delete", "coolin")
	// 	clitest.SetupConfig(t, uClient, root)
	// 	require.ErrorContains(t, inv.Run(), "...")
	// })

	t.Run("DeleteSelf", func(t *testing.T) {
		t.Parallel()
		t.Run("Owner", func(t *testing.T) {
			client := wirtualdtest.New(t, nil)
			_ = wirtualdtest.CreateFirstUser(t, client)
			inv, root := clitest.New(t, "users", "delete", "me")
			//nolint:gocritic // The point of the test is to validate that a user cannot delete
			// themselves, the owner user is probably the most important user to test this with.
			clitest.SetupConfig(t, client, root)
			require.ErrorContains(t, inv.Run(), "You cannot delete yourself!")
		})
		t.Run("UserAdmin", func(t *testing.T) {
			client := wirtualdtest.New(t, nil)
			owner := wirtualdtest.CreateFirstUser(t, client)
			userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())
			inv, root := clitest.New(t, "users", "delete", "me")
			clitest.SetupConfig(t, userAdmin, root)
			require.ErrorContains(t, inv.Run(), "You cannot delete yourself!")
		})
	})
}
