package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/rbac"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestUserList(t *testing.T) {
	t.Parallel()
	t.Run("Table", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())
		inv, root := clitest.New(t, "users", "list")
		clitest.SetupConfig(t, userAdmin, root)
		pty := ptytest.New(t).Attach(inv)
		errC := make(chan error)
		go func() {
			errC <- inv.Run()
		}()
		require.NoError(t, <-errC)
		pty.ExpectMatch("wirtual.dev")
	})
	t.Run("JSON", func(t *testing.T) {
		t.Parallel()

		client := wirtualdtest.New(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())
		inv, root := clitest.New(t, "users", "list", "-o", "json")
		clitest.SetupConfig(t, userAdmin, root)
		doneChan := make(chan struct{})

		buf := bytes.NewBuffer(nil)
		inv.Stdout = buf
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		<-doneChan

		var users []wirtualsdk.User
		err := json.Unmarshal(buf.Bytes(), &users)
		require.NoError(t, err, "unmarshal JSON output")
		require.Len(t, users, 2)
		for _, u := range users {
			assert.NotEmpty(t, u.ID)
			assert.NotEmpty(t, u.Email)
			assert.NotEmpty(t, u.Username)
			assert.NotEmpty(t, u.Name)
			assert.NotEmpty(t, u.CreatedAt)
			assert.NotEmpty(t, u.Status)
		}
	})
	t.Run("NoURLFileErrorHasHelperText", func(t *testing.T) {
		t.Parallel()

		inv, _ := clitest.New(t, "users", "list")
		err := inv.Run()
		require.Contains(t, err.Error(), "Try logging in using 'wirtual login <url>'.")
	})
	t.Run("SessionAuthErrorHasHelperText", func(t *testing.T) {
		t.Parallel()

		client := wirtualdtest.New(t, nil)
		inv, root := clitest.New(t, "users", "list")
		clitest.SetupConfig(t, client, root)

		err := inv.Run()

		var apiErr *wirtualsdk.Error
		require.ErrorAs(t, err, &apiErr)
		require.Contains(t, err.Error(), "Try logging in using 'wirtual login'.")
	})
}

func TestUserShow(t *testing.T) {
	t.Parallel()

	t.Run("Table", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())
		_, otherUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		inv, root := clitest.New(t, "users", "show", otherUser.Username)
		clitest.SetupConfig(t, userAdmin, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()
		pty.ExpectMatch(otherUser.Email)
		<-doneChan
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		client := wirtualdtest.New(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		userAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleUserAdmin())
		other, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		otherUser, err := other.User(ctx, wirtualsdk.Me)
		require.NoError(t, err, "fetch other user")
		inv, root := clitest.New(t, "users", "show", otherUser.Username, "-o", "json")
		clitest.SetupConfig(t, userAdmin, root)
		doneChan := make(chan struct{})

		buf := bytes.NewBuffer(nil)
		inv.Stdout = buf
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		<-doneChan

		var newUser wirtualsdk.User
		err = json.Unmarshal(buf.Bytes(), &newUser)
		require.NoError(t, err, "unmarshal JSON output")
		require.Equal(t, otherUser.ID, newUser.ID)
		require.Equal(t, otherUser.Username, newUser.Username)
		require.Equal(t, otherUser.Email, newUser.Email)
		require.Equal(t, otherUser.Name, newUser.Name)
	})
}
