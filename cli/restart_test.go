package cli_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/provisioner/echo"
	"github.com/wirtualdev/wirtualdev/v2/provisionersdk/proto"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestRestart(t *testing.T) {
	t.Parallel()

	echoResponses := prepareEchoResponses([]*proto.RichParameter{
		{
			Name:        ephemeralParameterName,
			Description: ephemeralParameterDescription,
			Mutable:     true,
			Ephemeral:   true,
		},
	})

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

		ctx := testutil.Context(t, testutil.WaitLong)

		inv, root := clitest.New(t, "restart", workspace.Name, "--yes")
		clitest.SetupConfig(t, member, root)

		pty := ptytest.New(t).Attach(inv)

		done := make(chan error, 1)
		go func() {
			done <- inv.WithContext(ctx).Run()
		}()
		pty.ExpectMatch("Stopping workspace")
		pty.ExpectMatch("Starting workspace")
		pty.ExpectMatch("workspace has been restarted")

		err := <-done
		require.NoError(t, err, "execute failed")
	})

	t.Run("PromptEphemeralParameters", func(t *testing.T) {
		t.Parallel()

		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, memberUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, echoResponses)
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

		inv, root := clitest.New(t, "restart", workspace.Name, "--prompt-ephemeral-parameters")
		clitest.SetupConfig(t, member, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			ephemeralParameterDescription, ephemeralParameterValue,
			"Restart workspace?", "yes",
			"Stopping workspace", "",
			"Starting workspace", "",
			"workspace has been restarted", "",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)

			if value != "" {
				pty.WriteLine(value)
			}
		}
		<-doneChan

		// Verify if build option is set
		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
		defer cancel()

		workspace, err := client.WorkspaceByOwnerAndName(ctx, memberUser.ID.String(), workspace.Name, wirtualsdk.WorkspaceOptions{})
		require.NoError(t, err)
		actualParameters, err := client.WorkspaceBuildParameters(ctx, workspace.LatestBuild.ID)
		require.NoError(t, err)
		require.Contains(t, actualParameters, wirtualsdk.WorkspaceBuildParameter{
			Name:  ephemeralParameterName,
			Value: ephemeralParameterValue,
		})
	})

	t.Run("EphemeralParameterFlags", func(t *testing.T) {
		t.Parallel()

		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, memberUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, echoResponses)
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

		inv, root := clitest.New(t, "restart", workspace.Name,
			"--ephemeral-parameter", fmt.Sprintf("%s=%s", ephemeralParameterName, ephemeralParameterValue))
		clitest.SetupConfig(t, member, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			"Restart workspace?", "yes",
			"Stopping workspace", "",
			"Starting workspace", "",
			"workspace has been restarted", "",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)

			if value != "" {
				pty.WriteLine(value)
			}
		}
		<-doneChan

		// Verify if build option is set
		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
		defer cancel()

		workspace, err := client.WorkspaceByOwnerAndName(ctx, memberUser.ID.String(), workspace.Name, wirtualsdk.WorkspaceOptions{})
		require.NoError(t, err)
		actualParameters, err := client.WorkspaceBuildParameters(ctx, workspace.LatestBuild.ID)
		require.NoError(t, err)
		require.Contains(t, actualParameters, wirtualsdk.WorkspaceBuildParameter{
			Name:  ephemeralParameterName,
			Value: ephemeralParameterValue,
		})
	})

	t.Run("with deprecated build-options flag", func(t *testing.T) {
		t.Parallel()

		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, memberUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, echoResponses)
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

		inv, root := clitest.New(t, "restart", workspace.Name, "--build-options")
		clitest.SetupConfig(t, member, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			ephemeralParameterDescription, ephemeralParameterValue,
			"Restart workspace?", "yes",
			"Stopping workspace", "",
			"Starting workspace", "",
			"workspace has been restarted", "",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)

			if value != "" {
				pty.WriteLine(value)
			}
		}
		<-doneChan

		// Verify if build option is set
		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
		defer cancel()

		workspace, err := client.WorkspaceByOwnerAndName(ctx, memberUser.ID.String(), workspace.Name, wirtualsdk.WorkspaceOptions{})
		require.NoError(t, err)
		actualParameters, err := client.WorkspaceBuildParameters(ctx, workspace.LatestBuild.ID)
		require.NoError(t, err)
		require.Contains(t, actualParameters, wirtualsdk.WorkspaceBuildParameter{
			Name:  ephemeralParameterName,
			Value: ephemeralParameterValue,
		})
	})

	t.Run("with deprecated build-option flag", func(t *testing.T) {
		t.Parallel()

		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, memberUser := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, echoResponses)
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID)
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

		inv, root := clitest.New(t, "restart", workspace.Name,
			"--build-option", fmt.Sprintf("%s=%s", ephemeralParameterName, ephemeralParameterValue))
		clitest.SetupConfig(t, member, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			"Restart workspace?", "yes",
			"Stopping workspace", "",
			"Starting workspace", "",
			"workspace has been restarted", "",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)

			if value != "" {
				pty.WriteLine(value)
			}
		}
		<-doneChan

		// Verify if build option is set
		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
		defer cancel()

		workspace, err := client.WorkspaceByOwnerAndName(ctx, memberUser.ID.String(), workspace.Name, wirtualsdk.WorkspaceOptions{})
		require.NoError(t, err)
		actualParameters, err := client.WorkspaceBuildParameters(ctx, workspace.LatestBuild.ID)
		require.NoError(t, err)
		require.Contains(t, actualParameters, wirtualsdk.WorkspaceBuildParameter{
			Name:  ephemeralParameterName,
			Value: ephemeralParameterValue,
		})
	})
}

