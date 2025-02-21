package containers

import (
	"context"
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
}
