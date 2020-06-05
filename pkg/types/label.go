package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
)

// Labels for label parsing and manipulation
type Labels []string

// ParseGithubLabels parse label from github
func ParseGithubLabels(githubLabels []*github.Label) Labels {
	var labels Labels

	for _, l := range githubLabels {
		labels = append(labels, l.GetName())
	}

	return labels
}

// HasLabel in labels
func (labels Labels) HasLabel(label string) bool {
	for _, l := range labels {
		if l == label {
			return true
		}
	}
	return false
}

// AddLabel adds label
func (labels *Labels) AddLabel(label string) {
	for _, l := range *labels {
		if l == label {
			return
		}
	}
	*labels = append(*labels, label)
}

// DelLabel dels label
func (labels *Labels) DelLabel(label string) {
	labelStrs := []string(*labels)
	for i, l := range *labels {
		if l == label {
			labelStrs = append(labelStrs[:i], labelStrs[i+1:]...)
		}
	}
	*labels = Labels(labelStrs)
}

// String trans labels into string
func (labels Labels) String() string {
	b, err := json.Marshal(labels)
	if err != nil || string(b) == "null" {
		return "[]"
	}
	return string(b)
}

// Value for storage into string
func (labels Labels) Value() (driver.Value, error) {
	return labels.String(), nil
}

// Scan implements Scanner interface
func (labels *Labels) Scan(src interface{}) error {
	switch s := src.(type) {
	case string:
		return json.Unmarshal([]byte(s), labels)
	case []byte:
		return json.Unmarshal(s, labels)
	default:
		return errors.New("incorrect types")
	}
}
