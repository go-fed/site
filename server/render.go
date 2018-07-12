package server

import (
	"bytes"
	"go/doc"
	"go/printer"
	"go/token"
	"html/template"
	"io"
	"strings"
	"time"
)

const (
	notFoundTemplate     = "notFound"
	homeTemplate         = "home"
	tutorialTemplate     = "tutorial"
	repositoryTemplate   = "repository"
	repositoriesTemplate = "repositories"
	tagTemplate          = "tag"
	packageTemplate      = "package"
)

type data struct {
	HomeData         *homeData
	TutorialData     *tutorialData
	RepositoriesData *repositoriesData
	RepositoryData   []*repositoryData
}

type repositoriesData struct {
	Path string
	C    *commonData
}

type commonData struct {
	PageTitle   string
	RefreshTime time.Time
	Data        *data
}

type homeData struct {
	Path string
	C    *commonData
}

type tutorialData struct {
	Path string
	C    *commonData
}

type repositoryData struct {
	Name    string
	Path    string
	TagData []*tagData
	C       *commonData
}

type tagData struct {
	Name        string
	Path        string
	PackageData []*packageData
	C           *commonData
	Parent      *repositoryData
}

type packageData struct {
	Name    string
	Path    string
	Package *doc.Package
	FSet    *token.FileSet
	C       *commonData
	Parent  *tagData
}

type renderer struct {
	paths []string
	t     *template.Template
}

func newRenderer(paths []string) (*renderer, error) {
	r := &renderer{
		paths: paths,
	}
	err := r.parseAllTemplates()
	return r, err
}

func (r *renderer) parseAllTemplates() error {
	var err error
	r.t, err = template.New("").Funcs(r.funcMap()).ParseFiles(r.paths...)
	return err
}

func (r *renderer) funcMap() template.FuncMap {
	return template.FuncMap{
		"ToHTML": func(s string) template.HTML {
			var b bytes.Buffer
			doc.ToHTML(&b, s, nil)
			return template.HTML(b.String())
		},
		"PrintAST": func(node interface{}, fset *token.FileSet) template.HTML {
			var b bytes.Buffer
			printer.Fprint(&b, fset, node)
			toAddComments := strings.Split(b.String(), "\n")
			var commented bytes.Buffer
			for i, elem := range toAddComments {
				if strings.HasPrefix(strings.TrimSpace(elem), "//") {
					elem = "<span class=\"comment\">" + elem + "</span>"
				}
				if i < len(toAddComments)-1 {
					elem += "\n"
				}
				commented.WriteString(elem)
			}
			return template.HTML(commented.String())
		},
	}
}

func (r *renderer) execute(w io.Writer, name string, data interface{}) error {
	return r.t.ExecuteTemplate(w, name, data)
}

func (r *renderer) NotFound(w io.Writer, data *commonData) error {
	return r.execute(w, notFoundTemplate, data)
}

func (r *renderer) Home(w io.Writer, data *homeData) error {
	return r.execute(w, homeTemplate, data)
}

func (r *renderer) Tutorial(w io.Writer, data *tutorialData) error {
	return r.execute(w, tutorialTemplate, data)
}

func (r *renderer) AllRepositories(w io.Writer, data *repositoriesData) error {
	return r.execute(w, repositoriesTemplate, data)
}

func (r *renderer) Repository(w io.Writer, data *repositoryData) error {
	return r.execute(w, repositoryTemplate, data)
}

func (r *renderer) Tag(w io.Writer, data *tagData) error {
	return r.execute(w, tagTemplate, data)
}

func (r *renderer) Package(w io.Writer, data *packageData) error {
	return r.execute(w, packageTemplate, data)
}
