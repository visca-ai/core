package cli_test

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbtestutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/database/dbtime"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/rbac"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/userpassword"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

//nolint:paralleltest, tparallel
func TestServerCreateAdminUser(t *testing.T) {
	const (
		username = "dean"
		email    = "dean@example.com"
		password = "SecurePa$$word123"
	)

	verifyUser := func(t *testing.T, dbURL, username, email, password string) {
		t.Helper()

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		sqlDB, err := sql.Open("postgres", dbURL)
		require.NoError(t, err)
		defer sqlDB.Close()
		db := database.New(sqlDB)

		pingCtx, pingCancel := context.WithTimeout(ctx, testutil.WaitShort)
		defer pingCancel()
		_, err = db.Ping(pingCtx)
		require.NoError(t, err, "ping db")

		user, err := db.GetUserByEmailOrUsername(ctx, database.GetUserByEmailOrUsernameParams{
			Email: email,
		})
		require.NoError(t, err)
		require.Equal(t, username, user.Username, "username does not match")
		require.Equal(t, email, user.Email, "email does not match")

		ok, err := userpassword.Compare(string(user.HashedPassword), password)
		require.NoError(t, err)
		require.True(t, ok, "password does not match")

		require.EqualValues(t, []string{wirtualsdk.RoleOwner}, user.RBACRoles, "user does not have owner role")

		// Check that user is admin in every org.
		orgs, err := db.GetOrganizations(ctx, database.GetOrganizationsParams{})
		require.NoError(t, err)
		orgIDs := make(map[uuid.UUID]struct{}, len(orgs))
		for _, org := range orgs {
			orgIDs[org.ID] = struct{}{}
		}

		orgMemberships, err := db.OrganizationMembers(ctx, database.OrganizationMembersParams{UserID: user.ID})
		require.NoError(t, err)
		orgIDs2 := make(map[uuid.UUID]struct{}, len(orgMemberships))
		for _, membership := range orgMemberships {
			orgIDs2[membership.OrganizationMember.OrganizationID] = struct{}{}
			assert.Equal(t, []string{rbac.RoleOrgAdmin()}, membership.OrganizationMember.Roles, "user is not org admin")
		}

		require.Equal(t, orgIDs, orgIDs2, "user is not in all orgs")
	}

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		if runtime.GOOS != "linux" || testing.Short() {
			// Skip on non-Linux because it spawns a PostgreSQL instance.
			t.SkipNow()
		}
		connectionURL, err := dbtestutil.Open(t)
		require.NoError(t, err)

		sqlDB, err := sql.Open("postgres", connectionURL)
		require.NoError(t, err)
		defer sqlDB.Close()
		db := database.New(sqlDB)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitMedium)
		defer cancel()

		pingCtx, pingCancel := context.WithTimeout(ctx, testutil.WaitShort)
		defer pingCancel()
		_, err = db.Ping(pingCtx)
		require.NoError(t, err, "ping db")

		// Insert a few orgs.
		org1Name, org1ID := "org1", uuid.New()
		org2Name, org2ID := "org2", uuid.New()
		_, err = db.InsertOrganization(ctx, database.InsertOrganizationParams{
			ID:        org1ID,
			Name:      org1Name,
			CreatedAt: dbtime.Now(),
			UpdatedAt: dbtime.Now(),
		})
		require.NoError(t, err)
		_, err = db.InsertOrganization(ctx, database.InsertOrganizationParams{
			ID:        org2ID,
			Name:      org2Name,
			CreatedAt: dbtime.Now(),
			UpdatedAt: dbtime.Now(),
		})
		require.NoError(t, err)

		inv, _ := clitest.New(t,
			"server", "create-admin-user",
			"--postgres-url", connectionURL,
			"--ssh-keygen-algorithm", "ed25519",
			"--username", username,
			"--email", email,
			"--password", password,
		)
		pty := ptytest.New(t)
		inv.Stdout = pty.Output()
		inv.Stderr = pty.Output()
		clitest.Start(t, inv)

		pty.ExpectMatchContext(ctx, "Creating user...")
		pty.ExpectMatchContext(ctx, "Generating user SSH key...")
		pty.ExpectMatchContext(ctx, fmt.Sprintf("Adding user to organization %q (%s) as admin...", org1Name, org1ID.String()))
		pty.ExpectMatchContext(ctx, fmt.Sprintf("Adding user to organization %q (%s) as admin...", org2Name, org2ID.String()))
		pty.ExpectMatchContext(ctx, "User created successfully.")
		pty.ExpectMatchContext(ctx, username)
		pty.ExpectMatchContext(ctx, email)
		pty.ExpectMatchContext(ctx, "****")

		verifyUser(t, connectionURL, username, email, password)
	})

	t.Run("Env", func(t *testing.T) {
		t.Parallel()
		if runtime.GOOS != "linux" || testing.Short() {
			// Skip on non-Linux because it spawns a PostgreSQL instance.
			t.SkipNow()
		}
		connectionURL, err := dbtestutil.Open(t)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitMedium)
		defer cancel()

		inv, _ := clitest.New(t, "server", "create-admin-user")
		inv.Environ.Set("WIRTUAL_PG_CONNECTION_URL", connectionURL)
		inv.Environ.Set("WIRTUAL_SSH_KEYGEN_ALGORITHM", "ed25519")
		inv.Environ.Set("WIRTUAL_USERNAME", username)
		inv.Environ.Set("WIRTUAL_EMAIL", email)
		inv.Environ.Set("WIRTUAL_PASSWORD", password)

		pty := ptytest.New(t)
		inv.Stdout = pty.Output()
		inv.Stderr = pty.Output()
		clitest.Start(t, inv)

		pty.ExpectMatchContext(ctx, "User created successfully.")
		pty.ExpectMatchContext(ctx, username)
		pty.ExpectMatchContext(ctx, email)
		pty.ExpectMatchContext(ctx, "****")

		verifyUser(t, connectionURL, username, email, password)
	})

	t.Run("Stdin", func(t *testing.T) {
		t.Parallel()

		if runtime.GOOS != "linux" || testing.Short() {
			// Skip on non-Linux because it spawns a PostgreSQL instance.
			t.SkipNow()
		}
		connectionURL, err := dbtestutil.Open(t)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitMedium)
		defer cancel()

		inv, _ := clitest.New(t,
			"server", "create-admin-user",
			"--postgres-url", connectionURL,
			"--ssh-keygen-algorithm", "ed25519",
		)
		pty := ptytest.New(t).Attach(inv)

		clitest.Start(t, inv)

		pty.ExpectMatchContext(ctx, "Username")
		pty.WriteLine(username)
		pty.ExpectMatchContext(ctx, "Email")
		pty.WriteLine(email)
		pty.ExpectMatchContext(ctx, "Password")
		pty.WriteLine(password)
		pty.ExpectMatchContext(ctx, "Confirm password")
		pty.WriteLine(password)

		pty.ExpectMatchContext(ctx, "User created successfully.")
		pty.ExpectMatchContext(ctx, username)
		pty.ExpectMatchContext(ctx, email)
		pty.ExpectMatchContext(ctx, "****")

		verifyUser(t, connectionURL, username, email, password)
	})

	t.Run("Validates", func(t *testing.T) {
		t.Parallel()

		if runtime.GOOS != "linux" || testing.Short() {
			// Skip on non-Linux because it spawns a PostgreSQL instance.
			t.SkipNow()
		}
		connectionURL, err := dbtestutil.Open(t)
		require.NoError(t, err)
		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		root, _ := clitest.New(t,
			"server", "create-admin-user",
			"--postgres-url", connectionURL,
			"--ssh-keygen-algorithm", "rsa4096",
			"--username", "$",
			"--email", "not-an-email",
			"--password", "x",
		)
		pty := ptytest.New(t)
		root.Stdout = pty.Output()
		root.Stderr = pty.Output()

		err = root.WithContext(ctx).Run()
		require.Error(t, err)
		require.ErrorContains(t, err, "'email' failed on the 'email' tag")
		require.ErrorContains(t, err, "'username' failed on the 'username' tag")
	})
}
