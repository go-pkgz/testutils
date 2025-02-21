package containers

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLTestContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	t.Run("create and cleanup container", func(t *testing.T) {
		mysql := NewMySQLTestContainer(ctx, t)
		defer func() { require.NoError(t, mysql.Close(ctx)) }()

		assert.NotEmpty(t, mysql.Host)
		assert.NotEmpty(t, mysql.Port)
		assert.Equal(t, "root", mysql.User)
		assert.Equal(t, "test", mysql.Database)
	})

	t.Run("custom database container", func(t *testing.T) {
		mysql := NewMySQLTestContainerWithDB(ctx, t, "custom")
		defer func() { require.NoError(t, mysql.Close(ctx)) }()

		assert.Equal(t, "custom", mysql.Database)
		assert.NotEmpty(t, mysql.ConnectionString())
		assert.Contains(t, mysql.DSN(), "parseTime=true")
	})

	t.Run("container is accessible", func(t *testing.T) {
		mysql := NewMySQLTestContainer(ctx, t)
		defer func() { require.NoError(t, mysql.Close(ctx)) }()

		db, err := sql.Open("mysql", mysql.DSN())
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
		mysql1 := NewMySQLTestContainer(ctx, t)
		defer func() { require.NoError(t, mysql1.Close(ctx)) }()

		mysql2 := NewMySQLTestContainer(ctx, t)
		defer func() { require.NoError(t, mysql2.Close(ctx)) }()

		assert.NotEqual(t, mysql1.Port, mysql2.Port)
		assert.NotEqual(t, mysql1.ConnectionString(), mysql2.ConnectionString())

		// verify both containers are accessible
		db1, err := sql.Open("mysql", mysql1.DSN())
		require.NoError(t, err)
		defer db1.Close()

		db2, err := sql.Open("mysql", mysql2.DSN())
		require.NoError(t, err)
		defer db2.Close()

		require.NoError(t, db1.Ping())
		require.NoError(t, db2.Ping())
	})
}
