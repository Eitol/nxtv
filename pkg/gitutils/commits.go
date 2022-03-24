package gitutils

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrInCheckout     = errors.New("failed to checkout")
	ErrInGitLog       = errors.New("failed to get git log")
	ErrCheckingRef    = errors.New("failed to check ref")
	ErrGetWorkingTree = errors.New("failed to get working tree")
)

type Commit struct {
	Message string
	Date    time.Time
	Author  string
	Hash    string
}

func GetCommits(path, sourceBranch, targetBranch string) ([]Commit, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, errors.Wrap(ErrOpenRepo, err.Error())
	}
	wt, err := r.Worktree()
	sRef := plumbing.ReferenceName("refs/heads/" + sourceBranch)
	tRef := plumbing.ReferenceName("refs/heads/" + targetBranch)
	if err != nil {
		return nil, errors.Wrap(ErrGetWorkingTree, err.Error())
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Create: false,
		Branch: sRef,
		Force:  true,
	})
	if err != nil {
		return nil, errors.Wrap(ErrInCheckout, err.Error()+": branch "+sourceBranch)
	}
	trRef, err := r.Reference(tRef, true)
	if err != nil {
		return nil, errors.Wrap(ErrCheckingRef, err.Error()+": '"+targetBranch+"'")
	}
	log, err := r.Log(&git.LogOptions{
		From:  trRef.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, errors.Wrap(ErrInGitLog, err.Error()+": "+targetBranch)
	}
	var commits []Commit
	for {
		commit, err := log.Next()
		if err != nil {
			break
		}
		commits = append(commits, Commit{
			Message: commit.Message,
			Date:    commit.Author.When,
			Author:  commit.Author.Name,
			Hash:    commit.Hash.String(),
		})
	}
	return commits, nil
}
