package clitest

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/config"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbtestutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

// UpdateGoldenFiles indicates golden files should be updated.
// To update the golden files:
// make update-golden-files
var UpdateGoldenFiles = flag.Bool("update", false, "update .golden files")

var timestampRegex = regexp.MustCompile(`(?i)\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(.\d+)?(Z|[+-]\d+:\d+)`)

type CommandHelpCase struct {
	Name string
	Cmd  []string
}

func DefaultCases() []CommandHelpCase {
	return []CommandHelpCase{
		{
			Name: "wirtual --help",
			Cmd:  []string{"--help"},
		},
		{
			Name: "wirtual server --help",
			Cmd:  []string{"server", "--help"},
		},
	}
}

// TestCommandHelp will test the help output of the given commands
// using golden files.
func TestCommandHelp(t *testing.T, getRoot func(t *testing.T) *serpent.Command, cases []CommandHelpCase) {
	t.Parallel()
	rootClient, replacements := prepareTestData(t)

	root := getRoot(t)

ExtractCommandPathsLoop:
	for _, cp := range extractVisibleCommandPaths(nil, root.Children) {
		name := fmt.Sprintf("wirtual %s --help", strings.Join(cp, " "))
		cmd := append(cp, "--help")
		for _, tt := range cases {
			if tt.Name == name {
				continue ExtractCommandPathsLoop
			}
		}
		cases = append(cases, CommandHelpCase{Name: name, Cmd: cmd})
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			ctx := testutil.Context(t, testutil.WaitLong)

			var outBuf bytes.Buffer

			caseCmd := getRoot(t)

			inv, cfg := NewWithCommand(t, caseCmd, tt.Cmd...)
			inv.Stderr = &outBuf
			inv.Stdout = &outBuf
			inv.Environ.Set("WIRTUAL_URL", rootClient.URL.String())
			inv.Environ.Set("WIRTUAL_SESSION_TOKEN", rootClient.SessionToken())
			inv.Environ.Set("WIRTUAL_CACHE_DIRECTORY", "~/.cache")

			SetupConfig(t, rootClient, cfg)

			StartWithWaiter(t, inv.WithContext(ctx)).RequireSuccess()

			TestGoldenFile(t, tt.Name, outBuf.Bytes(), replacements)
		})
	}
}

// TestGoldenFile will test the given bytes slice input against the
// golden file with the given file name, optionally using the given replacements.
func TestGoldenFile(t *testing.T, fileName string, actual []byte, replacements map[string]string) {
	if len(actual) == 0 {
		t.Fatal("no output")
	}

	for k, v := range replacements {
		actual = bytes.ReplaceAll(actual, []byte(k), []byte(v))
	}

	actual = normalizeGoldenFile(t, actual)
	goldenPath := filepath.Join("testdata", strings.ReplaceAll(fileName, " ", "_")+".golden")
	if *UpdateGoldenFiles {
		t.Logf("update golden file for: %q: %s", fileName, goldenPath)
		err := os.WriteFile(goldenPath, actual, 0o600)
		require.NoError(t, err, "update golden file")
	}

	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "read golden file, run \"make update-golden-files\" and commit the changes")

	expected = normalizeGoldenFile(t, expected)
	require.Equal(
		t, string(expected), string(actual),
		"golden file mismatch: %s, run \"make update-golden-files\", verify and commit the changes",
		goldenPath,
	)
}

