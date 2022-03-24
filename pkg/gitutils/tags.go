package gitutils

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/pkg/errors"
	"sort"
)

var (
	ErrGetTag   = fmt.Errorf("error getting tag")
	ErrOpenRepo = fmt.Errorf("error opening repo")
)

type RepoTags struct {
	Latest *semver.Version
	All    []*semver.Version
}

func (r *RepoTags) GetVersionsArray() []string {
	versions := make([]string, len(r.All))
	for i, v := range r.All {
		versions[i] = v.String()
	}
	return versions
}

func GetTags(path string) (*RepoTags, error) {
	tags, err := getRawGitTags(path)
	if err != nil {
		return nil, err
	}
	var versions []*semver.Version
	for {
		tag, err := tags.Next()
		if err != nil || tag == nil || !tag.Name().IsTag() {
			break
		}
		tagName := tag.Name().Short()
		v, err := semver.NewVersion(tagName)
		if err == nil {
			versions = append(versions, v)
		}
	}
	if len(versions) == 0 {
		return nil, errors.Wrap(ErrGetTag, "no tags found")
	}
	versions = sortVersions(versions)
	return &RepoTags{
		Latest: versions[0],
		All:    versions,
	}, nil
}

func sortVersions(versions []*semver.Version) []*semver.Version {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})
	return versions
}

func getRawGitTags(path string) (storer.ReferenceIter, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, errors.Wrap(ErrOpenRepo, err.Error())
	}
	tags, err := r.Tags()
	if err != nil {
		return nil, errors.Wrap(ErrGetTag, err.Error())
	}
	return tags, nil
}
