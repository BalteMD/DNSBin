package db

import (
	"context"
	"dnsbin/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func createRandomDNSLog(t *testing.T) DnsLog {
	arg := CreateDNSLogParams{
		DnsQueryRecord: util.RandomString(12),
		Type:           util.RandomString(6),
		IpAddress:      util.RandomString(12),
		Location:       util.RandomString(6),
	}

	dnsLog, err := testStore.CreateDNSLog(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, dnsLog)

	require.Equal(t, arg.DnsQueryRecord, dnsLog.DnsQueryRecord)
	require.Equal(t, arg.Type, dnsLog.Type)
	require.Equal(t, arg.IpAddress, dnsLog.IpAddress)
	require.Equal(t, arg.Location, dnsLog.Location)
	require.NotZero(t, dnsLog.CreatedAt)

	return dnsLog
}

func TestCreateUser(t *testing.T) {
	createRandomDNSLog(t)
}
