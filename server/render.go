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
	notFoundTemplate         = "notFound"
	homeTemplate             = "home"
	goFedTutorialTemplate    = "gofedtutorial"
	activityTutorialTemplate = "activitytutorial"
	activityStreamsTemplate  = "activitystreamsref"
	activityPubTemplate      = "activitypubref"
	httpSigsTemplate         = "httpsigsref"
	apcoreTemplate           = "apcoreref"
	activityGlanceTemplate   = "activityglance"
)

type data struct {
	HomeData              *homeData
	GoFedTutorialData     *tutorialData
	ActivityTutorialData  *tutorialData
	ActivityStreamsData   *referenceData
	ActivityPubData       *referenceData
	HttpSigData           *referenceData
	ApCoreData            *referenceData
	ActivityPubGlanceData *tutorialData
}

type commonData struct {
	Favicons    []*faviconData
	PageTitle   string
	RefreshTime time.Time
	Data        *data
}

type faviconData struct {
	Path string
	Size string
}

type homeData struct {
	Path string
	C    *commonData
}

type tutorialData struct {
	Path string
	C    *commonData
}

type referenceData struct {
	Path string
	C    *commonData
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

func (r *renderer) GoFedTutorial(w io.Writer, data *tutorialData) error {
	return r.execute(w, goFedTutorialTemplate, data)
}

func (r *renderer) ActivityTutorial(w io.Writer, data *tutorialData) error {
	return r.execute(w, activityTutorialTemplate, data)
}

func (r *renderer) ActivityStreamsReference(w io.Writer, data *referenceData) error {
	return r.execute(w, activityStreamsTemplate, data)
}

func (r *renderer) ActivityPubReference(w io.Writer, data *referenceData) error {
	return r.execute(w, activityPubTemplate, data)
}

func (r *renderer) HttpSigReference(w io.Writer, data *referenceData) error {
	return r.execute(w, httpSigsTemplate, data)
}

func (r *renderer) ApCoreReference(w io.Writer, data *referenceData) error {
	return r.execute(w, apcoreTemplate, data)
}

func (r *renderer) ActivityGlanceTutorial(w io.Writer, data *tutorialData) error {
	return r.execute(w, activityGlanceTemplate, data)
}
