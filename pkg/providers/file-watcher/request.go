package file

import (
	"fmt"

	"github.com/pkg/errors"
)

func (w *Watcher) getNoticeChannel() string {
	if w.repo.WatchFileChannel != "" {
		return w.repo.WatchFileChannel
	}
	return w.repo.GithubBotChannel
}

func (w *Watcher) noticeUpdate(path, branch, sha1, sha2 string) error {
	if len(sha1) > 7 {
		sha1 = sha1[:7]
	}
	if len(sha2) > 7 {
		sha2 = sha2[:7]
	}
	diffURL := fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s", w.repo.Owner, w.repo.Repo, sha1, sha2)
	msg := fmt.Sprintf("%s/%s file `%s` at %s updated\n%s", w.repo.Owner, w.repo.Repo, path, branch, diffURL)
	err := w.opr.Slack.SendMessage(w.getNoticeChannel(), msg)
	return errors.Wrap(err, "send update")
}
