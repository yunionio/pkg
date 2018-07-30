package helm

import (
	"k8s.io/helm/pkg/repo"
)

type RepoFile struct {
	*repo.RepoFile
}

func (r *RepoFile) Get(name string) (*repo.Entry, error) {
	var ret *repo.Entry
	for _, rf := range r.Repositories {
		if rf.Name == name {
			ret = rf
		}
	}
	return ret, nil
}
