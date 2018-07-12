package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	homePath       = "/"
	tutorialPath   = "/tutorial"
	repositoryPath = "/repo"
	tagPath        = "/tag"
	packagePath    = "/pkg"
	faviconPath    = "/favicon.ico"
)

type ProjectPackage struct {
	TaggedRepositoryPackages
	Project string
}

type ProjectPackages []ProjectPackage

func (p ProjectPackages) Len() int           { return len(p) }
func (p ProjectPackages) Less(i, j int) bool { return p[i].Project < p[j].Project }
func (p ProjectPackages) Swap(i, j int)      { i, j = j, i }
func (p ProjectPackages) Sort() {
	sort.Sort(p)
	for _, x := range p {
		x.TaggedRepositoryPackages.Sort()
	}
}

type RepositoryOptions struct {
	ProjectName         string
	HttpsCloneURL       *url.URL
	DiskCacheFilePath   string
	GitOperationTimeout time.Duration
}

type ServerOptions struct {
	TemplateFiles []string
	Repositories  []RepositoryOptions
	HttpServer    *http.Server
	RefreshRate   time.Duration
	Favicon       string
	SiteTitle     string
}

type Server struct {
	repo      map[string]*repository
	favicon   []byte
	siteTitle string
	renderer  *renderer
	mux       *http.ServeMux
	server    *http.Server
	refresh   time.Duration
}

func NewServer(opts ServerOptions) (*Server, error) {
	renderer, err := newRenderer(opts.TemplateFiles)
	if err != nil {
		return nil, err
	}
	favicon, err := ioutil.ReadFile(opts.Favicon)
	if err != nil {
		return nil, err
	}
	s := &Server{
		repo:      make(map[string]*repository),
		favicon:   favicon,
		siteTitle: opts.SiteTitle,
		renderer:  renderer,
		server:    opts.HttpServer,
		refresh:   opts.RefreshRate,
	}
	for _, repo := range opts.Repositories {
		s.repo[repo.ProjectName] = newRepository(repo.HttpsCloneURL, repo.DiskCacheFilePath, repo.GitOperationTimeout)
	}
	return s, nil
}

func (s *Server) createServeMux(projectPkgs ProjectPackages) {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc(faviconPath, func(w http.ResponseWriter, req *http.Request) {
		b := bytes.NewReader(s.favicon)
		http.ServeContent(w, req, "image/x-icon", time.Now(), b)
	})
	cd := &commonData{
		PageTitle:   s.siteTitle,
		RefreshTime: time.Now().UTC(),
	}
	d := &data{
		HomeData: &homeData{
			Path: homePath,
			C:    cd,
		},
		TutorialData: &tutorialData{
			Path: tutorialPath,
			C:    cd,
		},
		RepositoriesData: &repositoriesData{
			Path: repositoryPath,
			C:    cd,
		},
	}
	cd.Data = d
	s.mux.HandleFunc(homePath, func(w http.ResponseWriter, req *http.Request) {
		if homePath != req.URL.Path {
			if err := s.renderer.NotFound(w, cd); err != nil {
				log.Println(err)
			}
		} else {
			if err := s.renderer.Home(w, d.HomeData); err != nil {
				log.Println(err)
			}
		}
	})
	s.mux.HandleFunc(tutorialPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.Tutorial(w, d.TutorialData); err != nil {
			log.Println(err)
		}
	})
	s.mux.HandleFunc(repositoryPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.AllRepositories(w, d.RepositoriesData); err != nil {
			log.Println(err)
		}
	})
	for _, repoPkgs := range projectPkgs {
		myRepoPath := fmt.Sprintf("%s/%s", repositoryPath, repoPkgs.Project)
		rData := &repositoryData{
			Name: repoPkgs.Project,
			Path: myRepoPath,
			C:    cd,
		}
		d.RepositoryData = append(d.RepositoryData, rData)
		s.mux.HandleFunc(myRepoPath, func(w http.ResponseWriter, req *http.Request) {
			if err := s.renderer.Repository(w, rData); err != nil {
				log.Println(err)
			}
		})
		for _, pkgs := range repoPkgs.TaggedRepositoryPackages {
			myTagPath := fmt.Sprintf("%s/%s%s/%s", repositoryPath, repoPkgs.Project, tagPath, pkgs.Tag)
			tData := &tagData{
				Name:   pkgs.Tag,
				Path:   myTagPath,
				C:      cd,
				Parent: rData,
			}
			rData.TagData = append(rData.TagData, tData)
			s.mux.HandleFunc(myTagPath, func(w http.ResponseWriter, req *http.Request) {
				if err := s.renderer.Tag(w, tData); err != nil {
					log.Println(err)
				}
			})
			for _, pkg := range pkgs.RepositoryPackages {
				p := fmt.Sprintf("%s/%s%s/%s%s/%s", repositoryPath, repoPkgs.Project, tagPath, pkgs.Tag, packagePath, pkg.ImportPath)
				parts := strings.Split(pkg.ImportPath, "/")
				pData := &packageData{
					Name:    parts[len(parts)-1],
					Path:    p,
					Package: pkg.P,
					FSet:    pkg.F,
					C:       cd,
					Parent:  tData,
				}
				tData.PackageData = append(tData.PackageData, pData)
				s.mux.HandleFunc(p, func(w http.ResponseWriter, req *http.Request) {
					if err := s.renderer.Package(w, pData); err != nil {
						log.Println(err)
					}
				})
			}
		}
	}
	s.server.Handler = s.mux
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string, shutdownDone <-chan struct{}) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	errMux := make(chan error)
	pkgMux := make(chan projectRepositoryPackages)
	s.watchErrors(errMux, ctx, wg)
	s.watchPackages(pkgMux, ctx, wg)
	s.consumeErrors(errMux, ctx, wg)
	s.consumePackages(pkgMux, ctx, wg)
	s.syncLauncher(ctx, wg)
	err := s.server.ListenAndServeTLS(certFile, keyFile)
	cancelFunc()
	wg.Wait()
	<-shutdownDone
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}