func TestRestartWithParameters(t *testing.T) {
	t.Parallel()

	echoResponses := &echo.Responses{
		Parse: echo.ParseComplete,
		ProvisionPlan: []*proto.Response{
			{
				Type: &proto.Response_Plan{
					Plan: &proto.PlanComplete{
						Parameters: []*proto.RichParameter{
							{
								Name:        immutableParameterName,
								Description: immutableParameterDescription,
								Required:    true,
							},
						},
					},
				},
			},
		},
		ProvisionApply: echo.ApplyComplete,
	}

	t.Run("DoNotAskForImmutables", func(t *testing.T) {
		t.Parallel()

		// Create the workspace
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, echoResponses)
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID, func(cwr *wirtualsdk.CreateWorkspaceRequest) {
			cwr.RichParameterValues = []wirtualsdk.WorkspaceBuildParameter{
				{
					Name:  immutableParameterName,
					Value: immutableParameterValue,
				},
			}
		})
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

		// Restart the workspace again
		inv, root := clitest.New(t, "restart", workspace.Name, "-y")
		clitest.SetupConfig(t, member, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		pty.ExpectMatch("workspace has been restarted")
		<-doneChan

		// Verify if immutable parameter is set
		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
		defer cancel()

		workspace, err := client.WorkspaceByOwnerAndName(ctx, workspace.OwnerName, workspace.Name, wirtualsdk.WorkspaceOptions{})
		require.NoError(t, err)
		actualParameters, err := client.WorkspaceBuildParameters(ctx, workspace.LatestBuild.ID)
		require.NoError(t, err)
		require.Contains(t, actualParameters, wirtualsdk.WorkspaceBuildParameter{
			Name:  immutableParameterName,
			Value: immutableParameterValue,
		})
	})

	t.Run("AlwaysPrompt", func(t *testing.T) {
		t.Parallel()

		// Create the workspace
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		member, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID)
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, mutableParamsResponse)
		wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)
		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		workspace := wirtualdtest.CreateWorkspace(t, member, template.ID, func(cwr *wirtualsdk.CreateWorkspaceRequest) {
			cwr.RichParameterValues = []wirtualsdk.WorkspaceBuildParameter{
				{
					Name:  mutableParameterName,
					Value: mutableParameterValue,
				},
			}
		})
		wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, client, workspace.LatestBuild.ID)

		inv, root := clitest.New(t, "restart", workspace.Name, "-y", "--always-prompt")
		clitest.SetupConfig(t, member, root)
		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		// We should be prompted for the parameters again.
		newValue := "xyz"
		pty.ExpectMatch(mutableParameterName)
		pty.WriteLine(newValue)
		pty.ExpectMatch("workspace has been restarted")
		<-doneChan

		// Verify that the updated values are persisted.
		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
		defer cancel()

		workspace, err := client.WorkspaceByOwnerAndName(ctx, workspace.OwnerName, workspace.Name, wirtualsdk.WorkspaceOptions{})
		require.NoError(t, err)
		actualParameters, err := client.WorkspaceBuildParameters(ctx, workspace.LatestBuild.ID)
		require.NoError(t, err)
		require.Contains(t, actualParameters, wirtualsdk.WorkspaceBuildParameter{
			Name:  mutableParameterName,
			Value: newValue,
		})
	})
}
