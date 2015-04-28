package gospell

import (
	"net/http"
	"github.com/tonyshaw/GoSpell/engine"
	"appengine"
	"log"
	"text/template"
)

const itemNum = 5
const artKind = "article"
var templ *template.Template

type listStruct struct {
	Selected int //0home, 1project, 2 resume, 3 none
	Contents []engine.PageStruct
}

func init(){
	http.HandleFunc("/admin/upload",upload)
	http.HandleFunc("/admin/uploadit",uploadit)
	http.HandleFunc("/home", home)
	http.HandleFunc("/article", article)
	http.HandleFunc("/", home)
	var err error
	if templ, err = template.ParseFiles("template.html"); err != nil {
		log.Println("Template Error ",err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	var data listStruct
	data.Selected = 0
	data.Contents = engine.Query(artKind, 0, itemNum, appengine.NewContext(r))
	log.Println("---> ", templ.ExecuteTemplate(w, "HomePage", data))
}

func article(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	templ.ExecuteTemplate(w, "Article", engine.Get(key, appengine.NewContext(r)))
}

func upload(w http.ResponseWriter, r *http.Request) {
	templ.ExecuteTemplate(w,"Upload","")
}
func uploadit(w http.ResponseWriter, r *http.Request) {
	var item engine.PageStruct
	item.Title = r.FormValue("title")
	item.Content = r.FormValue("content")
	item.Kind = artKind;
	engine.Save(&item, appengine.NewContext(r))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

