package cli_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/pretty"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/cli/cliui"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestLogin(t *testing.T) {
	t.Parallel()
	t.Run("InitialUserNoTTY", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		root, _ := clitest.New(t, "login", client.URL.String())
		err := root.Run()
		require.Error(t, err)
	})

	t.Run("InitialUserBadLoginURL", func(t *testing.T) {
		t.Parallel()
		badLoginURL := "https://fcca2077f06e68aaf9"
		root, _ := clitest.New(t, "login", badLoginURL)
		err := root.Run()
		errMsg := fmt.Sprintf("Failed to check server %q for first user, is the URL correct and is wirtual accessible from your browser?", badLoginURL)
		require.ErrorContains(t, err, errMsg)
	})

	t.Run("InitialUserNonWirtualURLFail", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer ts.Close()

		badLoginURL := ts.URL
		root, _ := clitest.New(t, "login", badLoginURL)
		err := root.Run()
		errMsg := fmt.Sprintf("Failed to check server %q for first user, is the URL correct and is wirtual accessible from your browser?", badLoginURL)
		require.ErrorContains(t, err, errMsg)
	})

	t.Run("InitialUserNonWirtualURLSuccess", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(wirtualsdk.BuildVersionHeader, "something")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer ts.Close()

		badLoginURL := ts.URL
		root, _ := clitest.New(t, "login", badLoginURL)
		err := root.Run()
		// this means we passed the check for a valid wirtual server
		require.ErrorContains(t, err, "the initial user cannot be created in non-interactive mode")
	})

	t.Run("InitialUserTTY", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		// The --force-tty flag is required on Windows, because the `isatty` library does not
		// accurately detect Windows ptys when they are not attached to a process:
		// https://github.com/mattn/go-isatty/issues/59
		doneChan := make(chan struct{})
		root, _ := clitest.New(t, "login", "--force-tty", client.URL.String())
		pty := ptytest.New(t).Attach(root)
		go func() {
			defer close(doneChan)
			err := root.Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			"first user?", "yes",
			"username", wirtualdtest.FirstUserParams.Username,
			"name", wirtualdtest.FirstUserParams.Name,
			"email", wirtualdtest.FirstUserParams.Email,
			"password", wirtualdtest.FirstUserParams.Password,
			"password", wirtualdtest.FirstUserParams.Password, // confirm
			"trial", "yes",
			"firstName", wirtualdtest.TrialUserParams.FirstName,
			"lastName", wirtualdtest.TrialUserParams.LastName,
			"phoneNumber", wirtualdtest.TrialUserParams.PhoneNumber,
			"jobTitle", wirtualdtest.TrialUserParams.JobTitle,
			"companyName", wirtualdtest.TrialUserParams.CompanyName,
			// `developers` and `country` `cliui.Select` automatically selects the first option during tests.
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)
			pty.WriteLine(value)
		}
		pty.ExpectMatch("Welcome to Wirtual")
		<-doneChan
		ctx := testutil.Context(t, testutil.WaitShort)
		resp, err := client.LoginWithPassword(ctx, wirtualsdk.LoginWithPasswordRequest{
			Email:    wirtualdtest.FirstUserParams.Email,
			Password: wirtualdtest.FirstUserParams.Password,
		})
		require.NoError(t, err)
		client.SetSessionToken(resp.SessionToken)
		me, err := client.User(ctx, wirtualsdk.Me)
		require.NoError(t, err)
		assert.Equal(t, wirtualdtest.FirstUserParams.Username, me.Username)
		assert.Equal(t, wirtualdtest.FirstUserParams.Name, me.Name)
		assert.Equal(t, wirtualdtest.FirstUserParams.Email, me.Email)
	})

	t.Run("InitialUserTTYWithNoTrial", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		// The --force-tty flag is required on Windows, because the `isatty` library does not
		// accurately detect Windows ptys when they are not attached to a process:
		// https://github.com/mattn/go-isatty/issues/59
		doneChan := make(chan struct{})
		root, _ := clitest.New(t, "login", "--force-tty", client.URL.String())
		pty := ptytest.New(t).Attach(root)
		go func() {
			defer close(doneChan)
			err := root.Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			"first user?", "yes",
			"username", wirtualdtest.FirstUserParams.Username,
			"name", wirtualdtest.FirstUserParams.Name,
			"email", wirtualdtest.FirstUserParams.Email,
			"password", wirtualdtest.FirstUserParams.Password,
			"password", wirtualdtest.FirstUserParams.Password, // confirm
			"trial", "no",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)
			pty.WriteLine(value)
		}
		pty.ExpectMatch("Welcome to Wirtual")
		<-doneChan
		ctx := testutil.Context(t, testutil.WaitShort)
		resp, err := client.LoginWithPassword(ctx, wirtualsdk.LoginWithPasswordRequest{
			Email:    wirtualdtest.FirstUserParams.Email,
			Password: wirtualdtest.FirstUserParams.Password,
		})
		require.NoError(t, err)
		client.SetSessionToken(resp.SessionToken)
		me, err := client.User(ctx, wirtualsdk.Me)
		require.NoError(t, err)
		assert.Equal(t, wirtualdtest.FirstUserParams.Username, me.Username)
		assert.Equal(t, wirtualdtest.FirstUserParams.Name, me.Name)
		assert.Equal(t, wirtualdtest.FirstUserParams.Email, me.Email)
	})

	t.Run("InitialUserTTYNameOptional", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		// The --force-tty flag is required on Windows, because the `isatty` library does not
		// accurately detect Windows ptys when they are not attached to a process:
		// https://github.com/mattn/go-isatty/issues/59
		doneChan := make(chan struct{})
		root, _ := clitest.New(t, "login", "--force-tty", client.URL.String())
		pty := ptytest.New(t).Attach(root)
		go func() {
			defer close(doneChan)
			err := root.Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			"first user?", "yes",
			"username", wirtualdtest.FirstUserParams.Username,
			"name", "",
			"email", wirtualdtest.FirstUserParams.Email,
			"password", wirtualdtest.FirstUserParams.Password,
			"password", wirtualdtest.FirstUserParams.Password, // confirm
			"trial", "yes",
			"firstName", wirtualdtest.TrialUserParams.FirstName,
			"lastName", wirtualdtest.TrialUserParams.LastName,
			"phoneNumber", wirtualdtest.TrialUserParams.PhoneNumber,
			"jobTitle", wirtualdtest.TrialUserParams.JobTitle,
			"companyName", wirtualdtest.TrialUserParams.CompanyName,
			// `developers` and `country` `cliui.Select` automatically selects the first option during tests.
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)
			pty.WriteLine(value)
		}
		pty.ExpectMatch("Welcome to Wirtual")
		<-doneChan
		ctx := testutil.Context(t, testutil.WaitShort)
		resp, err := client.LoginWithPassword(ctx, wirtualsdk.LoginWithPasswordRequest{
			Email:    wirtualdtest.FirstUserParams.Email,
			Password: wirtualdtest.FirstUserParams.Password,
		})
		require.NoError(t, err)
		client.SetSessionToken(resp.SessionToken)
		me, err := client.User(ctx, wirtualsdk.Me)
		require.NoError(t, err)
		assert.Equal(t, wirtualdtest.FirstUserParams.Username, me.Username)
		assert.Equal(t, wirtualdtest.FirstUserParams.Email, me.Email)
		assert.Empty(t, me.Name)
	})

	t.Run("InitialUserTTYFlag", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		// The --force-tty flag is required on Windows, because the `isatty` library does not
		// accurately detect Windows ptys when they are not attached to a process:
		// https://github.com/mattn/go-isatty/issues/59
		inv, _ := clitest.New(t, "--url", client.URL.String(), "login", "--force-tty")
		pty := ptytest.New(t).Attach(inv)

		clitest.Start(t, inv)

		pty.ExpectMatch(fmt.Sprintf("Attempting to authenticate with flag URL: '%s'", client.URL.String()))
		matches := []string{
			"first user?", "yes",
			"username", wirtualdtest.FirstUserParams.Username,
			"name", wirtualdtest.FirstUserParams.Name,
			"email", wirtualdtest.FirstUserParams.Email,
			"password", wirtualdtest.FirstUserParams.Password,
			"password", wirtualdtest.FirstUserParams.Password, // confirm
			"trial", "yes",
			"firstName", wirtualdtest.TrialUserParams.FirstName,
			"lastName", wirtualdtest.TrialUserParams.LastName,
			"phoneNumber", wirtualdtest.TrialUserParams.PhoneNumber,
			"jobTitle", wirtualdtest.TrialUserParams.JobTitle,
			"companyName", wirtualdtest.TrialUserParams.CompanyName,
			// `developers` and `country` `cliui.Select` automatically selects the first option during tests.
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)
			pty.WriteLine(value)
		}
		pty.ExpectMatch("Welcome to Wirtual")
		ctx := testutil.Context(t, testutil.WaitShort)
		resp, err := client.LoginWithPassword(ctx, wirtualsdk.LoginWithPasswordRequest{
			Email:    wirtualdtest.FirstUserParams.Email,
			Password: wirtualdtest.FirstUserParams.Password,
		})
		require.NoError(t, err)
		client.SetSessionToken(resp.SessionToken)
		me, err := client.User(ctx, wirtualsdk.Me)
		require.NoError(t, err)
		assert.Equal(t, wirtualdtest.FirstUserParams.Username, me.Username)
		assert.Equal(t, wirtualdtest.FirstUserParams.Name, me.Name)
		assert.Equal(t, wirtualdtest.FirstUserParams.Email, me.Email)
	})

	t.Run("InitialUserFlags", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		inv, _ := clitest.New(
			t, "login", client.URL.String(),
			"--first-user-username", wirtualdtest.FirstUserParams.Username,
			"--first-user-full-name", wirtualdtest.FirstUserParams.Name,
			"--first-user-email", wirtualdtest.FirstUserParams.Email,
			"--first-user-password", wirtualdtest.FirstUserParams.Password,
			"--first-user-trial",
		)
		pty := ptytest.New(t).Attach(inv)
		w := clitest.StartWithWaiter(t, inv)
		pty.ExpectMatch("firstName")
		pty.WriteLine(wirtualdtest.TrialUserParams.FirstName)
		pty.ExpectMatch("lastName")
		pty.WriteLine(wirtualdtest.TrialUserParams.LastName)
		pty.ExpectMatch("phoneNumber")
		pty.WriteLine(wirtualdtest.TrialUserParams.PhoneNumber)
		pty.ExpectMatch("jobTitle")
		pty.WriteLine(wirtualdtest.TrialUserParams.JobTitle)
		pty.ExpectMatch("companyName")
		pty.WriteLine(wirtualdtest.TrialUserParams.CompanyName)
		// `developers` and `country` `cliui.Select` automatically selects the first option during tests.
		pty.ExpectMatch("Welcome to Wirtual")
		w.RequireSuccess()
		ctx := testutil.Context(t, testutil.WaitShort)
		resp, err := client.LoginWithPassword(ctx, wirtualsdk.LoginWithPasswordRequest{
			Email:    wirtualdtest.FirstUserParams.Email,
			Password: wirtualdtest.FirstUserParams.Password,
		})
		require.NoError(t, err)
		client.SetSessionToken(resp.SessionToken)
		me, err := client.User(ctx, wirtualsdk.Me)
		require.NoError(t, err)
		assert.Equal(t, wirtualdtest.FirstUserParams.Username, me.Username)
		assert.Equal(t, wirtualdtest.FirstUserParams.Name, me.Name)
		assert.Equal(t, wirtualdtest.FirstUserParams.Email, me.Email)
	})

	t.Run("InitialUserFlagsNameOptional", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		inv, _ := clitest.New(
			t, "login", client.URL.String(),
			"--first-user-username", wirtualdtest.FirstUserParams.Username,
			"--first-user-email", wirtualdtest.FirstUserParams.Email,
			"--first-user-password", wirtualdtest.FirstUserParams.Password,
			"--first-user-trial",
		)
		pty := ptytest.New(t).Attach(inv)
		w := clitest.StartWithWaiter(t, inv)
		pty.ExpectMatch("firstName")
		pty.WriteLine(wirtualdtest.TrialUserParams.FirstName)
		pty.ExpectMatch("lastName")
		pty.WriteLine(wirtualdtest.TrialUserParams.LastName)
		pty.ExpectMatch("phoneNumber")
		pty.WriteLine(wirtualdtest.TrialUserParams.PhoneNumber)
		pty.ExpectMatch("jobTitle")
		pty.WriteLine(wirtualdtest.TrialUserParams.JobTitle)
		pty.ExpectMatch("companyName")
		pty.WriteLine(wirtualdtest.TrialUserParams.CompanyName)
		// `developers` and `country` `cliui.Select` automatically selects the first option during tests.
		pty.ExpectMatch("Welcome to Wirtual")
		w.RequireSuccess()
		ctx := testutil.Context(t, testutil.WaitShort)
		resp, err := client.LoginWithPassword(ctx, wirtualsdk.LoginWithPasswordRequest{
			Email:    wirtualdtest.FirstUserParams.Email,
			Password: wirtualdtest.FirstUserParams.Password,
		})
		require.NoError(t, err)
		client.SetSessionToken(resp.SessionToken)
		me, err := client.User(ctx, wirtualsdk.Me)
		require.NoError(t, err)
		assert.Equal(t, wirtualdtest.FirstUserParams.Username, me.Username)
		assert.Equal(t, wirtualdtest.FirstUserParams.Email, me.Email)
		assert.Empty(t, me.Name)
	})

	t.Run("InitialUserTTYConfirmPasswordFailAndReprompt", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		client := wirtualdtest.New(t, nil)
		// The --force-tty flag is required on Windows, because the `isatty` library does not
		// accurately detect Windows ptys when they are not attached to a process:
		// https://github.com/mattn/go-isatty/issues/59
		doneChan := make(chan struct{})
		root, _ := clitest.New(t, "login", "--force-tty", client.URL.String())
		pty := ptytest.New(t).Attach(root)
		go func() {
			defer close(doneChan)
			err := root.WithContext(ctx).Run()
			assert.NoError(t, err)
		}()

		matches := []string{
			"first user?", "yes",
			"username", wirtualdtest.FirstUserParams.Username,
			"name", wirtualdtest.FirstUserParams.Name,
			"email", wirtualdtest.FirstUserParams.Email,
			"password", wirtualdtest.FirstUserParams.Password,
			"password", "something completely different",
		}
		for i := 0; i < len(matches); i += 2 {
			match := matches[i]
			value := matches[i+1]
			pty.ExpectMatch(match)
			pty.WriteLine(value)
		}

		// Validate that we reprompt for matching passwords.
		pty.ExpectMatch("Passwords do not match")
		pty.ExpectMatch("Enter a " + pretty.Sprint(cliui.DefaultStyles.Field, "password"))
		pty.WriteLine(wirtualdtest.FirstUserParams.Password)
		pty.ExpectMatch("Confirm")
		pty.WriteLine(wirtualdtest.FirstUserParams.Password)
		pty.ExpectMatch("trial")
		pty.WriteLine("yes")
		pty.ExpectMatch("firstName")
		pty.WriteLine(wirtualdtest.TrialUserParams.FirstName)
		pty.ExpectMatch("lastName")
		pty.WriteLine(wirtualdtest.TrialUserParams.LastName)
		pty.ExpectMatch("phoneNumber")
		pty.WriteLine(wirtualdtest.TrialUserParams.PhoneNumber)
		pty.ExpectMatch("jobTitle")
		pty.WriteLine(wirtualdtest.TrialUserParams.JobTitle)
		pty.ExpectMatch("companyName")
		pty.WriteLine(wirtualdtest.TrialUserParams.CompanyName)
		pty.ExpectMatch("Welcome to Wirtual")
		<-doneChan
	})

	t.Run("ExistingUserValidTokenTTY", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		wirtualdtest.CreateFirstUser(t, client)

		doneChan := make(chan struct{})
		root, _ := clitest.New(t, "login", "--force-tty", client.URL.String(), "--no-open")
		pty := ptytest.New(t).Attach(root)
		go func() {
			defer close(doneChan)
			err := root.Run()
			assert.NoError(t, err)
		}()

		pty.ExpectMatch(fmt.Sprintf("Attempting to authenticate with argument URL: '%s'", client.URL.String()))
		pty.ExpectMatch("Paste your token here:")
		pty.WriteLine(client.SessionToken())
		if runtime.GOOS != "windows" {
			// For some reason, the match does not show up on Windows.
			pty.ExpectMatch(client.SessionToken())
		}
		pty.ExpectMatch("Welcome to Wirtual")
		<-doneChan
	})

	t.Run("ExistingUserURLSavedInConfig", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		url := client.URL.String()
		wirtualdtest.CreateFirstUser(t, client)

		inv, root := clitest.New(t, "login", "--no-open")
		clitest.SetupConfig(t, client, root)

		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		pty.ExpectMatch(fmt.Sprintf("Attempting to authenticate with config URL: '%s'", url))
		pty.ExpectMatch("Paste your token here:")
		pty.WriteLine(client.SessionToken())
		<-doneChan
	})

	t.Run("ExistingUserURLSavedInEnv", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		url := client.URL.String()
		wirtualdtest.CreateFirstUser(t, client)

		inv, _ := clitest.New(t, "login", "--no-open")
		inv.Environ.Set("WIRTUAL_URL", url)

		doneChan := make(chan struct{})
		pty := ptytest.New(t).Attach(inv)
		go func() {
			defer close(doneChan)
			err := inv.Run()
			assert.NoError(t, err)
		}()

		pty.ExpectMatch(fmt.Sprintf("Attempting to authenticate with environment URL: '%s'", url))
		pty.ExpectMatch("Paste your token here:")
		pty.WriteLine(client.SessionToken())
		<-doneChan
	})

	t.Run("ExistingUserInvalidTokenTTY", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		wirtualdtest.CreateFirstUser(t, client)

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()
		doneChan := make(chan struct{})
		root, _ := clitest.New(t, "login", client.URL.String(), "--no-open")
		pty := ptytest.New(t).Attach(root)
		go func() {
			defer close(doneChan)
			err := root.WithContext(ctx).Run()
			// An error is expected in this case, since the login wasn't successful:
			assert.Error(t, err)
		}()

		pty.ExpectMatch("Paste your token here:")
		pty.WriteLine("an-invalid-token")
		if runtime.GOOS != "windows" {
			// For some reason, the match does not show up on Windows.
			pty.ExpectMatch("an-invalid-token")
		}
		pty.ExpectMatch("That's not a valid token!")
		cancelFunc()
		<-doneChan
	})

	// TokenFlag should generate a new session token and store it in the session file.
	t.Run("TokenFlag", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		wirtualdtest.CreateFirstUser(t, client)
		root, cfg := clitest.New(t, "login", client.URL.String(), "--token", client.SessionToken())
		err := root.Run()
		require.NoError(t, err)
		sessionFile, err := cfg.Session().Read()
		require.NoError(t, err)
		// This **should not be equal** to the token we passed in.
		require.NotEqual(t, client.SessionToken(), sessionFile)
	})

	t.Run("KeepOrganizationContext", func(t *testing.T) {
		t.Parallel()
		client := wirtualdtest.New(t, nil)
		first := wirtualdtest.CreateFirstUser(t, client)
		root, cfg := clitest.New(t, "login", client.URL.String(), "--token", client.SessionToken())

		err := cfg.Organization().Write(first.OrganizationID.String())
		require.NoError(t, err, "write bad org to config")

		err = root.Run()
		require.NoError(t, err)
		sessionFile, err := cfg.Session().Read()
		require.NoError(t, err)
		require.NotEqual(t, client.SessionToken(), sessionFile)

		// Organization config should be deleted since the org does not exist
		selected, err := cfg.Organization().Read()
		require.NoError(t, err)
		require.Equal(t, selected, first.OrganizationID.String())
	})
}
