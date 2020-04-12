package repo

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// CloneBare a git bare repository from the specified url to the destination path. Use Options to force the use of SSH Key and or PGP Key on this repo
func CloneBare(path, url string, opts ...Option) (Repo, error) {
	r := gitRepo{path: path, url: url}
	for _, f := range opts {
		if err := f(&r); err != nil {
			return r, err
		}
	}
	if r.verbose {
		r.log("Cloning %s\n", r.url)
	}
	_, err := r.runCmd("git", "clone", "--bare", r.url, ".")
	if err != nil {
		return r, err
	}
	return r, nil
}

// NewBare instanciance a bare repo instance from the path assuming the repo has already been cloned in.
func NewBare(path string, opts ...Option) (b BareRepo, err error) {
	b = BareRepo{&gitRepo{path: path}}
	p, err := findRefsDirectory(path)
	b.repo.(*gitRepo).setPath(p)
	if err != nil {
		return b, err
	}

	output, _ := b.repo.(*gitRepo).runCmd("git", "rev-parse", "--is-bare-repository")
	if !strings.Contains(output, "true") {
		return b, errors.New("path is not a bare repository")
	}

	for _, f := range opts {
		if err := f(b.repo); err != nil {
			return b, err
		}
	}

	return b, nil
}

func findRefsDirectory(p string) (string, error) {
	p = path.Join(p)
	p, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	if p == string(filepath.Separator) {
		return "", errors.New("refs directory not found")
	}

	if checkRefsDirectory(p) {
		return p, nil
	}

	parent := filepath.Dir(p)
	return findRefsDirectory(parent)
}

func checkRefsDirectory(path string) bool {
	dotGit := filepath.Join(path, "refs")
	if _, err := os.Stat(dotGit); err != nil || os.IsNotExist(err) {
		return false
	}
	return true
}

func (b BareRepo) ListFiles() ([]string, error) {
	output, err := b.repo.(*gitRepo).runCmd("git", "ls-tree", "--full-tree", "--name-only", "-r", "HEAD")
	if err != nil {
		return nil, err
	}
	output = strings.TrimSpace(output)
	files := strings.Split(output, "\n")
	return files, nil
}

var singleSpacePattern = regexp.MustCompile(`\s+`)

func (b BareRepo) FileSize(filename string) (int64, error) {

	output, err := b.repo.(*gitRepo).runCmd("git", "ls-tree", "--full-tree", "--long", "-r", "HEAD")
	if err != nil {
		return -1, err
	}
	output = strings.TrimSpace(output)
	files := strings.Split(output, "\n")
	for _, file := range files {
		file = strings.Replace(file, "\t", " ", -1)
		file = singleSpacePattern.ReplaceAllString(file, " ")
		tuple := strings.SplitN(file, " ", 5)
		if len(tuple) != 5 {
			return -1, errors.New("unable to file size: " + file)
		}
		if tuple[4] == filename {
			return strconv.ParseInt(tuple[3], 10, 64)
		}
	}
	return -1, errors.New("unable to file size: file " + filename + " not found")
}

func (b BareRepo) ReadFile(filename string) (io.Reader, error) {
	output, err := b.repo.(*gitRepo).runCmd("git", "show", "HEAD:"+filename)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(output), nil
}
