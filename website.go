// +build ignore

package main

import (
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var DataDir = os.Getenv("DATA_DIR")
var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
var templatesDir = os.Getenv("TEMPLATES_DIR")

// the struct which contains the complete
// array of all attributes in the page file
type Pagexml struct {
	XMLName xml.Name `xml:"page"`
	Type    string   `xml:"type,attr"`
	Title   string   `xml:"title"`
	Date    string   `xml:"date"`
	Body    []byte   `xml:"body"`
}

// neuteredFileSystem is used to prevent directory listing of static assets
type neuteredFileSystem struct {
	fs http.FileSystem
}

type Common struct {
	Pages []string
	About []string
}

type Page struct {
	Pagexml [2]Pagexml
	Common  Common
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	// Check if path exists
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	// If path exists, check if is a file or a directory.
	// If is a directory, stop here with an error saying that file
	// does not exist. So user will get a 404 error code for a file/directory
	// that does not exist, and for directories that exist.
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		return nil, os.ErrNotExist
	}

	// If file exists and the path is not a directory, let's return the file
	return f, nil
}

func (pxml *Pagexml) savexml() error {
	filenamexml := pxml.Title + ".xml"
	filenamexml = filepath.Join(DataDir, filenamexml)

	t := time.Now()

	savev := &Pagexml{Type: "Normal", Title: pxml.Title, Date: t.Format("20060102150405"), Body: pxml.Body}

	output, err := xml.MarshalIndent(savev, "  ", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filenamexml, output, 0600)
}

//var CssStylesheet = "css/style.txt"
func loadPagexml(title string) (*Pagexml, error) {
	// Open our xmlFile
	filenamexml := title + ".xml"
	filenamexml = filepath.Join(DataDir, filenamexml)
	xmlFile, err := os.Open(filenamexml)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our Pagexml
	var pagexmlv Pagexml
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &pagexmlv)

	return &Pagexml{Title: pagexmlv.Title, Body: pagexmlv.Body}, nil
	//return &Page{Title: title, Body: body, Css: css}, nil
}

func getPages() ([]string, error) {
	var files []string

	err := filepath.Walk(DataDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		//if filepath.Ext(path) != ".xml" {
		//	return nil
		//}
		r, _ := regexp.Compile(".xml")
		if !r.MatchString(info.Name()) {
			return nil
		}
		filename := r.ReplaceAllString(info.Name(), "")
		files = append(files, filename)
		return nil
	})
	if err != nil {
		panic(err)
	}

	return files, nil
}

func getCommon() (*Common, error) {
	var about []string
	pages, err := getPages()
	if err != nil {
		//
	}
	about = append(about, "This is a sentence about me.")
	return &Common{Pages: pages, About: about}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	var p *Page
	p = new(Page)
	pxml, err := loadPagexml(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	pcom, _ := getCommon()
	(*p).Pagexml[0] = *pxml
	(*p).Common = *pcom
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	var p *Page
	p = new(Page)
	pxml, err := loadPagexml(title)
	if err != nil {
		pxml = &Pagexml{Title: title}
	}
	pcom, err := getCommon()
	(*p).Pagexml[0] = *pxml
	(*p).Common = *pcom
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	pxml := &Pagexml{Title: title, Body: []byte(body)}
	err := pxml.savexml()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

//var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	pattern := filepath.Join(templatesDir, "*.tmpl")
	var templates = template.Must(template.ParseGlob(pattern))
	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	//http.HandleFunc("", makeHandler(viewHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	// Serve static files while preventing directory listing
	//mux := http.NewServeMux()
	fs := http.FileServer(neuteredFileSystem{http.Dir(staticAssetsDir)})
	//mux.Handle("/", fs)
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
