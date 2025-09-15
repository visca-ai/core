package cli

import (
	"fmt"
	"time"

	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func (r *RootCmd) stop() *serpent.Command {
	var bflags buildFlags
	client := new(wirtualsdk.Client)
	cmd := &serpent.Command{
		Annotations: workspaceCommand,
		Use:         "stop <workspace>",
		Short:       "Stop a workspace",
		Middleware: serpent.Chain(
			serpent.RequireNArgs(1),
			r.InitClient(client),
		),
		Options: serpent.OptionSet{
			cliui.SkipPromptOption(),
		},
		Handler: func(inv *serpent.Invocation) error {
			_, err := cliui.Prompt(inv, cliui.PromptOptions{
				Text:      "Confirm stop workspace?",
				IsConfirm: true,
			})
			if err != nil {
				return err
			}

			workspace, err := namedWorkspace(inv.Context(), client, inv.Args[0])
			if err != nil {
				return err
			}
			wbr := wirtualsdk.CreateWorkspaceBuildRequest{
				Transition: wirtualsdk.WorkspaceTransitionStop,
			}
			if bflags.provisionerLogDebug {
				wbr.LogLevel = wirtualsdk.ProvisionerLogLevelDebug
			}
			build, err := client.CreateWorkspaceBuild(inv.Context(), workspace.ID, wbr)
			if err != nil {
				return err
			}

			err = cliui.WorkspaceBuild(inv.Context(), inv.Stdout, client, build.ID)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				inv.Stdout,
				"\nThe %s workspace has been stopped at %s!\n", cliui.Keyword(workspace.Name),

				cliui.Timestamp(time.Now()),
			)
			return nil
		},
	}
	cmd.Options = append(cmd.Options, bflags.cliOptions()...)

	return cmd
}
