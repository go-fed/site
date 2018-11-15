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
	faviconPath    = "/favicon"
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

type FaviconOptions struct {
	Png16Path  string
	Png32Path  string
	Png48Path  string
	Png96Path  string
	Png192Path string
}

type ServerOptions struct {
	TemplateFiles []string
	Repositories  []RepositoryOptions
	HttpServer    *http.Server
	RefreshRate   time.Duration
	Favicon       FaviconOptions
	SiteTitle     string
	OrgDataPath   string
	OrgDataName   string
}

type Server struct {
	repo        map[string]*repository
	favicon16   []byte
	favicon32   []byte
	favicon48   []byte
	favicon96   []byte
	favicon192  []byte
	siteTitle   string
	orgDataPath string
	orgDataName string
	renderer    *renderer
	mux         *http.ServeMux
	server      *http.Server
	refresh     time.Duration
}

func NewServer(opts ServerOptions) (*Server, error) {
	renderer, err := newRenderer(opts.TemplateFiles)
	if err != nil {
		return nil, err
	}
	s := &Server{
		repo:        make(map[string]*repository),
		siteTitle:   opts.SiteTitle,
		orgDataPath: opts.OrgDataPath,
		orgDataName: opts.OrgDataName,
		renderer:    renderer,
		server:      opts.HttpServer,
		refresh:     opts.RefreshRate,
	}
	if len(opts.Favicon.Png16Path) > 0 {
		favicon, err := ioutil.ReadFile(opts.Favicon.Png16Path)
		if err != nil {
			return nil, err
		}
		s.favicon16 = favicon
	}
	if len(opts.Favicon.Png32Path) > 0 {
		favicon, err := ioutil.ReadFile(opts.Favicon.Png32Path)
		if err != nil {
			return nil, err
		}
		s.favicon32 = favicon
	}
	if len(opts.Favicon.Png48Path) > 0 {
		favicon, err := ioutil.ReadFile(opts.Favicon.Png48Path)
		if err != nil {
			return nil, err
		}
		s.favicon48 = favicon
	}
	if len(opts.Favicon.Png96Path) > 0 {
		favicon, err := ioutil.ReadFile(opts.Favicon.Png96Path)
		if err != nil {
			return nil, err
		}
		s.favicon96 = favicon
	}
	if len(opts.Favicon.Png192Path) > 0 {
		favicon, err := ioutil.ReadFile(opts.Favicon.Png192Path)
		if err != nil {
			return nil, err
		}
		s.favicon192 = favicon
	}
	for _, repo := range opts.Repositories {
		s.repo[repo.ProjectName] = newRepository(repo.HttpsCloneURL, repo.DiskCacheFilePath, repo.GitOperationTimeout)
	}
	return s, nil
}

func (s *Server) createServeMux(projectPkgs ProjectPackages) {
	s.mux = http.NewServeMux()
	var favicons []*faviconData
	if len(s.favicon16) > 0 {
		path := faviconPath + "-16.png"
		s.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			b := bytes.NewReader(s.favicon16)
			http.ServeContent(w, req, "image/png", time.Now(), b)
		})
		favicons = append(favicons, &faviconData{
			Path: path,
			Size: "16x16",
		})
	}
	if len(s.favicon48) > 0 {
		path := faviconPath + "-48.png"
		s.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			b := bytes.NewReader(s.favicon48)
			http.ServeContent(w, req, "image/png", time.Now(), b)
		})
		favicons = append(favicons, &faviconData{
			Path: path,
			Size: "48x48",
		})
	}
	if len(s.favicon96) > 0 {
		path := faviconPath + "-96.png"
		s.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			b := bytes.NewReader(s.favicon96)
			http.ServeContent(w, req, "image/png", time.Now(), b)
		})
		favicons = append(favicons, &faviconData{
			Path: path,
			Size: "96x96",
		})
	}
	if len(s.favicon192) > 0 {
		path := faviconPath + "-192.png"
		s.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			b := bytes.NewReader(s.favicon192)
			http.ServeContent(w, req, "image/png", time.Now(), b)
		})
		favicons = append(favicons, &faviconData{
			Path: path,
			Size: "192x192",
		})
	}
	if len(s.favicon32) > 0 {
		path := faviconPath + "-32.png"
		s.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			b := bytes.NewReader(s.favicon32)
			http.ServeContent(w, req, "image/png", time.Now(), b)
		})
		favicons = append(favicons, &faviconData{
			Path: path,
			Size: "32x32",
		})
	}
	cd := &commonData{
		Favicons:    favicons,
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
	if len(s.orgDataPath) > 0 && len(s.orgDataName) > 0 {
		d.OrgData = &orgData{
			Path: s.orgDataPath,
			Name: s.orgDataName,
		}
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
	if projectPkgs == nil {
		s.mux.HandleFunc(repositoryPath, func(w http.ResponseWriter, req *http.Request) {
			if err := s.renderer.Refreshing(w, d.HomeData); err != nil {
				log.Println(err)
			}
		})
	} else {
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
	}
	s.server.Handler = s.mux
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string, shutdownDone <-chan struct{}) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	errMux := make(chan error)
	pkgMux := make(chan projectRepositoryPackages)
	clearMux := make(chan string)
	s.watchErrors(errMux, ctx, wg)
	s.watchPackages(pkgMux, clearMux, ctx, wg)
	s.consumeErrors(errMux, ctx, wg)
	s.consumePackages(pkgMux, clearMux, ctx, wg)
	s.syncLauncher(ctx, wg)
	err := s.server.ListenAndServeTLS(certFile, keyFile)
	log.Println("Server ListenAndServeTLS Done")
	cancelFunc()
	log.Println("Cancelled context")
	wg.Wait()
	log.Println("Server WaitGroups done")
	<-shutdownDone
	log.Println("Received shutdownDone")
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

func (s *Server) watchPackages(mux chan<- projectRepositoryPackages, clear chan<- string, c context.Context, wg *sync.WaitGroup) {
	for name, repo := range s.repo {
		wg.Add(1)
		go func(projectName string, r *repository) {
			defer wg.Done()
			begin := r.BeginSync()
			ch := r.Packages()
			for {
				select {
				case <-begin:
					clear <- projectName
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

func (s *Server) consumePackages(mux <-chan projectRepositoryPackages, clear <-chan string, c context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		m := make(ProjectPackages, 0, len(s.repo))
		for {
			select {
			case <-clear:
				s.createServeMux(nil)
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
