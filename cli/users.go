package cli

import (
	"github.com/wirtualdev/serpent"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func (r *RootCmd) users() *serpent.Command {
	cmd := &serpent.Command{
		Short:   "Manage users",
		Use:     "users [subcommand]",
		Aliases: []string{"user"},
		Handler: func(inv *serpent.Invocation) error {
			return inv.Command.HelpHandler(inv)
		},
		Children: []*serpent.Command{
			r.userCreate(),
			r.userList(),
			r.userSingle(),
			r.userDelete(),
			r.createUserStatusCommand(wirtualsdk.UserStatusActive),
			r.createUserStatusCommand(wirtualsdk.UserStatusSuspended),
		},
	}
	return cmd
}
