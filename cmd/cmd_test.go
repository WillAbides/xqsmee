package cmd

import (
	"testing"

	"io/ioutil"
	"path/filepath"

	"github.com/WillAbides/xqsmee/queue/redisqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireReadFile(t *testing.T, filename string) []byte {
	t.Helper()
	b, err := ioutil.ReadFile(filename)
	require.Nil(t, err)
	return b
}

func testdata(filename string) string {
	return filepath.Join("../testdata", filename)
}

func Test_cmdCfg_serverConfig(t *testing.T) {
	cfg := &srvCmdCfg{
		NoTLS:       true,
		RedisPrefix: "foo",
		RedisURL:    "redisurl",
		MaxActive:   13,
	}
	srv, err := cfg.serverConfig()
	assert.Nil(t, err)
	redisQueue, ok := srv.Queue.(*redisqueue.Queue)
	assert.True(t, ok)
	assert.Equal(t, 13, redisQueue.Pool.MaxActive)
	assert.Equal(t, "foo", redisQueue.Prefix)
}

func Test_tlsData(t *testing.T) {
	tdd := []struct {
		noTLS       bool
		tlsCertFile string
		tlsKeyFile  string
		exTlsCert   []byte
		exTlsKey    []byte
		exErr       string
	}{
		{
			noTLS: true,
		},
		{
			exErr: "you must specify both --tlskey and --tlscert unless --no-tls is set",
		},
		{
			tlsCertFile: testdata("server.crt"),
			exErr:       "you must specify both --tlskey and --tlscert unless --no-tls is set",
		},
		{
			tlsCertFile: testdata("server.crt"),
			tlsKeyFile:  testdata("server.key"),
			exTlsCert:   requireReadFile(t, testdata("server.crt")),
			exTlsKey:    requireReadFile(t, testdata("server.key")),
		},
		{
			tlsCertFile: "doesnotexist",
			tlsKeyFile:  "doesnotexist",
			exErr:       "failed reading tls certificate file: open doesnotexist: no such file or directory",
		},
		{
			tlsCertFile: testdata("server.crt"),
			tlsKeyFile:  "doesnotexist",
			exErr:       "failed reading tls key file: open doesnotexist: no such file or directory",
		},
	}

	for _, td := range tdd {
		t.Run("", func(t *testing.T) {
			gotTlsCert, gotTlsKey, gotErr := tlsData(td.noTLS, td.tlsCertFile, td.tlsKeyFile)
			if td.exErr == "" {
				assert.Nil(t, gotErr)
			} else {
				assert.EqualError(t, gotErr, td.exErr)
			}
			assert.Equal(t, td.exTlsCert, gotTlsCert)
			assert.Equal(t, td.exTlsKey, gotTlsKey)
		})
	}
}
