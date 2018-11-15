package server

import (
	"context"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	gitDir       = ".git"
	masterBranch = "master"
)

type FileSetPackage struct {
	P *doc.Package
	F *token.FileSet
}

type RepositoryPackage struct {
	FileSetPackage
	ImportPath string
}

type RepositoryPackages []RepositoryPackage

func (r RepositoryPackages) Len() int           { return len(r) }
func (r RepositoryPackages) Less(i, j int) bool { return r[i].ImportPath < r[j].ImportPath }
func (r RepositoryPackages) Swap(i, j int)      { i, j = j, i }
func (r RepositoryPackages) Sort()              { sort.Sort(r) }

type TaggedRepositoryPackage struct {
	RepositoryPackages
	Tag string
}

type TaggedRepositoryPackages []TaggedRepositoryPackage

func (t TaggedRepositoryPackages) Len() int           { return len(t) }
func (t TaggedRepositoryPackages) Less(i, j int) bool { return t[i].Tag < t[j].Tag }
func (t TaggedRepositoryPackages) Swap(i, j int)      { i, j = j, i }
func (t TaggedRepositoryPackages) Sort() {
	sort.Sort(t)
	for _, p := range t {
		p.RepositoryPackages.Sort()
	}
}

type repository struct {
	// Const at creation time
	cloneURL       *url.URL
	dest           string
	timeout        time.Duration
	errors         chan error
	taggedPackages chan TaggedRepositoryPackages
	beginSync      chan struct{}
	// Internal mutable state
	mu *sync.Mutex
}

func newRepository(cloneURL *url.URL, dest string, timeout time.Duration) *repository {
	if !strings.HasSuffix(dest, string(rune(os.PathSeparator))) {
		dest += string(os.PathSeparator)
	}
	return &repository{
		cloneURL:       cloneURL,
		dest:           dest,
		timeout:        timeout,
		errors:         make(chan error),
		taggedPackages: make(chan TaggedRepositoryPackages),
		beginSync:      make(chan struct{}),
		mu:             &sync.Mutex{},
	}
}

func (r *repository) Errors() <-chan error {
	return r.errors
}

func (r *repository) Packages() <-chan TaggedRepositoryPackages {
	return r.taggedPackages
}

func (r *repository) BeginSync() <-chan struct{} {
	return r.beginSync
}

func (r *repository) Sync() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := os.Stat(r.dest); os.IsNotExist(err) {
		if err := r.createDestFolder(); err != nil {
			r.errors <- err
			return
		}
	} else if err != nil {
		r.errors <- err
		return
	}

	tags, err := r.getTags()
	if err != nil {
		r.errors <- err
		return
	}
	r.beginSync <- struct{}{}
	tags = append(tags, masterBranch)
	results := make(TaggedRepositoryPackages, 0)
	for _, tag := range tags {
		if err := r.checkout(tag); err != nil {
			r.errors <- err
			continue
		}
		tagResult, err := r.getRepositoryPackages(tag)
		if err != nil {
			r.errors <- err
			continue
		}
		results = append(results, TaggedRepositoryPackage{
			RepositoryPackages: tagResult,
			Tag:                tag,
		})
	}
	if len(results) > 0 {
		r.taggedPackages <- results
	}
}

func (r *repository) getTags() ([]string, error) {
	if err := r.removeRelativeDirectory("repo4tags"); err != nil {
		return nil, err
	}
	_, _, err := r.exec("git", "clone", "--bare", r.cloneURL.String(), "repo4tags")
	if err != nil {
		return nil, err
	}
	stdout, _, err := r.execInSubdir("repo4tags", "git", "tag")
	s := strings.TrimSpace(string(stdout))
	return strings.Split(s, "\n"), err
}

func (r *repository) checkout(tag string) error {
	if err := r.removeRelativeDirectory(tag); err != nil {
		return err
	}
	_, _, err := r.exec("git", "clone", "--branch", tag, "--depth", "1", r.cloneURL.String(), tag)
	return err
}

func (r *repository) getRepositoryPackages(subdir string) (RepositoryPackages, error) {
	dirs, err := r.getSubdirsRecursively(filepath.Join(r.dest, subdir) + string(rune(os.PathSeparator)))
	if err != nil {
		return nil, err
	}
	prefix := strings.TrimSuffix(r.cloneURL.Host+r.cloneURL.Path, ".git")
	prefix += "/"
	pkgs := make(RepositoryPackages, 0)
	for _, dir := range dirs {
		fset := token.NewFileSet()
		pkg, err := parser.ParseDir(fset, dir, func(f os.FileInfo) bool {
			return !strings.HasSuffix(f.Name(), "_test.go")
		}, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		intermed := strings.TrimPrefix(dir, r.dest)
		intermed = strings.Replace(intermed, string(rune(os.PathSeparator)), "/", -1)
		intermed = strings.TrimSuffix(intermed, "/")
		for _, p := range pkg {
			importPath := prefix + intermed
			docPkg := doc.New(p, importPath, 0)
			pkgs = append(pkgs, RepositoryPackage{
				ImportPath: importPath,
				FileSetPackage: FileSetPackage{
					P: docPkg,
					F: fset,
				},
			})
		}
	}
	return pkgs, nil
}

func (r *repository) removeRelativeDirectory(relativePath string) error {
	return os.RemoveAll(filepath.Join(r.dest, relativePath))
}

func (r *repository) createDestFolder() error {
	if err := os.MkdirAll(r.dest, 0777); os.IsExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (r *repository) exec(cmdLine string, args ...string) ([]byte, []byte, error) {
	return r.execInSubdir("", cmdLine, args...)
}

func (r *repository) execInSubdir(subdir, cmdLine string, args ...string) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cmdLine, args...)
	cmd.Dir = filepath.Join(r.dest, subdir)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.errors <- err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		r.errors <- err
	}
	if err := cmd.Start(); err != nil {
		r.errors <- err
	}
	stdoutBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		r.errors <- err
	}
	stderrBytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		r.errors <- err
	}
	err = cmd.Wait()
	return stdoutBytes, stderrBytes, err
}

func (r *repository) getSubdirsRecursively(baseDir string) ([]string, error) {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		if f.Name()[0] == '.' {
			continue
		}
		subdir := baseDir + f.Name() + string(rune(os.PathSeparator))
		dirs = append(dirs, subdir)
		subdirs, err := r.getSubdirsRecursively(subdir)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, subdirs...)
	}
	return dirs, nil
}
