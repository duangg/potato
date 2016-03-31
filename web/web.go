package web

import (
	"html/template"
	"net/http"

	"gopkg.in/fsnotify.v1"
	"github.com/SpruceX/potato/api"
	"github.com/SpruceX/potato/utils"
)

var Templates *template.Template

type HtmlTemplatePage api.Page

func NewHtmlTemplatePage(templateName, title string) *HtmlTemplatePage {
	if len(title) > 0 {
		title = utils.Cfg.ServiceSettings.SiteName + " - " + title
	}
	props := make(map[string]string)
	return &HtmlTemplatePage{TemplateName: templateName, Title: title, SiteName: utils.Cfg.ServiceSettings.SiteName, Props: props}
}

func InitWeb() {
	log.Printf("Initializing web routes")
	mainRoute := api.Srv.Router
	log.Printf("Using static direcotry at %v", "web/static")
	mainRoute.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mainRoute.Handle("/", api.AppHandler(dashboard))
	mainRoute.Handle("/backup/dashboard", api.AppHandler(dashboard))
	mainRoute.Handle("/backup/backups", api.AppHandler(backups))
	mainRoute.Handle("/backup/crons", api.AppHandler(crons))
	mainRoute.Handle("/backup/servers", api.AppHandler(servers))
	mainRoute.Handle("/backup/topology", api.AppHandler(topology))
	watchAndParseTemplates()
}

func mkSlice(items ...interface{}) []interface{} {
	return items
}

var funcMap = template.FuncMap{"mkslice": mkSlice}

func watchAndParseTemplates() {

	templatesDir := utils.FindDir("web/templates/")
	log.Printf("Parsing templates at %v", templatesDir)
	var err error
	if Templates, err = template.New("templates").Funcs(funcMap).ParseGlob(templatesDir + "*.html"); err != nil {
		log.Printf("Failed to parse templates %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create directory watcher %v", err)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("Re-parsing templates because of modified file %v", event.Name)
					if Templates, err = template.New("templates").Funcs(funcMap).ParseGlob(templatesDir + "*.html"); err != nil {
						log.Printf("Failed to parse templates %v", err)
					} else {
						Templates = Templates.Funcs(funcMap)
					}
				}
			case err := <-watcher.Errors:
				log.Printf("Failed in directory watcher %v", err)
			}
		}
	}()

	err = watcher.Add(templatesDir)
	if err != nil {
		log.Printf("Failed to add directory to watcher %v", err)
	}
}

func (t *HtmlTemplatePage) Render(c *api.Context, w http.ResponseWriter) {
	if err := Templates.ExecuteTemplate(w, t.TemplateName, t); err != nil {
		c.SetUnknownError(t.TemplateName, err.Error())
	}
}

func backups(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("home", "Home")
	p.Render(c, w)
}

func login(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("login", "Login")
	p.Render(c, w)
}

func crons(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("crons", "crons")
	p.Render(c, w)
}

func servers(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("servers", "servers")
	p.Render(c, w)
}

func storage(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("storage", "storage")
	p.Render(c, w)
}
func admin(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("admin", "Admin")
	p.Render(c, w)
}

func dashboard(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("dashboard", "Dashboard")
	p.Render(c, w)
}

func topology(c *api.Context, w http.ResponseWriter, r *http.Request) {
	p := NewHtmlTemplatePage("topology", "Topology")
	p.Render(c, w)
}
