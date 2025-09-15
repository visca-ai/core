package cli_test

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbgen"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/rbac"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
)

func TestShowOrganizationRoles(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		ownerClient, db := wirtualdtest.NewWithDatabase(t, &wirtualdtest.Options{})
		owner := wirtualdtest.CreateFirstUser(t, ownerClient)
		client, _ := wirtualdtest.CreateAnotherUser(t, ownerClient, owner.OrganizationID, rbac.RoleUserAdmin())

		const expectedRole = "test-role"
		dbgen.CustomRole(t, db, database.CustomRole{
			Name:            expectedRole,
			DisplayName:     "Expected",
			SitePermissions: nil,
			OrgPermissions:  nil,
			UserPermissions: nil,
			OrganizationID: uuid.NullUUID{
				UUID:  owner.OrganizationID,
				Valid: true,
			},
		})

		ctx := testutil.Context(t, testutil.WaitMedium)
		inv, root := clitest.New(t, "organization", "roles", "show")
		clitest.SetupConfig(t, client, root)

		buf := new(bytes.Buffer)
		inv.Stdout = buf
		err := inv.WithContext(ctx).Run()
		require.NoError(t, err)
		require.Contains(t, buf.String(), expectedRole)
	})
}
