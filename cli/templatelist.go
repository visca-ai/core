package cli

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func (r *RootCmd) templateList() *serpent.Command {
	formatter := cliui.NewOutputFormatter(
		cliui.TableFormat([]templateTableRow{}, []string{"name", "organization name", "last updated", "used by"}),
		cliui.JSONFormat(),
	)

	client := new(wirtualsdk.Client)
	cmd := &serpent.Command{
		Use:     "list",
		Short:   "List all the templates available for the organization",
		Aliases: []string{"ls"},
		Middleware: serpent.Chain(
			r.InitClient(client),
		),
		Handler: func(inv *serpent.Invocation) error {
			templates, err := client.Templates(inv.Context(), wirtualsdk.TemplateFilter{})
			if err != nil {
				return err
			}

			if len(templates) == 0 {
				_, _ = fmt.Fprintf(inv.Stderr, "%s No templates found! Create one:\n\n", Caret)
				_, _ = fmt.Fprintln(inv.Stderr, color.HiMagentaString("  $ wirtual templates push <directory>\n"))
				return nil
			}

			rows := templatesToRows(templates...)
			out, err := formatter.Format(inv.Context(), rows)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(inv.Stdout, out)
			return err
		},
	}

	formatter.AttachOptions(&cmd.Options)
	return cmd
}