// normalizeGoldenFile replaces any strings that are system or timing dependent
// with a placeholder so that the golden files can be compared with a simple
// equality check.
func normalizeGoldenFile(t *testing.T, byt []byte) []byte {
	// Replace any timestamps with a placeholder.
	byt = timestampRegex.ReplaceAll(byt, []byte("[timestamp]"))

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	configDir := config.DefaultDir()
	byt = bytes.ReplaceAll(byt, []byte(configDir), []byte("~/.config/wirtualv2"))

	byt = bytes.ReplaceAll(byt, []byte(wirtualsdk.DefaultCacheDir()), []byte("[cache dir]"))

	// The home directory changes depending on the test environment.
	byt = bytes.ReplaceAll(byt, []byte(homeDir), []byte("~"))
	for _, r := range []struct {
		old string
		new string
	}{
		{"\r\n", "\n"},
		{`~\.cache\wirtual`, "~/.cache/wirtual"},
		{`C:\Users\RUNNER~1\AppData\Local\Temp`, "/tmp"},
		{os.TempDir(), "/tmp"},
	} {
		byt = bytes.ReplaceAll(byt, []byte(r.old), []byte(r.new))
	}
	return byt
}

func extractVisibleCommandPaths(cmdPath []string, cmds []*serpent.Command) [][]string {
	var cmdPaths [][]string
	for _, c := range cmds {
		if c.Hidden {
			continue
		}
		cmdPath := append(cmdPath, c.Name())
		cmdPaths = append(cmdPaths, cmdPath)
		cmdPaths = append(cmdPaths, extractVisibleCommandPaths(cmdPath, c.Children)...)
	}
	return cmdPaths
}

func prepareTestData(t *testing.T) (*wirtualsdk.Client, map[string]string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()

	// This needs to be a fixed timezone because timezones increase the length
	// of timestamp strings. The increased length can pad table formatting's
	// and differ the table header spacings.
	//nolint:gocritic
	db, pubsub := dbtestutil.NewDB(t, dbtestutil.WithTimezone("UTC"))
	rootClient := wirtualdtest.New(t, &wirtualdtest.Options{
		Database:                 db,
		Pubsub:                   pubsub,
		IncludeProvisionerDaemon: true,
	})
	firstUser := wirtualdtest.CreateFirstUser(t, rootClient)
	secondUser, err := rootClient.CreateUserWithOrgs(ctx, wirtualsdk.CreateUserRequestWithOrgs{
		Email:           "testuser2@wirtual.dev",
		Username:        "testuser2",
		Password:        wirtualdtest.FirstUserParams.Password,
		OrganizationIDs: []uuid.UUID{firstUser.OrganizationID},
	})
	require.NoError(t, err)
	version := wirtualdtest.CreateTemplateVersion(t, rootClient, firstUser.OrganizationID, nil)
	version = wirtualdtest.AwaitTemplateVersionJobCompleted(t, rootClient, version.ID)
	template := wirtualdtest.CreateTemplate(t, rootClient, firstUser.OrganizationID, version.ID, func(req *wirtualsdk.CreateTemplateRequest) {
		req.Name = "test-template"
	})
	workspace := wirtualdtest.CreateWorkspace(t, rootClient, template.ID, func(req *wirtualsdk.CreateWorkspaceRequest) {
		req.Name = "test-workspace"
	})
	workspaceBuild := wirtualdtest.AwaitWorkspaceBuildJobCompleted(t, rootClient, workspace.LatestBuild.ID)

	replacements := map[string]string{
		firstUser.UserID.String():            "[first user ID]",
		secondUser.ID.String():               "[second user ID]",
		firstUser.OrganizationID.String():    "[first org ID]",
		version.ID.String():                  "[version ID]",
		version.Name:                         "[version name]",
		version.Job.ID.String():              "[version job ID]",
		version.Job.FileID.String():          "[version file ID]",
		version.Job.WorkerID.String():        "[version worker ID]",
		template.ID.String():                 "[template ID]",
		workspace.ID.String():                "[workspace ID]",
		workspaceBuild.ID.String():           "[workspace build ID]",
		workspaceBuild.Job.ID.String():       "[workspace build job ID]",
		workspaceBuild.Job.FileID.String():   "[workspace build file ID]",
		workspaceBuild.Job.WorkerID.String(): "[workspace build worker ID]",
	}

	return rootClient, replacements
}
