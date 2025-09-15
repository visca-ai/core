package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func createOpts(t *testing.T) *wirtualdtest.Options {
	t.Helper()

	dt := wirtualdtest.DeploymentValues(t)
	return &wirtualdtest.Options{
		DeploymentValues: dt,
	}
}

func TestNotifications(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		command      string
		expectPaused bool
	}{
		{
			name:         "PauseNotifications",
			command:      "pause",
			expectPaused: true,
		},
		{
			name:         "ResumeNotifications",
			command:      "resume",
			expectPaused: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			ownerClient, db := wirtualdtest.NewWithDatabase(t, createOpts(t))
			_ = wirtualdtest.CreateFirstUser(t, ownerClient)

			// when
			inv, root := clitest.New(t, "notifications", tt.command)
			clitest.SetupConfig(t, ownerClient, root)

			var buf bytes.Buffer
			inv.Stdout = &buf
			err := inv.Run()
			require.NoError(t, err)

			// then
			ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
			t.Cleanup(cancel)
			settingsJSON, err := db.GetNotificationsSettings(ctx)
			require.NoError(t, err)

			var settings wirtualsdk.NotificationsSettings
			err = json.Unmarshal([]byte(settingsJSON), &settings)
			require.NoError(t, err)
			require.Equal(t, tt.expectPaused, settings.NotifierPaused)
		})
	}
}

func TestPauseNotifications_RegularUser(t *testing.T) {
	t.Parallel()

	// given
	ownerClient, db := wirtualdtest.NewWithDatabase(t, createOpts(t))
	owner := wirtualdtest.CreateFirstUser(t, ownerClient)
	anotherClient, _ := wirtualdtest.CreateAnotherUser(t, ownerClient, owner.OrganizationID)

	// when
	inv, root := clitest.New(t, "notifications", "pause")
	clitest.SetupConfig(t, anotherClient, root)

	var buf bytes.Buffer
	inv.Stdout = &buf
	err := inv.Run()
	var sdkError *wirtualsdk.Error
	require.Error(t, err)
	require.ErrorAsf(t, err, &sdkError, "error should be of type *wirtualsdk.Error")
	assert.Equal(t, http.StatusForbidden, sdkError.StatusCode())
	assert.Contains(t, sdkError.Message, "Forbidden.")

	// then
	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
	t.Cleanup(cancel)
	settingsJSON, err := db.GetNotificationsSettings(ctx)
	require.NoError(t, err)

	var settings wirtualsdk.NotificationsSettings
	err = json.Unmarshal([]byte(settingsJSON), &settings)
	require.NoError(t, err)
	require.False(t, settings.NotifierPaused) // still running
}
