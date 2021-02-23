module main

go 1.15

replace Gowebsite/page => /page

replace Gowebsite/routes => /routes

replace Gowebsite/encryption => /encryption

replace Gowebsite/databaseserver => /databaseserver

require (
	Gowebsite/databaseserver v0.0.0-00010101000000-000000000000
	Gowebsite/encryption v0.0.0-00010101000000-000000000000 // indirect
	Gowebsite/page v0.0.0-00010101000000-000000000000
	Gowebsite/routes v0.0.0-00010101000000-000000000000
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
)
