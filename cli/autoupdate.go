package cli

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func (r *RootCmd) autoupdate() *serpent.Command {
	client := new(wirtualsdk.Client)
	cmd := &serpent.Command{
		Annotations: workspaceCommand,
		Use:         "autoupdate <workspace> <always|never>",
		Short:       "Toggle auto-update policy for a workspace",
		Middleware: serpent.Chain(
			serpent.RequireNArgs(2),
			r.InitClient(client),
		),
		Handler: func(inv *serpent.Invocation) error {
			policy := strings.ToLower(inv.Args[1])
			err := validateAutoUpdatePolicy(policy)
			if err != nil {
				return xerrors.Errorf("validate policy: %w", err)
			}

			workspace, err := namedWorkspace(inv.Context(), client, inv.Args[0])
			if err != nil {
				return xerrors.Errorf("get workspace: %w", err)
			}

			err = client.UpdateWorkspaceAutomaticUpdates(inv.Context(), workspace.ID, wirtualsdk.UpdateWorkspaceAutomaticUpdatesRequest{
				AutomaticUpdates: wirtualsdk.AutomaticUpdates(policy),
			})
			if err != nil {
				return xerrors.Errorf("update workspace automatic updates policy: %w", err)
			}
			_, _ = fmt.Fprintf(inv.Stdout, "Updated workspace %q auto-update policy to %q\n", workspace.Name, policy)
			return nil
		},
	}

	cmd.Options = append(cmd.Options, cliui.SkipPromptOption())
	return cmd
}

func validateAutoUpdatePolicy(arg string) error {
	switch wirtualsdk.AutomaticUpdates(arg) {
	case wirtualsdk.AutomaticUpdatesAlways, wirtualsdk.AutomaticUpdatesNever:
		return nil
	default:
		return xerrors.Errorf("invalid option %q must be either of %q or %q", arg, wirtualsdk.AutomaticUpdatesAlways, wirtualsdk.AutomaticUpdatesNever)
	}
}
