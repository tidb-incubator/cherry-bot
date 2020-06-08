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
	assert.Equal(t, cfg.Repos["owner_2-repo_2"].WebhookSecret, "secret", "repo secret")
	assert.Equal(t, cfg.Repos["owner_2-repo_2"].WatchFileChannel, "watch-file-channel", "read watch-file-channel")
	assert.Equal(t, cfg.Repos["owner_1-repo_1"].WebhookSecret, "secret_1", "repo secret")
	assert.Equal(t, cfg.Repos["owner_1-repo_1"].WatchFileChannel, "watch-file-channel-1", "read watch-file-channel")
	assert.Equal(t, cfg.Repos["owner_1-repo_1"].CherryPickMilestone, true, "repo assign milestone field")
	assert.Equal(t, cfg.Repos["owner_2-repo_2"].CherryPickMilestone, false, "repo assign milestone field")
	assert.Equal(t, cfg.Repos["owner_1-repo_1"].CherryPickAssign, true, "repo assign pull field")
	assert.Equal(t, cfg.Repos["owner_2-repo_2"].CherryPickAssign, false, "repo assign pull field")
	assert.Equal(t, cfg.Repos["owner_1-repo_1"].MergeSIGControl, true, "merge sig control field")
	assert.Equal(t, cfg.Repos["owner_2-repo_2"].MergeSIGControl, false, "merge sig control field")
	assert.Equal(t, cfg.Member.Orgs, []string{"pingcap"}, "read member orgs")
	assert.Equal(t, cfg.Member.Users, []string{"sre-bot"}, "read member users")
}
