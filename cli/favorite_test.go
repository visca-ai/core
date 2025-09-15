package cli_test

import (
	"bytes"
	"testing"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbfake"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"

	"github.com/stretchr/testify/require"
)

func TestFavoriteUnfavorite(t *testing.T) {
	t.Parallel()

	var (
		client, db           = wirtualdtest.NewWithDatabase(t, nil)
		owner                = wirtualdtest.CreateFirstUser(t, client)
		memberClient, member = wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		ws                   = dbfake.WorkspaceBuild(t, db, database.WorkspaceTable{OwnerID: member.ID, OrganizationID: owner.OrganizationID}).Do()
	)

	inv, root := clitest.New(t, "favorite", ws.Workspace.Name)
	clitest.SetupConfig(t, memberClient, root)

	var buf bytes.Buffer
	inv.Stdout = &buf
	err := inv.Run()
	require.NoError(t, err)

	updated := wirtualdtest.MustWorkspace(t, memberClient, ws.Workspace.ID)
	require.True(t, updated.Favorite)

	buf.Reset()

	inv, root = clitest.New(t, "unfavorite", ws.Workspace.Name)
	clitest.SetupConfig(t, memberClient, root)
	inv.Stdout = &buf
	err = inv.Run()
	require.NoError(t, err)
	updated = wirtualdtest.MustWorkspace(t, memberClient, ws.Workspace.ID)
	require.False(t, updated.Favorite)
}
