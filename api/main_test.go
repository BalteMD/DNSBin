package api

import (
	db "dnsbin/db/sqlc"
	"dnsbin/notify"
	"dnsbin/util"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store, telegram *notify.Telegram) *Server {
	config, err := util.LoadConfig("../")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	server, err := NewServer(config, store, telegram)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
