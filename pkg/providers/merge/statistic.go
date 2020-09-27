package merge

import "fmt"

// StatusStatistic is the merge status statistic for a repo
type StatusStatistic struct {
	Owner       string  `json:"owner"`
	Repo        string  `json:"repo"`
	SuccessRate float32 `json:"success_rate"`
	Incomplete  uint64  `json:"incomplete"`
	Finish      uint64  `json:"finish"`
	Success     uint64  `json:"success"`
	TestFail    uint64  `json:"test_fail"`
	MergeFail   uint64  `json:"merge_fail"`
}

// StatusCount is the count for one status of a repo
type StatusCount struct {
	Owner  string `gorm:"column:owner"`
	Repo   string `gorm:"column:repo"`
	Status int    `gorm:"column:status"`
	Count  uint64 `gorm:"column:cnt"`
}

func MakeStatusStatistic(counts []*StatusCount) []*StatusStatistic {
	repoKey := func(owner, repo string) string {
		return fmt.Sprintf("%s/%s", owner, repo)
	}

	repoMap := make(map[string]*StatusStatistic)
	for _, count := range counts {
		key := repoKey(count.Owner, count.Repo)
		repo, ok := repoMap[key]
		if !ok {
			repo = &StatusStatistic{
				Owner: count.Owner,
				Repo:  count.Repo,
			}
			repoMap[key] = repo
		}

		switch count.Status {
		case mergeIncomplete:
			repo.Incomplete = count.Count
		case mergeFinish:
			repo.Finish = count.Count
		case mergeSuccess:
			repo.Success = count.Count
		case mergeTestFail:
			repo.TestFail = count.Count
		case mergeMergeFail:
			repo.MergeFail = count.Count
		}
	}

	statistic := make([]*StatusStatistic, 0, len(repoMap))
	for _, repo := range repoMap {
		repo.SuccessRate = float32(repo.Success) / float32(repo.Success+repo.TestFail+repo.MergeFail)
		statistic = append(statistic, repo)
	}
	return statistic
}

func (m *merge) StatisticRepo(owner, repo string) (*StatusStatistic, error) {
	results := m.opr.DB.Raw("select owner, repo, status, count(1) as cnt from `auto_merges` where owner = ? and repo = ? group by owner, repo, status", owner, repo)
	if results.Error != nil {
		return nil, results.Error
	}

	counts := make([]*StatusCount, 0)
	if err := results.Scan(&counts).Error; err != nil {
		return nil, err
	}

	if len(counts) == 0 {
		return &StatusStatistic{Owner: owner, Repo: repo}, nil
	}

	return MakeStatusStatistic(counts)[0], nil
}

func (m *merge) StatisticAllRepos() ([]*StatusStatistic, error) {
	results := m.opr.DB.Raw("select owner, repo, status, count(1) as cnt from `auto_merges` group by owner, repo, status")
	if results.Error != nil {
		return nil, results.Error
	}

	counts := make([]*StatusCount, 0)
	if err := results.Scan(&counts).Error; err != nil {
		return nil, err
	}

	if len(counts) == 0 {
		return nil, nil
	}

	return MakeStatusStatistic(counts), nil
}
