package cli_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbfake"

	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/provisioner/echo"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/rbac"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
)

func TestStatePull(t *testing.T) {
	t.Parallel()
	t.Run("File", func(t *testing.T) {
		t.Parallel()
		client, store := wirtualdtest.NewWithDatabase(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, taUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		wantState := []byte("some state")
		r := dbfake.WorkspaceBuild(t, store, database.WorkspaceTable{
			OrganizationID: owner.OrganizationID,
			OwnerID:        taUser.ID,
		}).
			Seed(database.WorkspaceBuild{ProvisionerState: wantState}).
			Do()
		statefilePath := filepath.Join(t.TempDir(), "state")
		inv, root := clitest.New(t, "state", "pull", r.Workspace.Name, statefilePath)
		clitest.SetupConfig(t, templateAdmin, root)
		err := inv.Run()
		require.NoError(t, err)
		gotState, err := os.ReadFile(statefilePath)
		require.NoError(t, err)
		require.Equal(t, wantState, gotState)
	})
	t.Run("Stdout", func(t *testing.T) {
		t.Parallel()
		client, store := wirtualdtest.NewWithDatabase(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, taUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		wantState := []byte("some state")
		r := dbfake.WorkspaceBuild(t, store, database.WorkspaceTable{
			OrganizationID: owner.OrganizationID,
			OwnerID:        taUser.ID,
		}).
			Seed(database.WorkspaceBuild{ProvisionerState: wantState}).
			Do()
		inv, root := clitest.New(t, "state", "pull", r.Workspace.Name)
		var gotState bytes.Buffer
		inv.Stdout = &gotState
		clitest.SetupConfig(t, templateAdmin, root)
		err := inv.Run()
		require.NoError(t, err)
		require.Equal(t, wantState, bytes.TrimSpace(gotState.Bytes()))
	})
	t.Run("OtherUserBuild", func(t *testing.T) {
		t.Parallel()
		client, store := wirtualdtest.NewWithDatabase(t, nil)
		owner := wirtualdtest.CreateFirstUser(t, client)
		_, taUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		wantState := []byte("some state")
		r := dbfake.WorkspaceBuild(t, store, database.WorkspaceTable{
			OrganizationID: owner.OrganizationID,
			OwnerID:        taUser.ID,
		}).
			Seed(database.WorkspaceBuild{ProvisionerState: wantState}).
			Do()
		inv, root := clitest.New(t, "state", "pull", taUser.Username+"/"+r.Workspace.Name,
			"--build", fmt.Sprintf("%d", r.Build.BuildNumber))
		var gotState bytes.Buffer
		inv.Stdout = &gotState
		//nolint: gocritic // this tests owner pulling another user's state
		clitest.SetupConfig(t, client, root)
		err := inv.Run()
		require.NoError(t, err)
		require.Equal(t, wantState, bytes.TrimSpace(gotState.Bytes()))
	})
}

func TestStatePush(t *testing.T) {
	t.Parallel()
	t.Run("File", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, templateAdmin, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)
		stateFile, err := os.CreateTemp(t.TempDir(), "")
		require.NoError(t, err)
		wantState := []byte("some magic state")
		_, err = stateFile.Write(wantState)
		require.NoError(t, err)
		err = stateFile.Close()
		require.NoError(t, err)
		inv, root := clitest.New(t, "state", "push", workspace.Name, stateFile.Name())
		clitest.SetupConfig(t, templateAdmin, root)
		err = inv.Run()
		require.NoError(t, err)
	})

	t.Run("Stdin", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, templateAdmin, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)
		inv, root := clitest.New(t, "state", "push", "--build", strconv.Itoa(int(workspace.LatestBuild.BuildNumber)), workspace.Name, "-")
		clitest.SetupConfig(t, templateAdmin, root)
		inv.Stdin = strings.NewReader("some magic state")
		err := inv.Run()
		require.NoError(t, err)
	})

	t.Run("OtherUserBuild", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, taUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, templateAdmin, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)
		inv, root := clitest.New(t, "state", "push",
			"--build", strconv.Itoa(int(workspace.LatestBuild.BuildNumber)),
			taUser.Username+"/"+workspace.Name,
			"-")
		//nolint: gocritic // this tests owner pushing another user's state
		clitest.SetupConfig(t, client, root)
		inv.Stdin = strings.NewReader("some magic state")
		err := inv.Run()
		require.NoError(t, err)
	})
}
