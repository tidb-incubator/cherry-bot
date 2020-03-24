package config

import (
	"path"
	"runtime"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func GetTestConfig() (*Config, error) {
	_, localFile, _, _ := runtime.Caller(0)
	pathStr := path.Join(path.Dir(localFile), "config.example.toml")
	cfg, err := GetConfig(&pathStr)
	if err != nil {
		return nil, errors.Wrap(err, "get test config")
	}
	return cfg, nil
}

func TestLoadConfig(t *testing.T) {
	cfg, err := GetTestConfig()
	assert.Nil(t, err)
	assert.Equal(t, err, nil, "create bot failed")
	assert.Equal(t, cfg.Database.Address, "127.0.0.1", "wrong address")
	assert.Equal(t, cfg.Repos["owner-repo"].WebhookSecret, "secret", "wrong repo secret")
	assert.Equal(t, cfg.Repos["owner-repo"].WatchFileChannel, "watch-file-channel", "read watch-file-channel")
	assert.Equal(t, cfg.Repos["owner_1-repo_1"].WebhookSecret, "secret_1", "wrong repo secret")
	assert.Equal(t, cfg.Repos["owner_1-repo_1"].WatchFileChannel, "watch-file-channel-1", "read watch-file-channel")
}
