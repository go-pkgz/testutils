package containers

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHTestContainer(t *testing.T) {
	ctx := context.Background()

	t.Run("create and cleanup container", func(t *testing.T) {
		ssh := NewSSHTestContainer(ctx, t)
		defer func() { require.NoError(t, ssh.Close(ctx)) }()

		assert.NotEmpty(t, ssh.Host)
		assert.NotEmpty(t, ssh.Port)
		assert.Equal(t, "test", ssh.User)
	})

	t.Run("custom user container", func(t *testing.T) {
		ssh := NewSSHTestContainerWithUser(ctx, t, "custom")
		defer func() { require.NoError(t, ssh.Close(ctx)) }()

		assert.NotEmpty(t, ssh.Host)
		assert.NotEmpty(t, ssh.Port)
		assert.Equal(t, "custom", ssh.User)
	})

	t.Run("container is accessible", func(t *testing.T) {
		ssh := NewSSHTestContainer(ctx, t)
		defer func() { require.NoError(t, ssh.Close(ctx)) }()

		// use ssh-keyscan to verify the host is accessible
		// values come from the test container itself, so this is not vulnerable to command injection
		// ssh.Port and ssh.Host are generated by the test container and are not user input
		cmd := exec.Command("ssh-keyscan", "-p", ssh.Port.Port(), ssh.Host) // #nosec G204 -- these params are from our test container
		out, err := cmd.CombinedOutput()
		require.NoError(t, err)
		t.Logf("ssh-keyscan output: %s", out)
		assert.Contains(t, string(out), "ssh-", "should return ssh key")
	})

	t.Run("multiple containers", func(t *testing.T) {
		ssh1 := NewSSHTestContainer(ctx, t)
		defer func() { require.NoError(t, ssh1.Close(ctx)) }()

		ssh2 := NewSSHTestContainer(ctx, t)
		defer func() { require.NoError(t, ssh2.Close(ctx)) }()

		assert.NotEqual(t, ssh1.Port, ssh2.Port)
		assert.NotEqual(t, ssh1.Address(), ssh2.Address())
	})
}
