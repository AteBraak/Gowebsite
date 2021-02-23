package routes

import (
	"Gowebsite/databaseserver"
	"Gowebsite/page"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var templatesDir = os.Getenv("TEMPLATES_DIR")

func HomeHandler(w http.ResponseWriter, r *http.Request, title string) {
	var titlelist []string
	//len(page.PageList) < 1 {
	if len(page.PageList) < 1 {
		err := page.GetLatestPages()
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < len(page.PageList); i++ {
		//titlelist[i] = page.PageList[i].Title
		titlelist = append(titlelist, page.PageList[i].Title)
		if i == 1 {
			break
		}
	}
	p, err := page.GetPagedata(r, titlelist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "home", p)
}

func ViewHandler(w http.ResponseWriter, r *http.Request, title string) {
	var p *page.Page
	p = new(page.Page)
	pxml, err := page.LoadPagexml(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	pcom, _ := page.GetCommon()

	session, err := page.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	puser, err := page.GetUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	(*p).Pagexml[0] = *pxml
	(*p).Common = *pcom
	(*p).User = *puser

	renderTemplate(w, "view", p)
}

func EditHandler(w http.ResponseWriter, r *http.Request, title string) {
	// block to handle non editors trying to edit
	session, err := page.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	puser, err := page.GetUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	editor := false

	for _, access := range puser.Access {
		if access == "admin" {
			editor = true
		}
	}
	if !editor {
		http.Redirect(w, r, "/view/"+title, http.StatusFound)
		return
	}
	// -end- block to handle non editors trying to edit

	var p *page.Page
	p = new(page.Page)
	pxml, err := page.LoadPagexml(title)
	if err != nil {
		pxml = &page.Pagexml{Title: title}
	}
	pcom, err := page.GetCommon()
	(*p).Pagexml[0] = *pxml
	(*p).Common = *pcom
	(*p).User = *puser
	renderTemplate(w, "edit", p)
}

func SaveHandler(w http.ResponseWriter, r *http.Request, title string) {
	// block to handle non editors trying to save
	session, err := page.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	puser, err := page.GetUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	editor := false

	for _, access := range puser.Access {
		if access == "admin" {
			editor = true
		}
	}
	if !editor {
		http.Redirect(w, r, "/view/"+title, http.StatusFound)
		return
	}
	// -end- block to handle non editors trying to save

	body := r.FormValue("body")
	pxml := &page.Pagexml{Title: title, Body: []byte(body)}
	err = pxml.Savexml()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func SignHandler(w http.ResponseWriter, r *http.Request, title string) {
	var p *page.Page
	p = new(page.Page)
	pcom, _ := page.GetCommon()
	(*p).Common = *pcom

	//Signin->Login or  Signup->Adduser

	if title == "signin" {
		renderTemplate(w, "signin", p)
	} else if title == "signup" {
		renderTemplate(w, "signup", p)
	} else if title == "login" {
		password := r.FormValue("password")
		username := r.FormValue("username")
		pass, err := databaseserver.CheckUserPassword(username, &password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if pass {
			session, _ := page.Store.Get(r, "session-name")
			// Set some session values.

			userdata, _ := databaseserver.GetUser(username)
			user := &page.User{
				Id:     userdata.Userid,
				Name:   userdata.Username,
				Access: userdata.Access,
				Folder: userdata.Username,
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
		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")
		err := databaseserver.NewUser(username, &password, email, []string{"user"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session, _ := page.Store.Get(r, "session-name")
		userdata, _ := databaseserver.GetUser(username)
		user := &page.User{
			Id:     userdata.Userid,
			Name:   userdata.Username,
			Access: userdata.Access,
			Folder: userdata.Username,
		}
		session.Values["user"] = user
		// Save it before we write to the response/return from the handler.
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	} else if title == "signout" {
		session, err := page.Store.Get(r, "session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["user"] = page.User{}
		session.Options.MaxAge = -1
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/home", http.StatusFound)
	}
}

func AdminHandler(w http.ResponseWriter, r *http.Request, title string) {

	// block to handle non editors trying to edit
	session, err := page.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	puser, err := page.GetUser(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	editor := false

	for _, access := range puser.Access {
		if access == "admin" {
			editor = true
		}
	}
	if !editor {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// -end- block to handle non editors trying to edit

	var titlelist []string
	//len(page.PageList) < 1 {
	if len(page.PageList) < 1 {
		err := page.GetLatestPages()
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < len(page.PageList); i++ {
		//titlelist[i] = page.PageList[i].Title
		titlelist = append(titlelist, page.PageList[i].Title)
		if i == 1 {
			break
		}
	}
	p, err := page.GetPagedata(r, titlelist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "admin", p)
}

//var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *page.Page) {
	pattern := filepath.Join(templatesDir, "*.tmpl")
	var templates = template.Must(template.ParseGlob(pattern))
	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var validPath = regexp.MustCompile("^(/(edit|save|sign|view)/([a-zA-Z0-9]+)$|/(admi|home)$)")

func MakeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[3])
	}
}
