package types

import (
	"testing"

	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/assert"
)

// TestSQLLabel tests label
func TestSQLLabel(t *testing.T) {
	// assert.Equal(t, len(ss.Databases[dbname].Tables), 6)
	// assert.Equal(t, len(ss.Databases[dbname].Tables["users"].Indexes), 2)
	var (
		labelNeedsCherryPick = "needs-cherry-pick-2.1"
		labelContributor     = "type/contributor ⭐️"
		labelDNM             = "status/DNM"
	)

	githubLabels := []*github.Label{
		{
			Name: &labelNeedsCherryPick,
		},
		{
			Name: &labelContributor,
		},
	}

	labels := ParseGithubLabels(githubLabels)

	assert.Equal(t, labels, Labels([]string{labelNeedsCherryPick, labelContributor}))
	assert.Equal(t, labels.String(), `["needs-cherry-pick-2.1","type/contributor ⭐️"]`)

	labels.AddLabel(labelDNM)
	assert.Equal(t, labels.String(), `["needs-cherry-pick-2.1","type/contributor ⭐️","status/DNM"]`)

	labels.DelLabel(labelContributor)
	assert.Equal(t, labels.String(), `["needs-cherry-pick-2.1","status/DNM"]`)

	labels.DelLabel(labelDNM)
	labels.DelLabel(labelNeedsCherryPick)
	assert.Equal(t, labels.String(), `[]`)
}

// TestParseLabel tests parsing labels from string
func TestParseLabel(t *testing.T) {
	var (
		labels     Labels
		testLabels = `["needs-cherry-pick-2.1","type/contributor ⭐️","status/DNM"]`
	)

	labels.Scan(testLabels)
	assert.Equal(t, labels.String(), testLabels)
}
