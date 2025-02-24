package database_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetDatabase(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		options := testutils.GetPgOptions(t)
		log := testutils.GetLog(t)

		parts := strings.SplitN(options.Addr, ":", 2)
		address, portStr := parts[0], parts[1]
		port, err := strconv.ParseInt(portStr, 10, 64)
		assert.NoError(t, err, "must be able to convert port to int")
		config := config.Configuration{
			PostgreSQL: config.PostgreSQL{
				Address:  address,
				Port:     int(port),
				Username: options.User,
				Password: options.Password,
				Database: options.Database,
				Migrate:  false,
			},
		}

		db, err := database.GetDatabase(log, config, nil)
		assert.NoError(t, err, "must be able to connect to database")
		assert.NoError(t, db.Ping(t.Context()), "must be able to ping database")
		assert.NoError(t, db.Close(), "must be able to close db connection")
	})
}
