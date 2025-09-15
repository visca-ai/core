package cli

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/wirtualdev/pretty"
	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func (r *RootCmd) createOrganization() *serpent.Command {
	client := new(wirtualsdk.Client)

	cmd := &serpent.Command{
		Use:   "create <organization name>",
		Short: "Create a new organization.",
		Middleware: serpent.Chain(
			r.InitClient(client),
			serpent.RequireNArgs(1),
		),
		Options: serpent.OptionSet{
			cliui.SkipPromptOption(),
		},
		Handler: func(inv *serpent.Invocation) error {
			orgName := inv.Args[0]

			err := wirtualsdk.NameValid(orgName)
			if err != nil {
				return xerrors.Errorf("organization name %q is invalid: %w", orgName, err)
			}

			// This check is not perfect since not all users can read all organizations.
			// So ignore the error and if the org already exists, prevent the user
			// from creating it.
			existing, _ := client.OrganizationByName(inv.Context(), orgName)
			if existing.ID != uuid.Nil {
				return xerrors.Errorf("organization %q already exists", orgName)
			}

			_, err = cliui.Prompt(inv, cliui.PromptOptions{
				Text: fmt.Sprintf("Are you sure you want to create an organization with the name %s?\n%s",
					pretty.Sprint(cliui.DefaultStyles.Code, orgName),
					pretty.Sprint(cliui.BoldFmt(), "This action is irreversible."),
				),
				IsConfirm: true,
				Default:   cliui.ConfirmNo,
			})
			if err != nil {
				return err
			}

			organization, err := client.CreateOrganization(inv.Context(), wirtualsdk.CreateOrganizationRequest{
				Name: orgName,
			})
			if err != nil {
				return xerrors.Errorf("failed to create organization: %w", err)
			}

			_, _ = fmt.Fprintf(inv.Stdout, "Organization %s (%s) created.\n", organization.Name, organization.ID)
			return nil
		},
	}

	return cmd
}
