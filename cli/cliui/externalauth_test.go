package cliui_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestExternalAuth(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitShort)
	defer cancel()

	ptty := ptytest.New(t)
	cmd := &serpent.Command{
		Handler: func(inv *serpent.Invocation) error {
			var fetched atomic.Bool
			return cliui.ExternalAuth(inv.Context(), inv.Stdout, cliui.ExternalAuthOptions{
				Fetch: func(ctx context.Context) ([]wirtualsdk.TemplateVersionExternalAuth, error) {
					defer fetched.Store(true)
					return []wirtualsdk.TemplateVersionExternalAuth{{
						ID:              "github",
						DisplayName:     "GitHub",
						Type:            wirtualsdk.EnhancedExternalAuthProviderGitHub.String(),
						Authenticated:   fetched.Load(),
						AuthenticateURL: "https://example.com/gitauth/github",
					}}, nil
				},
				FetchInterval: time.Millisecond,
			})
		},
	}

	inv := cmd.Invoke().WithContext(ctx)

	ptty.Attach(inv)
	done := make(chan struct{})
	go func() {
		defer close(done)
		err := inv.Run()
		assert.NoError(t, err)
	}()
	ptty.ExpectMatchContext(ctx, "You must authenticate with")
	ptty.ExpectMatchContext(ctx, "https://example.com/gitauth/github")
	ptty.ExpectMatchContext(ctx, "Successfully authenticated with GitHub")
	<-done
}
