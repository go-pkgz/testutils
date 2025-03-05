package containers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoTestContainer(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		version int
	}{
		{"mongo 5", 5},
		{"mongo 6", 6},
		{"mongo 7", 7},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mongo := NewMongoTestContainer(ctx, t, tc.version)
			defer func() { require.NoError(t, mongo.Close(ctx)) }()

			assert.NotEmpty(t, mongo.URI)
			assert.Contains(t, mongo.URI, "mongodb://")

			coll := mongo.Collection("test_db")
			_, err := coll.InsertOne(ctx, bson.M{"test": "value"})
			require.NoError(t, err)

			var result bson.M
			err = coll.FindOne(ctx, bson.M{"test": "value"}).Decode(&result)
			require.NoError(t, err)
			assert.Equal(t, "value", result["test"])
		})
	}

	t.Run("multiple collections in same container", func(t *testing.T) {
		mongo := NewMongoTestContainer(ctx, t, 7)
		defer func() { require.NoError(t, mongo.Close(ctx)) }()

		coll1 := mongo.Collection("test_db")
		coll2 := mongo.Collection("test_db")

		_, err := coll1.InsertOne(ctx, bson.M{"collection": "1"})
		require.NoError(t, err)
		_, err = coll2.InsertOne(ctx, bson.M{"collection": "2"})
		require.NoError(t, err)

		assert.NotEqual(t, coll1.Name(), coll2.Name())
	})

	t.Run("close with original environment variable", func(t *testing.T) {
		// save current MONGO_TEST value
		origEnv := os.Getenv("MONGO_TEST")
		testValue := "mongodb://original-value:27017"
		os.Setenv("MONGO_TEST", testValue)
		defer func() {
			// restore original value
			if origEnv == "" {
				os.Unsetenv("MONGO_TEST")
			} else {
				os.Setenv("MONGO_TEST", origEnv)
			}
		}()

		mongo := NewMongoTestContainer(ctx, t, 7)
		err := mongo.Close(ctx)
		require.NoError(t, err)

		// verify environment restored
		restoredValue := os.Getenv("MONGO_TEST")
		assert.Equal(t, testValue, restoredValue, "Original environment value should be restored")
	})

	t.Run("close with no original environment variable", func(t *testing.T) {
		// save current MONGO_TEST value
		origEnv := os.Getenv("MONGO_TEST")
		os.Unsetenv("MONGO_TEST")
		defer func() {
			// restore original value
			if origEnv != "" {
				os.Setenv("MONGO_TEST", origEnv)
			}
		}()

		mongo := NewMongoTestContainer(ctx, t, 7)
		err := mongo.Close(ctx)
		require.NoError(t, err)

		// verify environment is unset
		_, exists := os.LookupEnv("MONGO_TEST")
		assert.False(t, exists, "MONGO_TEST environment variable should be unset")
	})
}
