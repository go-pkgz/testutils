package containers

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresTestContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	t.Run("create and cleanup container", func(t *testing.T) {
		pg := NewPostgresTestContainer(ctx, t)
		defer func() { require.NoError(t, pg.Close(ctx)) }()

		assert.NotEmpty(t, pg.Host)
		assert.NotEmpty(t, pg.Port)
		assert.Equal(t, "postgres", pg.User)
		assert.Equal(t, "test", pg.Database)
	})

	t.Run("custom database container", func(t *testing.T) {
		pg := NewPostgresTestContainerWithDB(ctx, t, "custom")
		defer func() { require.NoError(t, pg.Close(ctx)) }()

		assert.Equal(t, "custom", pg.Database)
		assert.NotEmpty(t, pg.ConnectionString())
	})

	t.Run("container is accessible", func(t *testing.T) {
		pg := NewPostgresTestContainer(ctx, t)
		defer func() { require.NoError(t, pg.Close(ctx)) }()

		db, err := sql.Open("postgres", pg.ConnectionString())
		require.NoError(t, err)
		defer db.Close()

		err = db.Ping()
		require.NoError(t, err, "should be able to ping database")

		// verify we can execute queries
		var result int
		err = db.QueryRow("SELECT 1").Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("multiple containers", func(t *testing.T) {
		pg1 := NewPostgresTestContainer(ctx, t)
		defer func() { require.NoError(t, pg1.Close(ctx)) }()

		pg2 := NewPostgresTestContainer(ctx, t)
		defer func() { require.NoError(t, pg2.Close(ctx)) }()

		assert.NotEqual(t, pg1.Port, pg2.Port)
		assert.NotEqual(t, pg1.ConnectionString(), pg2.ConnectionString())

		// verify both containers are accessible
		db1, err := sql.Open("postgres", pg1.ConnectionString())
		require.NoError(t, err)
		defer db1.Close()

		db2, err := sql.Open("postgres", pg2.ConnectionString())
		require.NoError(t, err)
		defer db2.Close()

		require.NoError(t, db1.Ping())
		require.NoError(t, db2.Ping())
	})
}
