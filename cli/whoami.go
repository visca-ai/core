package cli

import (
	"fmt"

	"github.com/wirtualdev/pretty"
	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func (r *RootCmd) whoami() *serpent.Command {
	client := new(wirtualsdk.Client)
	cmd := &serpent.Command{
		Annotations: workspaceCommand,
		Use:         "whoami",
		Short:       "Fetch authenticated user info for Wirtual deployment",
		Middleware: serpent.Chain(
			serpent.RequireNArgs(0),
			r.InitClient(client),
		),
		Handler: func(inv *serpent.Invocation) error {
			ctx := inv.Context()
			// Fetch the user info
			resp, err := client.User(ctx, wirtualsdk.Me)
			// Get Wirtual instance url
			clientURL := client.URL

			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(inv.Stdout, Caret+"Wirtual is running at %s, You're authenticated as %s !\n", pretty.Sprint(cliui.DefaultStyles.Keyword, clientURL), pretty.Sprint(cliui.DefaultStyles.Keyword, resp.Username))
			return err
		},
	}
	return cmd
}
