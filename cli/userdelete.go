package cli

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/wirtualdev/pretty"
	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func (r *RootCmd) userDelete() *serpent.Command {
	client := new(wirtualsdk.Client)
	cmd := &serpent.Command{
		Use:   "delete <username|user_id>",
		Short: "Delete a user by username or user_id.",
		Middleware: serpent.Chain(
			serpent.RequireNArgs(1),
			r.InitClient(client),
		),
		Handler: func(inv *serpent.Invocation) error {
			ctx := inv.Context()
			user, err := client.User(ctx, inv.Args[0])
			if err != nil {
				return xerrors.Errorf("fetch user: %w", err)
			}

			err = client.DeleteUser(ctx, user.ID)
			if err != nil {
				return xerrors.Errorf("delete user: %w", err)
			}

			_, _ = fmt.Fprintln(inv.Stderr,
				"Successfully deleted "+pretty.Sprint(cliui.DefaultStyles.Keyword, user.Username)+".",
			)
			return nil
		},
	}
	return cmd
}
