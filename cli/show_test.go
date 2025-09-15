package cli_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
)

func TestShow(t *testing.T) {
	t.Parallel()
	t.Run("Exists", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, completeWithAgent())
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

		args := []string{
			"show",
			workspace.Name,
		}
		inv, root := clitest.New(t, args...)
		clitest.SetupConfig(t, member, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()
		matches := []struct {
			match string
			write string
		}{
			{match: "compute.main"},
			{match: "smith (linux, i386)"},
			{match: "wirtual ssh " + workspace.Name},
		}
		for _, m := range matches {
			pty.ExpectMatch(m.match)
			if len(m.write) > 0 {
				pty.WriteLine(m.write)
			}
		}
		<-doneChan
	})
}
