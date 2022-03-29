package nxtv

import (
	"fmt"
	"github.com/Eitol/nxtv/pkg/gitutils"
	"github.com/Masterminds/semver"
	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
	"github.com/pkg/errors"
	"strings"
)

var (
	ErrTheBranchesMustBeNotEq = fmt.Errorf("the branches must be not equals")
	ErrNoDiffBetweenBranches  = fmt.Errorf("no diff between branches")
	ErrGettingTags            = fmt.Errorf("getting tags")
)

type VersionUpgradeType string

const (
	PatchVersionUpgrade VersionUpgradeType = "patch"
	MinorVersionUpgrade VersionUpgradeType = "minor"
	MajorVersionUpgrade VersionUpgradeType = "major"
)

type Output struct {
	Error             string             `json:"error,omitempty" bson:"error"`
	Versions          []string           `json:"versions" bson:"versions"`
	LatestVersion     string             `json:"latestVersion" bson:"latestVersion"`
	RelevantCommitMsg string             `json:"relevantCommitMsg" bson:"relevantCommitMsg"`
	UpgradeType       VersionUpgradeType `json:"upgradeType" bson:"upgradeType"`
	NextVersion       string             `json:"nextVersion" bson:"nextVersion"`
}

func GetNextVersionBasedOnMR(path, sourceBranch, targetBranch string) (*Output, error) {
	if targetBranch == sourceBranch {
		return nil, ErrTheBranchesMustBeNotEq
	}
	tags, err := gitutils.GetTags(path)
	if err != nil {
		return nil, errors.Wrap(err, ErrGettingTags.Error())
	}
	tCommits, err := gitutils.GetCommits(path, sourceBranch, targetBranch)
	if err != nil {
		return nil, err
	}
	sCommits, err := gitutils.GetCommits(path, targetBranch, sourceBranch)
	if err != nil {
		return nil, err
	}
	var diffCommits []gitutils.Commit
	for _, s := range sCommits {
		exist := false
		for _, t := range tCommits {
			if s.Hash == t.Hash {
				exist = true
				break
			}
		}
		if !exist {
			diffCommits = append(diffCommits, s)
		}
	}
	if len(diffCommits) == 0 {
		return nil, ErrNoDiffBetweenBranches
	}
	versionUpgradeType, relevantCommit := getUpgradeTypeAndRelevantCommitFromDiff(diffCommits)
	if relevantCommit == "" {
		relevantCommit = strings.TrimSpace(diffCommits[0].Message)
	}
	latestVersion := tags.Latest.String()
	nextVersion := increaseVersion(latestVersion, versionUpgradeType)
	return &Output{
		Versions:          tags.GetVersionsArray(),
		LatestVersion:     latestVersion,
		NextVersion:       nextVersion,
		UpgradeType:       versionUpgradeType,
		RelevantCommitMsg: relevantCommit,
	}, nil
}

func getUpgradeTypeAndRelevantCommitFromDiff(diffCommits []gitutils.Commit) (VersionUpgradeType, string) {
	var versionUpgradeType = PatchVersionUpgrade
	var relevantCommit string
	for _, c := range diffCommits {
		msg := strings.TrimSpace(c.Message)
		m, err := parser.NewMachine().Parse([]byte(msg))
		if err == nil {
			cc, ok := m.(*conventionalcommits.ConventionalCommit)
			if ok {
				if cc.IsBreakingChange() {
					relevantCommit = msg
					versionUpgradeType = MajorVersionUpgrade
				} else if cc.Type == "feat" && versionUpgradeType != MajorVersionUpgrade {
					relevantCommit = msg
					versionUpgradeType = MinorVersionUpgrade
				}
			}
		}
	}
	relevantCommit = strings.Split(relevantCommit, "\n")[0]
	return versionUpgradeType, relevantCommit
}

func increaseVersion(latestVersion string, versionUpgradeType VersionUpgradeType) string {
	nextVersionSemver, _ := semver.NewVersion(latestVersion)
	if versionUpgradeType == PatchVersionUpgrade {
		*nextVersionSemver = nextVersionSemver.IncPatch()
	} else if versionUpgradeType == MinorVersionUpgrade {
		*nextVersionSemver = nextVersionSemver.IncMinor()
	} else if versionUpgradeType == MajorVersionUpgrade {
		*nextVersionSemver = nextVersionSemver.IncMajor()
	}
	return nextVersionSemver.String()
}
