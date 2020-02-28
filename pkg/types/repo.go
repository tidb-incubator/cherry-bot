package types

// Repo struct
type Repo struct {
	Owner string
	Repo  string
}

// GetOwner gets owner
func (r *Repo) GetOwner() string {
	return r.Owner
}

// GetRepo gets repo name
func (r *Repo) GetRepo() string {
	return r.Repo
}
