package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/provisioner/echo"
	"github.com/wirtualdev/wirtualdev/v2/provisioner/terraform/tfparse"
	"github.com/wirtualdev/wirtualdev/v2/provisionersdk"
	"github.com/wirtualdev/wirtualdev/v2/provisionersdk/proto"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbtestutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/rbac"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestTemplatePush(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)

		// Test the cli command.
		source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		inv, root := clitest.New(t, "templates", "push", template.Name, "--directory", source, "--test.provisioner", string(database.ProvisionerTypeEcho), "--name", "example")
		clitest.SetupConfig(t, templateAdmin, root)
		pty := ptytest.New(t).Attach(inv)

		execDone := make(chan error)
		go func() {
			execDone <- inv.Run()
		}()

		matches := []struct {
			match string
			write string
		}{
			{match: "Upload", write: "yes"},
		}
		for _, m := range matches {
			pty.ExpectMatch(m.match)
			pty.WriteLine(m.write)
		}

		require.NoError(t, <-execDone)

		// Assert that the template version changed.
		templateVersions, err := client.TemplateVersionsByTemplate(context.Background(), wirtualsdk.TemplateVersionsByTemplateRequest{
			TemplateID: template.ID,
		})
		require.NoError(t, err)
		assert.Len(t, templateVersions, 2)
		assert.NotEqual(t, template.ActiveVersionID, templateVersions[1].ID)
		require.Equal(t, "example", templateVersions[1].Name)
	})

	t.Run("Message less than or equal to 72 chars", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})

		wantMessage := strings.Repeat("a", 72)

		inv, root := clitest.New(t, "templates", "push", template.Name, "--directory", source, "--test.provisioner", string(database.ProvisionerTypeEcho), "--name", "example", "--message", wantMessage, "--yes")
		clitest.SetupConfig(t, templateAdmin, root)
		pty := ptytest.New(t).Attach(inv)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitMedium)
		defer cancel()

		inv = inv.WithContext(ctx)
		w := clitest.StartWithWaiter(t, inv)

		pty.ExpectNoMatchBefore(ctx, "Template message is longer than 72 characters", "Updated version at")

		w.RequireSuccess()

		// Assert that the template version changed.
		templateVersions, err := client.TemplateVersionsByTemplate(ctx, wirtualsdk.TemplateVersionsByTemplateRequest{
			TemplateID: template.ID,
		})
		require.NoError(t, err)
		assert.Len(t, templateVersions, 2)
		assert.NotEqual(t, template.ActiveVersionID, templateVersions[1].ID)
		require.Equal(t, wantMessage, templateVersions[1].Message)
	})

	t.Run("Message too long, warn but continue", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)
		source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		for i, tt := range []struct {
			wantMessage string
			wantMatch   string
		}{
			{wantMessage: strings.Repeat("a", 73), wantMatch: "Template message is longer than 72 characters"},
			{wantMessage: "This is my title\n\nAnd this is my body.", wantMatch: "Template message contains newlines"},
		} {
			inv, root := clitest.New(t, "templates", "push", template.Name,
				"--directory", source,
				"--test.provisioner", string(database.ProvisionerTypeEcho),
				"--message", tt.wantMessage,
				"--yes",
			)
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t).Attach(inv)

			inv = inv.WithContext(ctx)
			w := clitest.StartWithWaiter(t, inv)

			pty.ExpectMatchContext(ctx, tt.wantMatch)

			w.RequireSuccess()

			// Assert that the template version changed.
			templateVersions, err := client.TemplateVersionsByTemplate(ctx, wirtualsdk.TemplateVersionsByTemplateRequest{
				TemplateID: template.ID,
			})
			require.NoError(t, err)
			assert.Len(t, templateVersions, 2+i)
			assert.NotEqual(t, template.ActiveVersionID, templateVersions[1+i].ID)
			require.Equal(t, tt.wantMessage, templateVersions[1+i].Message)
		}
	})

	t.Run("NoLockfile", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)

		// Test the cli command.
		source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		require.NoError(t, os.Remove(filepath.Join(source, ".terraform.lock.hcl")))

		inv, root := clitest.New(t, "templates", "push", template.Name,
			"--directory", source,
			"--test.provisioner", string(database.ProvisionerTypeEcho),
			"--name", "example",
		)
		clitest.SetupConfig(t, templateAdmin, root)
		pty := ptytest.New(t).Attach(inv)

		execDone := make(chan error)
		go func() {
			execDone <- inv.Run()
		}()

		matches := []struct {
			match string
			write string
		}{
			{match: "No .terraform.lock.hcl file found"},
			{match: "Upload", write: "no"},
		}
		for _, m := range matches {
			pty.ExpectMatch(m.match)
			if m.write != "" {
				pty.WriteLine(m.write)
			}
		}

		// cmd should error once we say no.
		require.Error(t, <-execDone)
	})

	t.Run("NoLockfileIgnored", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)

		// Test the cli command.
		source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		require.NoError(t, os.Remove(filepath.Join(source, ".terraform.lock.hcl")))

		inv, root := clitest.New(t, "templates", "push", template.Name,
			"--directory", source,
			"--test.provisioner", string(database.ProvisionerTypeEcho),
			"--name", "example",
			"--ignore-lockfile",
		)
		clitest.SetupConfig(t, templateAdmin, root)
		pty := ptytest.New(t).Attach(inv)

		execDone := make(chan error)
		go func() {
			execDone <- inv.Run()
		}()

		{
			ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitMedium)
			defer cancel()

			pty.ExpectNoMatchBefore(ctx, "No .terraform.lock.hcl file found", "Upload")
			pty.WriteLine("no")
		}

		// cmd should error once we say no.
		require.Error(t, <-execDone)
	})

	t.Run("PushInactiveTemplateVersion", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)

		// Test the cli command.
		source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		inv, root := clitest.New(t, "templates", "push", template.Name,
			"--activate=false",
			"--directory", source,
			"--test.provisioner", string(database.ProvisionerTypeEcho),
			"--name", "example",
		)
		clitest.SetupConfig(t, templateAdmin, root)
		pty := ptytest.New(t).Attach(inv)
		w := clitest.StartWithWaiter(t, inv)

		matches := []struct {
			match string
			write string
		}{
			{match: "Upload", write: "yes"},
		}
		for _, m := range matches {
			pty.ExpectMatch(m.match)
			pty.WriteLine(m.write)
		}

		w.RequireSuccess()

		// Assert that the template version didn't change.
		templateVersions, err := client.TemplateVersionsByTemplate(context.Background(), wirtualsdk.TemplateVersionsByTemplateRequest{
			TemplateID: template.ID,
		})
		require.NoError(t, err)
		assert.Len(t, templateVersions, 2)
		assert.Equal(t, template.ActiveVersionID, templateVersions[0].ID)
		require.NotEqual(t, "example", templateVersions[0].Name)
	})

	t.Run("UseWorkingDir", func(t *testing.T) {
		t.Parallel()

		if runtime.GOOS == "windows" {
			t.Skip(`On Windows this test flakes with: "The process cannot access the file because it is being used by another process"`)
		}

		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		// Test the cli command.
		source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID,
			func(r *wirtualsdk.CreateTemplateRequest) {
				r.Name = filepath.Base(source)
			})

		// Don't pass the name of the template, it should use the
		// directory of the source.
		inv, root := clitest.New(t, "templates", "push",
			"--test.provisioner", string(database.ProvisionerTypeEcho),
			"--test.workdir", source,
		)
		clitest.SetupConfig(t, templateAdmin, root)
		pty := ptytest.New(t).Attach(inv)

		waiter := clitest.StartWithWaiter(t, inv)

		matches := []struct {
			match string
			write string
		}{
			{match: "Upload", write: "yes"},
		}
		for _, m := range matches {
			pty.ExpectMatch(m.match)
			pty.WriteLine(m.write)
		}

		waiter.RequireSuccess()

		// Assert that the template version changed.
		templateVersions, err := client.TemplateVersionsByTemplate(context.Background(), wirtualsdk.TemplateVersionsByTemplateRequest{
			TemplateID: template.ID,
		})
		require.NoError(t, err)
		assert.Len(t, templateVersions, 2)
		assert.NotEqual(t, template.ActiveVersionID, templateVersions[1].ID)
	})

	t.Run("Stdin", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
		owner := wirtualdtest.CreateFirstUser(t, client)
		templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
		version := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil)
		_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, version.ID)

		source, err := echo.Tar(&echo.Responses{
			Parse:          echo.ParseComplete,
			ProvisionApply: echo.ApplyComplete,
		})
		require.NoError(t, err)

		template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, version.ID)

		inv, root := clitest.New(
			t, "templates", "push", "--directory", "-",
			"--test.provisioner", string(database.ProvisionerTypeEcho),
			template.Name,
		)
		clitest.SetupConfig(t, templateAdmin, root)
		pty := ptytest.New(t)
		inv.Stdin = bytes.NewReader(source)
		inv.Stdout = pty.Output()

		execDone := make(chan error)
		go func() {
			execDone <- inv.Run()
		}()

		require.NoError(t, <-execDone)

		// Assert that the template version changed.
		templateVersions, err := client.TemplateVersionsByTemplate(context.Background(), wirtualsdk.TemplateVersionsByTemplateRequest{
			TemplateID: template.ID,
		})
		require.NoError(t, err)
		assert.Len(t, templateVersions, 2)
		assert.NotEqual(t, template.ActiveVersionID, templateVersions[1].ID)
	})

	t.Run("ProvisionerTags", func(t *testing.T) {
		t.Parallel()

		t.Run("WorkspaceTagsTerraform", func(t *testing.T) {
			t.Parallel()
			ctx := testutil.Context(t, testutil.WaitShort)

			// Start an instance **without** a built-in provisioner.
			// We're not actually testing that the Terraform applies.
			// What we test is that a provisioner job is created with the expected
			// tags based on the __content__ of the Terraform.
			store, ps := dbtestutil.NewDB(t)
			client := wirtualdtest.New(t, &wirtualdtest.Options{
				Database: store,
				Pubsub:   ps,
			})

			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

			// Create a tar file with some pre-defined content
			tarFile := testutil.CreateTar(t, map[string]string{
				"main.tf": `
variable "a" {
	type = string
	default = "1"
}
data "wirtual_parameter" "b" {
	type = string
	default = "2"
}
resource "null_resource" "test" {}
data "wirtual_workspace_tags" "tags" {
	tags = {
		"foo": "bar",
		"a": var.a,
		"b": data.wirtual_parameter.b.value,
	}
}`,
			})

			// Write the tar file to disk.
			tempDir := t.TempDir()
			err := tfparse.WriteArchive(tarFile, "application/x-tar", tempDir)
			require.NoError(t, err)

			// Run `wirtual templates push`
			templateName := strings.ReplaceAll(testutil.GetRandomName(t), "_", "-")
			var stdout, stderr strings.Builder
			inv, root := clitest.New(t, "templates", "push", templateName, "-d", tempDir, "--yes")
			inv.Stdout = &stdout
			inv.Stderr = &stderr
			clitest.SetupConfig(t, templateAdmin, root)

			// Don't forget to clean up!
			cancelCtx, cancel := context.WithCancel(ctx)
			t.Cleanup(cancel)
			done := make(chan error)
			go func() {
				done <- inv.WithContext(cancelCtx).Run()
			}()

			// Assert that a provisioner job was created with the desired tags.
			wantTags := database.StringMap(provisionersdk.MutateTags(uuid.Nil, map[string]string{
				"foo": "bar",
				"a":   "1",
				"b":   "2",
			}))
			require.Eventually(t, func() bool {
				jobs, err := store.GetProvisionerJobsCreatedAfter(ctx, time.Time{})
				if !assert.NoError(t, err) {
					return false
				}
				if len(jobs) == 0 {
					return false
				}
				return assert.EqualValues(t, wantTags, jobs[0].Tags)
			}, testutil.WaitShort, testutil.IntervalSlow)

			cancel()
			<-done

			require.Contains(t, stderr.String(), "No provisioners are available to handle the job!")
		})

		t.Run("ChangeTags", func(t *testing.T) {
			t.Parallel()

			// Start the first provisioner
			client, provisionerDocker, api := wirtualdtest.NewWithAPI(t, &wirtualdtest.Options{
				IncludeProvisionerDaemon: true,
				ProvisionerDaemonTags: map[string]string{
					"docker": "true",
				},
			})
			defer provisionerDocker.Close()

			// Start the second provisioner
			provisionerFoobar := wirtualdtest.NewTaggedProvisionerDaemon(t, api, "provisioner-foobar", map[string]string{
				"foobar": "foobaz",
			})
			defer provisionerFoobar.Close()

			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

			// Create the template with initial tagged template version.
			templateVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil, func(ctvr *wirtualsdk.CreateTemplateVersionRequest) {
				ctvr.ProvisionerTags = map[string]string{
					"docker": "true",
				}
			})
			templateVersion = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, templateVersion.ID)
			template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, templateVersion.ID)

			// Push new template version without provisioner tags. CLI should reuse tags from the previous version.
			source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
				Parse:          echo.ParseComplete,
				ProvisionApply: echo.ApplyComplete,
			})
			inv, root := clitest.New(t, "templates", "push", template.Name, "--directory", source, "--test.provisioner", string(database.ProvisionerTypeEcho), "--name", template.Name,
				"--provisioner-tag", "foobar=foobaz")
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t).Attach(inv)

			execDone := make(chan error)
			go func() {
				execDone <- inv.Run()
			}()

			matches := []struct {
				match string
				write string
			}{
				{match: "Upload", write: "yes"},
			}
			for _, m := range matches {
				pty.ExpectMatch(m.match)
				pty.WriteLine(m.write)
			}

			require.NoError(t, <-execDone)

			// Verify template version tags
			template, err := client.Template(context.Background(), template.ID)
			require.NoError(t, err)

			templateVersion, err = client.TemplateVersion(context.Background(), template.ActiveVersionID)
			require.NoError(t, err)
			require.EqualValues(t, map[string]string{"foobar": "foobaz", "owner": "", "scope": "organization"}, templateVersion.Job.Tags)
		})

		t.Run("DoNotChangeTags", func(t *testing.T) {
			t.Parallel()

			// Start the tagged provisioner
			client := wirtualdtest.New(t, &wirtualdtest.Options{
				IncludeProvisionerDaemon: true,
				ProvisionerDaemonTags: map[string]string{
					"docker": "true",
				},
			})
			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

			// Create the template with initial tagged template version.
			templateVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, nil, func(ctvr *wirtualsdk.CreateTemplateVersionRequest) {
				ctvr.ProvisionerTags = map[string]string{
					"docker": "true",
				}
			})
			templateVersion = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, templateVersion.ID)
			template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, templateVersion.ID)

			// Push new template version without provisioner tags. CLI should reuse tags from the previous version.
			source := clitest.CreateTemplateVersionSource(t, &echo.Responses{
				Parse:          echo.ParseComplete,
				ProvisionApply: echo.ApplyComplete,
			})
			inv, root := clitest.New(t, "templates", "push", template.Name, "--directory", source, "--test.provisioner", string(database.ProvisionerTypeEcho), "--name", template.Name)
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t).Attach(inv)

			execDone := make(chan error)
			go func() {
				execDone <- inv.Run()
			}()

			matches := []struct {
				match string
				write string
			}{
				{match: "Upload", write: "yes"},
			}
			for _, m := range matches {
				pty.ExpectMatch(m.match)
				pty.WriteLine(m.write)
			}

			require.NoError(t, <-execDone)

			// Verify template version tags
			template, err := client.Template(context.Background(), template.ID)
			require.NoError(t, err)

			templateVersion, err = client.TemplateVersion(context.Background(), template.ActiveVersionID)
			require.NoError(t, err)
			require.EqualValues(t, map[string]string{"docker": "true", "owner": "", "scope": "organization"}, templateVersion.Job.Tags)
		})
	})

	t.Run("Variables", func(t *testing.T) {
		t.Parallel()

		initialTemplateVariables := []*proto.TemplateVariable{
			{
				Name:         "first_variable",
				Description:  "This is the first variable",
				Type:         "string",
				DefaultValue: "abc",
				Required:     false,
				Sensitive:    true,
			},
		}

		t.Run("VariableIsRequired", func(t *testing.T) {
			t.Parallel()
			client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

			templateVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, createEchoResponsesWithTemplateVariables(initialTemplateVariables))
			_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, templateVersion.ID)
			template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, templateVersion.ID)

			// Test the cli command.
			modifiedTemplateVariables := append(initialTemplateVariables,
				&proto.TemplateVariable{
					Name:        "second_variable",
					Description: "This is the second variable.",
					Type:        "string",
					Required:    true,
				},
			)
			source := clitest.CreateTemplateVersionSource(t, createEchoResponsesWithTemplateVariables(modifiedTemplateVariables))
			tempDir := t.TempDir()
			removeTmpDirUntilSuccessAfterTest(t, tempDir)
			variablesFile, _ := os.CreateTemp(tempDir, "variables*.yaml")
			_, _ = variablesFile.WriteString(`second_variable: foobar`)
			inv, root := clitest.New(t, "templates", "push", template.Name,
				"--directory", source,
				"--test.provisioner", string(database.ProvisionerTypeEcho),
				"--name", "example",
				"--variables-file", variablesFile.Name(),
			)
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t)
			inv.Stdin = pty.Input()
			inv.Stdout = pty.Output()

			execDone := make(chan error)
			go func() {
				execDone <- inv.Run()
			}()

			matches := []struct {
				match string
				write string
			}{
				{match: "Upload", write: "yes"},
			}
			for _, m := range matches {
				pty.ExpectMatch(m.match)
				pty.WriteLine(m.write)
			}

			require.NoError(t, <-execDone)

			// Assert that the template version changed.
			templateVersions, err := client.TemplateVersionsByTemplate(context.Background(), wirtualsdk.TemplateVersionsByTemplateRequest{
				TemplateID: template.ID,
			})
			require.NoError(t, err)
			assert.Len(t, templateVersions, 2)
			assert.NotEqual(t, template.ActiveVersionID, templateVersions[1].ID)
			require.Equal(t, "example", templateVersions[1].Name)

			templateVariables, err := client.TemplateVersionVariables(context.Background(), templateVersions[1].ID)
			require.NoError(t, err)
			assert.Len(t, templateVariables, 2)
			require.Equal(t, "second_variable", templateVariables[1].Name)
			require.Equal(t, "foobar", templateVariables[1].Value)
		})

		t.Run("VariableIsRequiredButNotProvided", func(t *testing.T) {
			t.Parallel()
			client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

			templateVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, createEchoResponsesWithTemplateVariables(initialTemplateVariables))
			_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, templateVersion.ID)
			template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, templateVersion.ID)

			// Test the cli command.
			modifiedTemplateVariables := append(initialTemplateVariables,
				&proto.TemplateVariable{
					Name:        "second_variable",
					Description: "This is the second variable.",
					Type:        "string",
					Required:    true,
				},
			)
			source := clitest.CreateTemplateVersionSource(t, createEchoResponsesWithTemplateVariables(modifiedTemplateVariables))
			inv, root := clitest.New(t, "templates", "push", template.Name, "--directory", source, "--test.provisioner", string(database.ProvisionerTypeEcho), "--name", "example")
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t)
			inv.Stdin = pty.Input()
			inv.Stdout = pty.Output()

			execDone := make(chan error)
			go func() {
				execDone <- inv.Run()
			}()

			matches := []struct {
				match string
				write string
			}{
				{match: "Upload", write: "yes"},
			}
			for _, m := range matches {
				pty.ExpectMatch(m.match)
				pty.WriteLine(m.write)
			}

			wantErr := <-execDone
			require.Error(t, wantErr)
			require.Contains(t, wantErr.Error(), "required template variables need values")
		})

		t.Run("VariableIsOptionalButNotProvided", func(t *testing.T) {
			t.Parallel()
			client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

			templateVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, createEchoResponsesWithTemplateVariables(initialTemplateVariables))
			_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, templateVersion.ID)
			template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, templateVersion.ID)

			// Test the cli command.
			modifiedTemplateVariables := append(initialTemplateVariables,
				&proto.TemplateVariable{
					Name:         "second_variable",
					Description:  "This is the second variable",
					Type:         "string",
					DefaultValue: "abc",
					Required:     true,
				},
			)
			source := clitest.CreateTemplateVersionSource(t, createEchoResponsesWithTemplateVariables(modifiedTemplateVariables))
			inv, root := clitest.New(t, "templates", "push", template.Name,
				"--directory", source,
				"--test.provisioner", string(database.ProvisionerTypeEcho),
				"--name", "example",
			)
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t)
			inv.Stdin = pty.Input()
			inv.Stdout = pty.Output()

			execDone := make(chan error)
			go func() {
				execDone <- inv.Run()
			}()

			matches := []struct {
				match string
				write string
			}{
				{match: "Upload", write: "yes"},
			}
			for _, m := range matches {
				pty.ExpectMatch(m.match)
				pty.WriteLine(m.write)
			}

			require.NoError(t, <-execDone)

			// Assert that the template version changed.
			templateVersions, err := client.TemplateVersionsByTemplate(context.Background(), wirtualsdk.TemplateVersionsByTemplateRequest{
				TemplateID: template.ID,
			})
			require.NoError(t, err)
			assert.Len(t, templateVersions, 2)
			assert.NotEqual(t, template.ActiveVersionID, templateVersions[1].ID)
			require.Equal(t, "example", templateVersions[1].Name)

			templateVariables, err := client.TemplateVersionVariables(context.Background(), templateVersions[1].ID)
			require.NoError(t, err)
			assert.Len(t, templateVariables, 2)
			require.Equal(t, "second_variable", templateVariables[1].Name)
			require.Equal(t, "abc", templateVariables[1].Value)
			require.Equal(t, templateVariables[1].DefaultValue, templateVariables[1].Value)
		})

		t.Run("WithVariableOption", func(t *testing.T) {
			t.Parallel()
			client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())

			templateVersion := wirtualdtest.CreateTemplateVersion(t, client, owner.OrganizationID, createEchoResponsesWithTemplateVariables(initialTemplateVariables))
			_ = wirtualdtest.AwaitTemplateVersionJobCompleted(t, client, templateVersion.ID)
			template := wirtualdtest.CreateTemplate(t, client, owner.OrganizationID, templateVersion.ID)

			// Test the cli command.
			modifiedTemplateVariables := append(initialTemplateVariables,
				&proto.TemplateVariable{
					Name:        "second_variable",
					Description: "This is the second variable.",
					Type:        "string",
					Required:    true,
				},
			)
			source := clitest.CreateTemplateVersionSource(t, createEchoResponsesWithTemplateVariables(modifiedTemplateVariables))
			inv, root := clitest.New(t,
				"templates", "push", template.Name,
				"--directory", source,
				"--test.provisioner", string(database.ProvisionerTypeEcho),
				"--name", "example",
				"--variable", "second_variable=foobar",
			)
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t)
			inv.Stdin = pty.Input()
			inv.Stdout = pty.Output()

			execDone := make(chan error)
			go func() {
				execDone <- inv.Run()
			}()

			matches := []struct {
				match string
				write string
			}{
				{match: "Upload", write: "yes"},
			}
			for _, m := range matches {
				pty.ExpectMatch(m.match)
				pty.WriteLine(m.write)
			}

			require.NoError(t, <-execDone)

			// Assert that the template version changed.
			templateVersions, err := client.TemplateVersionsByTemplate(context.Background(), wirtualsdk.TemplateVersionsByTemplateRequest{
				TemplateID: template.ID,
			})
			require.NoError(t, err)
			assert.Len(t, templateVersions, 2)
			assert.NotEqual(t, template.ActiveVersionID, templateVersions[1].ID)
			require.Equal(t, "example", templateVersions[1].Name)

			templateVariables, err := client.TemplateVersionVariables(context.Background(), templateVersions[1].ID)
			require.NoError(t, err)
			assert.Len(t, templateVariables, 2)
			require.Equal(t, "second_variable", templateVariables[1].Name)
			require.Equal(t, "foobar", templateVariables[1].Value)
		})

		t.Run("CreateTemplate", func(t *testing.T) {
			t.Parallel()
			client := wirtualdtest.New(t, &wirtualdtest.Options{IncludeProvisionerDaemon: true})
			owner := wirtualdtest.CreateFirstUser(t, client)
			templateAdmin, _ := wirtualdtest.CreateAnotherUser(t, client, owner.OrganizationID, rbac.RoleTemplateAdmin())
			source := clitest.CreateTemplateVersionSource(t, completeWithAgent())

			const templateName = "my-template"
			args := []string{
				"templates",
				"push",
				templateName,
				"--directory", source,
				"--test.provisioner", string(database.ProvisionerTypeEcho),
			}
			inv, root := clitest.New(t, args...)
			clitest.SetupConfig(t, templateAdmin, root)
			pty := ptytest.New(t).Attach(inv)

			waiter := clitest.StartWithWaiter(t, inv)

			matches := []struct {
				match string
				write string
			}{
				{match: "Upload", write: "yes"},
				{match: "template has been created"},
			}
			for _, m := range matches {
				pty.ExpectMatch(m.match)
				if m.write != "" {
					pty.WriteLine(m.write)
				}
			}

			waiter.RequireSuccess()

			template, err := client.TemplateByName(context.Background(), owner.OrganizationID, templateName)
			require.NoError(t, err)
			require.Equal(t, templateName, template.Name)
			require.NotEqual(t, uuid.Nil, template.ActiveVersionID)
		})
	})
}

