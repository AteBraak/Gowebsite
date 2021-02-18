// +build ignore

package main

import (
	"encoding/gob"
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var DataDir = os.Getenv("DATA_DIR")
var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")
var templatesDir = os.Getenv("TEMPLATES_DIR")
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

// the struct which contains the complete
// array of all attributes in the page file
type Pagexml struct {
	XMLName      xml.Name  `xml:"page"`
	Type         string    `xml:"type,attr"`
	Title        string    `xml:"title"`
	Date         time.Time `xml:"date"`
	DateModified time.Time `xml:"datemodified"`
	Body         []byte    `xml:"body"`
}

// neuteredFileSystem is used to prevent directory listing of static assets
type neuteredFileSystem struct {
	fs http.FileSystem
}

type Common struct {
	Pages []string
	About []string
}

type User struct {
	Id     int
	Name   string
	Access []string
	Folder string
}

type Page struct {
	Pagexml [2]Pagexml
	Common  Common
	User    User
}

type PageList struct {
	Title string
	Date  time.Time
}

var pageList = []PageList{}

var userhash string

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
	var savev *Pagexml
	var newfile bool
	filenamexml := pxml.Title + ".xml"
	filenamexml = filepath.Join(DataDir, filenamexml)
	err := getLatestPages()
	if err != nil {
		panic(err)
	}
	newfile = true
	for i := 0; i < len(pageList); i++ {
		if pxml.Title == pageList[i].Title {
			newfile = false
		}
	}

	t := time.Now()
	if newfile {
		savev = &Pagexml{Type: "Normal", Title: pxml.Title, Date: t, DateModified: t, Body: pxml.Body}
	} else {
		oldpxml, _ := loadPagexml(pxml.Title)
		oldpxml.Body = pxml.Body
		oldpxml.DateModified = t
		savev = oldpxml
	}

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

	return &pagexmlv, nil
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

func getLatestPages() error {
	pages, err := getPages()
	if err != nil {
		panic(err)
	}
	pageList = nil
	for i := 0; i < len(pages); i++ {
		pxml, err := loadPagexml(pages[i])
		if err != nil {
			panic(err)
		}
		pageList = append(pageList, PageList{Title: pages[i], Date: pxml.Date})
	}
	return nil
}

func getUser(s *sessions.Session) (*User, error) {
	val := s.Values["user"]
	var user = User{}
	user, ok := val.(User)
	if !ok {
		return &User{Id: -1}, nil
	}
	return &user, nil
}

func getPagedata(r *http.Request, title []string) (*Page, error) {
	err := getLatestPages()
	if err != nil {
		panic(err)
	}
	var p *Page
	p = new(Page)
	var pxml *Pagexml
	for i := 0; i < len(title); i++ {
		pxml, err = loadPagexml(title[i])
		(*p).Pagexml[i] = *pxml
		if err != nil {
			return nil, err
		}
	}

	pcom, _ := getCommon()
	if err != nil {
		return nil, err
	}

	session, err := store.Get(r, "session-name")
	if err != nil {
		return nil, err
	}
	puser, err := getUser(session)

	(*p).Common = *pcom
	(*p).User = *puser
	return p, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func homeHandler(w http.ResponseWriter, r *http.Request, title string) {
	var titlelist []string
	//len(pageList) < 1 {
	if len(pageList) < 1 {
		err := getLatestPages()
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < len(pageList); i++ {
		//titlelist[i] = pageList[i].Title
		titlelist = append(titlelist, pageList[i].Title)
		if i == 1 {
			break
		}
	}
	p, err := getPagedata(r, titlelist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "home", p)
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

	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	puser, err := getUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	(*p).Pagexml[0] = *pxml
	(*p).Common = *pcom
	(*p).User = *puser

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

func signHandler(w http.ResponseWriter, r *http.Request, title string) {
	var p *Page
	p = new(Page)
	pcom, _ := getCommon()
	(*p).Common = *pcom

	//Signin->Login or  Signup->Adduser

	if title == "signin" {
		renderTemplate(w, "signin", p)
	} else if title == "signup" {
		renderTemplate(w, "signup", p)
	} else if title == "login" {
		//username := r.FormValue("username")
		password := r.FormValue("password")
		username := r.FormValue("username")
		if CheckPasswordHash(password, userhash) {
			session, _ := store.Get(r, "session-name")
			// Set some session values.
			user := &User{
				Id:     108,
				Name:   username,
				Access: []string{"admin"},
				Folder: " ",
			}
			session.Values["user"] = user
			// Save it before we write to the response/return from the handler.
			err := session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "password does not match", http.StatusInternalServerError)
		}
		http.Redirect(w, r, "/home", http.StatusFound)
	} else if title == "adduser" {
		//username := r.FormValue("username")
		password := r.FormValue("password")
		hash, err := HashPassword(password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userhash = hash
		http.Redirect(w, r, "/home", http.StatusFound)
	} else if title == "signout" {
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["user"] = User{}
		session.Options.MaxAge = -1
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/home", http.StatusFound)
	}
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

//var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var validPath = regexp.MustCompile("^(/(edit|save|sign|view)/([a-zA-Z0-9]+)$|/home$)")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[3])
	}
}

func init() {
	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	store = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	gob.Register(User{})

}

func main() {
	http.HandleFunc("/home", makeHandler(homeHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/sign/", makeHandler(signHandler))

	// Serve static files while preventing directory listing
	//mux := http.NewServeMux()
	fs := http.FileServer(neuteredFileSystem{http.Dir(staticAssetsDir)})
	//mux.Handle("/", fs)
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
