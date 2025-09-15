package cli_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
)

func TestRename(t *testing.T) {
	t.Parallel()

	client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true, AllowWorkspaceRenames: true})
	owner := wirtualdtest.CreateFirstUser(t, client)
	member, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
	version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
	wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
	template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
	workspace := wirtualdtest.CreateWorkspace(t, member, template.ID)
	wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()

	want := wirtualdtest.RandomUsername(t)
	inv, root := clitest.New(t, "rename", workspace.Name, want, "--yes")
	clitest.SetupConfig(t, member, root)
	pty := ptytest.New(t)
	pty.Attach(inv)
	clitest.Start(t, inv)

	pty.ExpectMatch("confirm rename:")
	pty.WriteLine(workspace.Name)
	pty.ExpectMatch("renamed to")

	ws, err := client.Workspace(ctx, workspace.ID)
	assert.NoError(t, err)

	got := ws.Name
	assert.Equal(t, want, got, "workspace name did not change")
}