func createEchoResponsesWithTemplateVariables(templateVariables []*proto.TemplateVariable) *echo.Responses {
	return &echo.Responses{
		Parse: []*proto.Response{
			{
				Type: &proto.Response_Parse{
					Parse: &proto.ParseComplete{
						TemplateVariables: templateVariables,
					},
				},
			},
		},
		ProvisionPlan:  echo.PlanComplete,
		ProvisionApply: echo.ApplyComplete,
	}
}

func completeWithAgent() *echo.Responses {
	return &echo.Responses{
		Parse: echo.ParseComplete,
		ProvisionPlan: []*proto.Response{
			{
				Type: &proto.Response_Plan{
					Plan: &proto.PlanComplete{
						Resources: []*proto.Resource{
							{
								Type: "compute",
								Name: "main",
								Agents: []*proto.Agent{
									{
										Name:            "smith",
										OperatingSystem: "linux",
										Architecture:    "i386",
									},
								},
							},
						},
					},
				},
			},
		},
		ProvisionApply: []*proto.Response{
			{
				Type: &proto.Response_Apply{
					Apply: &proto.ApplyComplete{
						Resources: []*proto.Resource{
							{
								Type: "compute",
								Name: "main",
								Agents: []*proto.Agent{
									{
										Name:            "smith",
										OperatingSystem: "linux",
										Architecture:    "i386",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Need this for Windows because of a known issue with Go:
// https://github.com/golang/go/issues/52986
func removeTmpDirUntilSuccessAfterTest(t *testing.T, tempDir string) {
	t.Helper()
	t.Cleanup(func() {
		err := os.RemoveAll(tempDir)
		for err != nil {
			err = os.RemoveAll(tempDir)
		}
	})
}
