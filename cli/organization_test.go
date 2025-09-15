package cli_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestCurrentOrganization(t *testing.T) {
	t.Parallel()

	// This test emulates 2 cases:
	// 1. The user is not a part of the default organization, but only belongs to one.
	// 2. The user is connecting to an older Wirtual instance.
	t.Run("no-default", func(t *testing.T) {
		t.Parallel()

		orgID := uuid.New()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]wirtualsdk.Organization{
				{
					MinimalOrganization: wirtualsdk.MinimalOrganization{
						ID:   orgID,
						Name: "not-default",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					IsDefault: false,
				},
			})
		}))
		defer srv.Close()

		client := wirtualsdk.New(must(url.Parse(srv.URL)))
		inv, root := clitest.New(t, "organizations", "show", "selected")
		clitest.SetupConfig(t, client, root)
		pty := ptytest.New(t).Attach(inv)
		errC := make(chan error)
		go func() {
			errC <- inv.Run()
		}()
		require.NoError(t, <-errC)
		pty.ExpectMatch(orgID.String())
	})
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}
