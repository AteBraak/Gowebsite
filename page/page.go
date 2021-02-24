package page

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

var DataDir = os.Getenv("DATA_DIR")

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

type Page struct {
	Pagexml []Pagexml
	Common  Common
	User    User
}

type Common struct {
	Pages []string
	About []string
}

type PageListing struct {
	Title string
	Date  time.Time
}

var PageList = []PageListing{}

type User struct {
	Id     int
	Name   string
	Access []string
	Folder string
}

var Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func GetUser(s *sessions.Session) (*User, error) {
	val := s.Values["user"]
	var user = User{}
	user, ok := val.(User)
	if !ok {
		return &User{Id: -1}, nil
	}
	return &user, nil
}

func (pxml *Pagexml) Savexml() error {
	var savev *Pagexml
	var newfile bool
	filenamexml := pxml.Title + ".xml"
	filenamexml = filepath.Join(DataDir, filenamexml)
	err := GetLatestPages()
	if err != nil {
		panic(err)
	}
	newfile = true
	for i := 0; i < len(PageList); i++ {
		if pxml.Title == PageList[i].Title {
			newfile = false
		}
	}

	t := time.Now()
	if newfile {
		savev = &Pagexml{Type: "Normal", Title: pxml.Title, Date: t, DateModified: t, Body: pxml.Body}
	} else {
		oldpxml, _ := LoadPagexml(pxml.Title)
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

func LoadPagexml(title string) (*Pagexml, error) {
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

//Gets list of pages
//function excludes the excludepages variable
func GetPages() ([]string, error) {
	var files []string

	//fileext := `xml`
	var regexMatch strings.Builder
	var regexRemove strings.Builder
	// Note for regex ?! does not work in golang
	regexMatch.WriteString(`((^|\/)(\w*))\.xml`)
	regexRemove.WriteString(`\.xml`)

	err := filepath.Walk(DataDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		//if filepath.Ext(path) != ".xml" {
		//	return nil
		//}
		rmatch, _ := regexp.Compile(regexMatch.String())
		rremove, _ := regexp.Compile(regexRemove.String())
		if !rmatch.MatchString(info.Name()) {
			return nil
		}
		filename := rremove.ReplaceAllString(info.Name(), "")
		files = append(files, filename)
		return nil
	})
	if err != nil {
		panic(err)
	}

	return files, nil
}

func GetCommon() (*Common, error) {
	var about []string
	pages, err := GetPages()
	if err != nil {
		//
	}
	pxml, err := LoadPagexml("-about")
	if err != nil {
		panic(err)
	}

	about = append(about, string(pxml.Body))
	return &Common{Pages: pages, About: about}, nil
}

//this updates the variable PageList with the list of pages
// to do: sort sent list based on latest to oldest
func GetLatestPages() error {
	pages, err := GetPages()
	if err != nil {
		panic(err)
	}
	PageList = nil
	for i := 0; i < len(pages); i++ {
		pxml, err := LoadPagexml(pages[i])
		if err != nil {
			panic(err)
		}
		PageList = append(PageList, PageListing{Title: pages[i], Date: pxml.Date})
	}
	return nil
}

func GetPagedata(r *http.Request, title []string) (*Page, error) {
	err := GetLatestPages()
	if err != nil {
		panic(err)
	}
	var p *Page
	p = new(Page)
	var pxml *Pagexml
	for i := 0; i < len(title); i++ {
		pxml, err = LoadPagexml(title[i])
		//(*p).Pagexml[i] = *pxml
		(*p).Pagexml = append((*p).Pagexml, *pxml)
		if err != nil {
			return nil, err
		}
	}

	pcom, _ := GetCommon()
	if err != nil {
		return nil, err
	}

	session, err := Store.Get(r, "session-name")
	if err != nil {
		return nil, err
	}
	puser, err := GetUser(session)

	(*p).Common = *pcom
	(*p).User = *puser
	return p, nil
}
