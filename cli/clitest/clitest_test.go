package clitest_test

import (
	"testing"

	"go.uber.org/goleak"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestCli(t *testing.T) {
	t.Parallel()
	clitest.CreateTemplateVersionSource(t, nil)
	client := wirtualdtest.New(t, nil)
	i, config := clitest.New(t)
	clitest.SetupConfig(t, client, config)
	pty := ptytest.New(t).Attach(i)
	clitest.Start(t, i)
	pty.ExpectMatch("wirtual")
}
