package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
)

func TestUserCreate(t *testing.T) {
	t.Parallel()
	t.Run("Prompts", func(t *testing.T) {
		t.Parallel()
		ctx := testutil.Context(t, testutil.WaitLong)
		client := wirtualdtest.New(t, nil)
		wirtualdtest.CreateFirstUser(t, client)
		inv, root := clitest.New(t, "users", "create")
		clitest.SetupConfig(t, client, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()
		matches := []string{
			"Username", "dean",
			"Email", "dean@wirtual.dev",
			"Full name (optional):", "Mr. Dean Deanington",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)
			pty.WriteLine(value)
		}
		_ = testutil.RequireRecvCtx(ctx, t, doneChan)
		created, err := client.User(ctx, matches[1])
		require.NoError(t, err)
		assert.Equal(t, matches[1], created.Username)
		assert.Equal(t, matches[3], created.Email)
		assert.Equal(t, matches[5], created.Name)
	})

	t.Run("PromptsNoName", func(t *testing.T) {
		t.Parallel()
		ctx := testutil.Context(t, testutil.WaitLong)
		client := wirtualdtest.New(t, nil)
		wirtualdtest.CreateFirstUser(t, client)
		inv, root := clitest.New(t, "users", "create")
		clitest.SetupConfig(t, client, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()
		matches := []string{
			"Username", "noname",
			"Email", "noname@wirtual.dev",
			"Full name (optional):", "",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)
			pty.WriteLine(value)
		}
		_ = testutil.RequireRecvCtx(ctx, t, doneChan)
		created, err := client.User(ctx, matches[1])
		require.NoError(t, err)
		assert.Equal(t, matches[1], created.Username)
		assert.Equal(t, matches[3], created.Email)
		assert.Empty(t, created.Name)
	})

	t.Run("Args", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		wirtualdtest.CreateFirstUser(t, client)
		args := []string{
			"users", "create",
			"-e", "dean@wirtual.dev",
			"-u", "dean",
			"-n", "Mr. Dean Deanington",
			"-p", "1n5ecureP4ssw0rd!",
		}
		inv, root := clitest.New(t, args...)
		clitest.SetupConfig(t, client, root)
		err := inv.Run()
		require.NoError(t, err)
		ctx := testutil.Context(t, testutil.WaitShort)
		created, err := client.User(ctx, "dean")
		require.NoError(t, err)
		assert.Equal(t, args[3], created.Email)
		assert.Equal(t, args[5], created.Username)
		assert.Equal(t, args[7], created.Name)
	})

	t.Run("ArgsNoName", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		wirtualdtest.CreateFirstUser(t, client)
		args := []string{
			"users", "create",
			"-e", "dean@wirtual.dev",
			"-u", "dean",
			"-p", "1n5ecureP4ssw0rd!",
		}
		inv, root := clitest.New(t, args...)
		clitest.SetupConfig(t, client, root)
		err := inv.Run()
		require.NoError(t, err)
		ctx := testutil.Context(t, testutil.WaitShort)
		created, err := client.User(ctx, args[5])
		require.NoError(t, err)
		assert.Equal(t, args[3], created.Email)
		assert.Equal(t, args[5], created.Username)
		assert.Empty(t, created.Name)
	})
}