func (s *Server) syncLauncher(c context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(s.refresh)
	s.sync()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				s.sync()
			case <-c.Done():
				return
			}
		}
	}()
}

func (s *Server) watchErrors(mux chan<- error, c context.Context, wg *sync.WaitGroup) {
	for _, repo := range s.repo {
		wg.Add(1)
		go func(r *repository) {
			defer wg.Done()
			ch := r.Errors()
			for {
				select {
				case e := <-ch:
					mux <- e
				case <-c.Done():
					return
				}
			}
		}(repo)
	}
}

type projectRepositoryPackages struct {
	ProjectName string
	Packages    TaggedRepositoryPackages
}

func (s *Server) watchPackages(mux chan<- projectRepositoryPackages, c context.Context, wg *sync.WaitGroup) {
	for name, repo := range s.repo {
		wg.Add(1)
		go func(projectName string, r *repository) {
			defer wg.Done()
			ch := r.Packages()
			for {
				select {
				case p := <-ch:
					mux <- projectRepositoryPackages{
						ProjectName: projectName,
						Packages:    p,
					}
				case <-c.Done():
					return
				}
			}
		}(name, repo)
	}
}

func (s *Server) sync() {
	log.Println("Beginning sync")
	var wg sync.WaitGroup
	for _, repo := range s.repo {
		wg.Add(1)
		go func(r *repository) {
			defer wg.Done()
			r.Sync()
		}(repo)
	}
	wg.Wait()
	log.Println("Done syncing")
}

func (s *Server) consumeErrors(mux <-chan error, c context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-mux:
				log.Println(e)
			case <-c.Done():
				return
			}
		}
	}()
}

func (s *Server) consumePackages(mux <-chan projectRepositoryPackages, c context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		m := make(ProjectPackages, 0, len(s.repo))
		for {
			select {
			case prp := <-mux:
				m = append(m, ProjectPackage{
					TaggedRepositoryPackages: prp.Packages,
					Project:                  prp.ProjectName,
				})
				if len(m) == len(s.repo) {
					m.Sort()
					s.createServeMux(m)
					m = make(ProjectPackages, 0, len(s.repo))
				}
			case <-c.Done():
				return
			}
		}
	}()
}
