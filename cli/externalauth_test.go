package cli_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/httpapi"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk/agentsdk"
)

func TestExternalAuth(t *testing.T) {
	t.Parallel()
	t.Run("CanceledWithURL", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpapi.Write(context.Background(), w, http.StatusOK, agentsdk.ExternalAuthResponse{
				URL: "https://github.com",
			})
		}))
		t.Cleanup(srv.Close)
		url := srv.URL
		inv, _ := clitest.New(t, "--agent-url", url, "--agent-token", "foo", "external-auth", "access-token", "github")
		pty := ptytest.New(t)
		inv.Stdout = pty.Output()
		waiter := clitest.StartWithWaiter(t, inv)
		pty.ExpectMatch("https://github.com")
		waiter.RequireIs(cliui.Canceled)
	})
	t.Run("SuccessWithToken", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpapi.Write(context.Background(), w, http.StatusOK, agentsdk.ExternalAuthResponse{
				AccessToken: "bananas",
			})
		}))
		t.Cleanup(srv.Close)
		url := srv.URL
		inv, _ := clitest.New(t, "--agent-url", url, "--agent-token", "foo", "external-auth", "access-token", "github")
		pty := ptytest.New(t)
		inv.Stdout = pty.Output()
		clitest.Start(t, inv)
		pty.ExpectMatch("bananas")
	})
	t.Run("NoArgs", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpapi.Write(context.Background(), w, http.StatusOK, agentsdk.ExternalAuthResponse{
				AccessToken: "bananas",
			})
		}))
		t.Cleanup(srv.Close)
		url := srv.URL
		inv, _ := clitest.New(t, "--agent-url", url, "--agent-token", "foo", "external-auth", "access-token")
		watier := clitest.StartWithWaiter(t, inv)
		watier.RequireContains("wanted 1 args but got 0")
	})
	t.Run("SuccessWithExtra", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpapi.Write(context.Background(), w, http.StatusOK, agentsdk.ExternalAuthResponse{
				AccessToken: "bananas",
				TokenExtra: map[string]interface{}{
					"hey": "there",
				},
			})
		}))
		t.Cleanup(srv.Close)
		url := srv.URL
		inv, _ := clitest.New(t, "--agent-url", url, "--agent-token", "foo", "external-auth", "access-token", "github", "--extra", "hey")
		pty := ptytest.New(t)
		inv.Stdout = pty.Output()
		clitest.Start(t, inv)
		pty.ExpectMatch("there")
	})
}
