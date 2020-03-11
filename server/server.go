package server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	homePath          = "/"
	gofedTutorialPath = "/overview"
	tutorialPath      = "/activitypub-tutorial"
	asRefPath         = "/ref/activity/streams"
	apRefPath         = "/ref/activity/pub"
	httpsigRefPath    = "/ref/httpsig"
	apcoreRefPath     = "/ref/apcore"
	glancePath        = "/activitypub-glance"
	faviconPath       = "/favicon"
)

type FaviconOptions struct {
	Png16Path  string
	Png32Path  string
	Png48Path  string
	Png96Path  string
	Png192Path string
}

type ServerOptions struct {
	TemplateFiles []string
	HttpServer    *http.Server
	RefreshRate   time.Duration
	Favicon       FaviconOptions
	SiteTitle     string
	OrgDataPath   string
	OrgDataName   string
}

type Server struct {
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
	s.createServeMux()
	return s, nil
}

func (s *Server) createServeMux() {
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
		GoFedTutorialData: &tutorialData{
			Path: gofedTutorialPath,
			C:    cd,
		},
		ActivityTutorialData: &tutorialData{
			Path: tutorialPath,
			C:    cd,
		},
		ActivityStreamsData: &referenceData{
			Path: asRefPath,
			C:    cd,
		},
		ActivityPubData: &referenceData{
			Path: apRefPath,
			C:    cd,
		},
		HttpSigData: &referenceData{
			Path: httpsigRefPath,
			C:    cd,
		},
		ApCoreData: &referenceData{
			Path: apcoreRefPath,
			C:    cd,
		},
		ActivityPubGlanceData: &tutorialData{
			Path: glancePath,
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

	s.mux.HandleFunc(gofedTutorialPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.GoFedTutorial(w, d.GoFedTutorialData); err != nil {
			log.Println(err)
		}
	})
	s.mux.HandleFunc(tutorialPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.ActivityTutorial(w, d.ActivityTutorialData); err != nil {
			log.Println(err)
		}
	})
	s.mux.HandleFunc(asRefPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.ActivityStreamsReference(w, d.ActivityStreamsData); err != nil {
			log.Println(err)
		}
	})
	s.mux.HandleFunc(apRefPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.ActivityPubReference(w, d.ActivityPubData); err != nil {
			log.Println(err)
		}
	})
	s.mux.HandleFunc(httpsigRefPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.HttpSigReference(w, d.HttpSigData); err != nil {
			log.Println(err)
		}
	})
	s.mux.HandleFunc(apcoreRefPath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.ApCoreReference(w, d.ApCoreData); err != nil {
			log.Println(err)
		}
	})
	s.mux.HandleFunc(glancePath, func(w http.ResponseWriter, req *http.Request) {
		if err := s.renderer.ActivityGlanceTutorial(w, d.ActivityPubGlanceData); err != nil {
			log.Println(err)
		}
	})
	s.server.Handler = s.mux
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	err := s.server.ListenAndServeTLS(certFile, keyFile)
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}
