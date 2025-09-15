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
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbfake"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestList(t *testing.T) {
	t.Parallel()
	t.Run("Single", func(t *testing.T) {
		t.Parallel()
		client, db := wirtualdtest.NewWithDatabase(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, memberUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		// setup template
		r := dbfake.WorkspaceBuild(t, db, database.WorkspaceTable{
			OrganizationID: owner.OrganizationID,
			OwnerID:        memberUser.ID,
		}).WithAgent().Do()

		inv, root := clitest.New(t, "ls")
		clitest.SetupConfig(t, member, root)
		pty := ptytest.New(t).Attach(inv)

		ctx, cancelFunc := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancelFunc()
		done := make(chan any)
		go func() {
			errC := inv.WithContext(ctx).Run()
			assert.NoError(t, errC)
			close(done)
		}()
		pty.ExpectMatch(r.Workspace.Name)
		pty.ExpectMatch("Started")
		cancelFunc()
		<-done
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		client, db := wirtualdtest.NewWithDatabase(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, memberUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		_ = dbfake.WorkspaceBuild(t, db, database.WorkspaceTable{
			OrganizationID: owner.OrganizationID,
			OwnerID:        memberUser.ID,
		}).WithAgent().Do()

		inv, root := clitest.New(t, "list", "--output=json")
		clitest.SetupConfig(t, member, root)

		ctx, cancelFunc := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancelFunc()

		out := bytes.NewBuffer(nil)
		inv.Stdout = out
		err := inv.WithContext(ctx).Run()
		require.NoError(t, err)

		var workspaces []wirtualsdk.Workspace
		require.NoError(t, json.Unmarshal(out.Bytes(), &workspaces))
		require.Len(t, workspaces, 1)
	})
}
