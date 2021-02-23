package main

import (
	"Gowebsite/databaseserver"
	"Gowebsite/page"
	"Gowebsite/routes"
	"encoding/gob"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var staticAssetsDir = os.Getenv("STATIC_ASSETS_DIR")

// neuteredFileSystem is used to prevent directory listing of static assets
type neuteredFileSystem struct {
	fs http.FileSystem
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

func init() {
	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	page.Store = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	page.Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	gob.Register(page.User{})

	var userDatabasename = "credentials"
	var userDatabasePassword = "password5"

	//to set userdatabase name and password
	err := databaseserver.SetUserDatabase(userDatabasename, userDatabasePassword)
	if err != nil {
		return
	}

	//uncomment -create user database - below to initialize db
	//err = databaseserver.CreateUserDatabase()
	//if err != nil {
	//	return
	//}
	//end of section -create user database -

}

func main() {
	http.HandleFunc("/home", routes.MakeHandler(routes.HomeHandler))
	http.HandleFunc("/admi", routes.MakeHandler(routes.AdminHandler))
	http.HandleFunc("/view/", routes.MakeHandler(routes.ViewHandler))
	http.HandleFunc("/edit/", routes.MakeHandler(routes.EditHandler))
	http.HandleFunc("/save/", routes.MakeHandler(routes.SaveHandler))
	http.HandleFunc("/sign/", routes.MakeHandler(routes.SignHandler))

	// Serve static files while preventing directory listing
	//mux := http.NewServeMux()
	fs := http.FileServer(neuteredFileSystem{http.Dir(staticAssetsDir)})
	//mux.Handle("/", fs)
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
