package repo

import (
	"io"
	"os"
	"regexp"
	"time"
)

// gitRepo is the main type of this lib
type gitRepo struct {
	path    string
	url     string
	sshKey  *sshKey
	verbose bool
	logger  func(format string, i ...interface{})
}

func (g *gitRepo) setUrl(s string) {
	g.url = s
}
func (g *gitRepo) setVerbose(s bool) {
	g.verbose = s
}
func (g *gitRepo) setLogger(s func(format string, i ...interface{})) {
	g.logger = s
}
func (g *gitRepo) setSSHKey(sshKey *sshKey) {
	g.sshKey = sshKey
}
func (g *gitRepo) setPath(s string) {
	g.path = s
}

type Repo interface {
	// FetchURL returns the git URL the the remote origin
	FetchURL() (string, error)
	// Name returns the name of the repo, deduced from the remote origin URL
	Name() (string, error)
	// LocalConfigGet returns data from the local git config
	LocalConfigGet(section, key string) (string, error)
	// LocalConfigSet set data in the local git config
	LocalConfigSet(section, key, value string) error
	// Commits returns all the commit between
	Commits(from, to string) ([]Commit, error)
	// GetCommit returns a commit
	GetCommit(hash string) (Commit, error)
	// GetCommitWithDiff return the commit data with the parsed diff
	GetCommitWithDiff(hash string) (Commit, error)
	Diff(hash string, filename string) (string, error)
	// ExistsDiff returns true if there are no commited diff in the repo.
	ExistsDiff() bool
	// LatestCommit returns the latest commit of the current branch
	LatestCommit() (Commit, error)
	// CurrentBranch returns the current branch
	CurrentBranch() (string, error)
	// VerifyTag returns the sha1 of the tag if exists, if it doesn't exist, it returns an error
	VerifyTag(tag string) (string, error)
	// FetchRemoteTag runs a git fetch then checkout the remote tag
	FetchRemoteTag(remote, tag string) error
	// LocalBranchExists returns if given branch exists locally and has upstream.
	LocalBranchExists(branch string) (exists, hasUpstream bool)
	// FetchRemoteBranch runs a git fetch then checkout the remote branch
	FetchRemoteBranch(remote, branch string) error
	// Checkout checkouts a branch on the local repository
	Checkout(branch string) error
	// CheckoutNewBranch checkouts a new branch on the local repository
	CheckoutNewBranch(branch string) error
	// DeleteBranch deletes a branch on the local repository
	DeleteBranch(branch string) error
	// Pull pulls a branch from a remote
	Pull(remote, branch string) error
	// ResetHard hard resets a ref
	ResetHard(hash string) error
	// DefaultBranch returns the default branch of the remote origin
	DefaultBranch() (string, error)
	// Glob returns the matching files in the repo
	Glob(s string) ([]string, error)
	// Open opens a file from the repo
	Open(s string) (*os.File, error)
	// Write writes a file in the repo
	Write(s string, content io.Reader) error
	// Add file contents to the index
	Add(s ...string) error
	// Remove file or directory
	Remove(s ...string) error
	// Commit the index
	Commit(m string, opts ...Option) error
	// Push (always with force) the branch
	Push(remote, branch string, opts ...Option) error
	// RemoteAdd run git remote add
	RemoteAdd(remote, branch, url string) error
	// RemoteShow run git remote show
	RemoteShow(remote string) (string, error)
	// Status run the git status command
	Status() (string, error)
	CurrentSnapshot() (map[string]File, error)
	HasDiverged() (bool, error)
	HookList() ([]string, error)
	DeleteHook(name string) error
	WriteHook(name string, content []byte) error
}

// Commit represent a git commit
type Commit struct {
	LongHash string
	Hash     string
	Author   string
	Subject  string
	Body     string
	Date     time.Time
	Files    map[string]File
}

type File struct {
	Filename   string
	Status     string
	Diff       string
	DiffDetail FileDiffDetail
}

type FileDiffDetail struct {
	Hunks []Hunk
}

func (d FileDiffDetail) Matches(regexp *regexp.Regexp) (hunks []Hunk, addedLinewMatch bool, removedLinewMatch bool) {
	for _, h := range d.Hunks {
		var hunkMatches bool
		for _, l := range h.RemovedLines {
			if regexp.MatchString(l) {
				removedLinewMatch = true
				break
			}
		}
		for _, l := range h.AddedLines {
			if regexp.MatchString(l) {
				addedLinewMatch = true
				break
			}
		}
		if hunkMatches {
			hunks = append(hunks, h)
		}
	}
	return hunks, addedLinewMatch, removedLinewMatch
}

type Hunk struct {
	Header       string
	Content      string
	RemovedLines []string
	AddedLines   []string
}

// CloneOpts is a optional structs for git clone command
type CloneOpts struct {
	Recursive               *bool
	NoStrictHostKeyChecking *bool
	Auth                    *AuthOpts
}

// AuthOpts is a optional structs for git command
type AuthOpts struct {
	Username   string
	Password   string
	PrivateKey *SSHKey
}

// SSHKey is a type for a ssh key
type SSHKey struct {
	Filename string
	Content  []byte
}

type BareRepo struct {
	repo Repo
}
