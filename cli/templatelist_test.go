package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/rbac"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestTemplateList(t *testing.T) {
	t.Parallel()
	t.Run("ListTemplates", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		firstVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, firstVersion.ID)
		firstTemplate := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, firstVersion.ID)

		secondVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, secondVersion.ID)
		secondTemplate := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, secondVersion.ID)

		inv, root := clitest.New(t, "templates", "list")
		clitest.SetupConfig(t, templateAdmin, root)

		pty := ptytest.New(t).Attach(inv)

		ctx, cancelFunc := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancelFunc()

		errC := make(chan error)
		go func() {
			errC <- inv.WithContext(ctx).Run()
		}()

		// expect that templates are listed alphabetically
		templatesList := []string{firstTemplate.Name, secondTemplate.Name}
		sort.Strings(templatesList)

		require.NoError(t, <-errC)

		for _, name := range templatesList {
			pty.ExpectMatch(name)
		}
	})
	t.Run("ListTemplatesJSON", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		firstVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, firstVersion.ID)
		_ = wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, firstVersion.ID)

		secondVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, secondVersion.ID)
		_ = wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, secondVersion.ID)

		inv, root := clitest.New(t, "templates", "list", "--output=json")
		clitest.SetupConfig(t, templateAdmin, root)

		ctx, cancelFunc := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancelFunc()

		out := bytes.NewBuffer(nil)
		inv.Stdout = out
		err := inv.WithContext(ctx).Run()
		require.NoError(t, err)

		var templates []wirtualsdk.Template
		require.NoError(t, json.Unmarshal(out.Bytes(), &templates))
		require.Len(t, templates, 2)
	})
	t.Run("NoTemplates", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{})
		owner := wirtualdtest.CreateFirstUser(t, client)

		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

		inv, root := clitest.New(t, "templates", "list")
		clitest.SetupConfig(t, templateAdmin, root)

		pty := ptytest.New(t)
		inv.Stdin = pty.Input()
		inv.Stderr = pty.Output()

		ctx, cancelFunc := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancelFunc()

		errC := make(chan error)
		go func() {
			errC <- inv.WithContext(ctx).Run()
		}()

		require.NoError(t, <-errC)

		pty.ExpectMatch("No templates found")
		pty.ExpectMatch("Create one:")
	})
}
