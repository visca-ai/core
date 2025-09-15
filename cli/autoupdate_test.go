package cli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestAutoUpdate(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)
		require.Equal(t, wirtualsdk.AutomaticUpdatesNever, workspace.AutomaticUpdates)

		expectedPolicy := wirtualsdk.AutomaticUpdatesAlways
		inv, root := clitest.New(t, "autoupdate", workspace.Name, string(expectedPolicy))
		clitest.SetupConfig(t, member, root)
		var buf bytes.Buffer
		inv.Stdout = &buf
		err := inv.Run()
		require.NoError(t, err)
		require.Contains(t, buf.String(), fmt.Sprintf("Updated workspace %q auto-update policy to %q", workspace.Name, expectedPolicy))

		workspace = wirtualdtest.MustWorkspace(t, client, workspace.ID)
		require.Equal(t, expectedPolicy, workspace.AutomaticUpdates)
	})

	t.Run("InvalidArgs", func(t *testing.T) {
		type testcase struct {
			Name          string
			Args          []string
			ErrorContains string
		}

		cases := []testcase{
			{
				Name:          "NoPolicy",
				Args:          []string{"autoupdate", "ws"},
				ErrorContains: "wanted 2 args but got 1",
			},
			{
				Name:          "InvalidPolicy",
				Args:          []string{"autoupdate", "ws", "sometimes"},
				ErrorContains: `invalid option "sometimes" must be either of`,
			},
		}

		for _, c := range cases {
			c := c
			t.Run(c.Name, func(t *testing.T) {
				t.Parallel()
				client := wirtualdtest.New(t, nil)
				_ = wirtualdtest.CreateFirstUser(t, client)

				inv, root := clitest.New(t, c.Args...)
				clitest.SetupConfig(t, client, root)
				err := inv.Run()
				require.Error(t, err)
				require.Contains(t, err.Error(), c.ErrorContains)
			})
		}
	})
}
