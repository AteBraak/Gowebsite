package databaseserver

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var userDatabasename = "credentials"
var userDatabasePassword = "Password2"

var password = "Password"
var password2 = "Password2"
var DataDir = os.Getenv("DATA_DIR")

type Userdata struct {
	Type         string   `xml:"type,attr"`
	Username     string   `xml:"username"`
	Userid       int      `xml:"userid"`
	PasswordHash string   `xml:"hash"`
	Email        string   `xml:"email"`
	Access       []string `xml:"access"`
}

type DataMask struct {
	Type         bool
	Username     bool
	Userid       bool
	PasswordHash bool
	Email        bool
	Access       int
}

type dataxml struct {
	XMLName    xml.Name   `xml:"users"`
	Totalusers int        `xml:"Totalusers,attr"`
	Users      []Userdata `xml:"user"`
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

//func encryptFile(filename string, data []byte, passphrase string) {
func encryptFile(filename string, data []byte, passphrase string) {
	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(encrypt(data, passphrase))
}

func decryptFile(filename string, passphrase string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return decrypt(data, passphrase)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func LoadDatabases(databasenames []string, data *dataxml, passwords []string) error {
	for i, databasename := range databasenames {
		err := LoadDatabase(databasename, data, passwords[i])
		if err != nil {
			fmt.Println(i)
			return err
		}
	}
	return nil
}

func LoadDatabase(databasename string, data *dataxml, password string) error {
	filename := databasename + ".db"
	filename = filepath.Join(DataDir, filename)
	if _, err := os.Stat(filename); err == nil {
		// pattern exists

		// we unmarshal our byteArray which contains our
		// xmlFiles content into 'users' which we defined above
		xml.Unmarshal(decryptFile(filename, password), &data)
	} else {
		return err
	}

	return nil
}

func CreateDatabase(databasename string, password string) error {
	data := dataxml{}
	filename := databasename + ".db"
	filename = filepath.Join(DataDir, filename)
	output, err := xml.MarshalIndent(data, "  ", "    ")
	if err != nil {
		return err
	}

	encryptFile(filename, output, password)

	return nil
}

func CreateUserDatabase() error {
	Databasename, DatabasePassword := getUserDatabase()
	err := CreateDatabase(Databasename, DatabasePassword)
	if err != nil {
		return err
	}

	return nil
}

func DeleteDatabase(databasename string, password string) error {
	data := dataxml{}
	err := LoadDatabase(databasename, &data, password)
	if err != nil {
		return err
	}

	filename := databasename + ".db"
	filename = filepath.Join(DataDir, filename)

	err = os.Remove(filename)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUserDatabase(passwordCheck string) error {
	Databasename, DatabasePassword := getUserDatabase()
	if passwordCheck != DatabasePassword {
		return nil
	}
	err := DeleteDatabase(Databasename, DatabasePassword)
	if err != nil {
		return err
	}
	return nil
}

func SaveDatabase(databasename string, data *dataxml, password string) error {

	filename := databasename + ".db"
	filename = filepath.Join(DataDir, filename)

	output, err := xml.MarshalIndent(data, "  ", "    ")
	if err != nil {
		return err
	}

	encryptFile(filename, output, password)

	return nil
}

func SetUserDatabase(Databasename string, DatabasePassword string) error {
	userDatabasename = Databasename
	userDatabasePassword = DatabasePassword
	return nil
}

func getUserDatabase() (string, string) {
	var Databasename = userDatabasename
	var DatabasePassword = userDatabasePassword
	return Databasename, DatabasePassword
}

func SaveUser(user Userdata, saveMask DataMask) error {
	Databasename, DatabasePassword := getUserDatabase()
	var Databasenames = []string{Databasename}
	var DatabasePasswords = []string{DatabasePassword}
	var save Userdata
	var newUser bool = true
	var i int //search index where user matchs in data
	var data dataxml

	if !saveMask.Type && !saveMask.Username && !saveMask.Userid && !saveMask.PasswordHash && !saveMask.Email && (saveMask.Access == 0) {
		return nil
	}

	err := LoadDatabases(Databasenames, &data, DatabasePasswords)
	if err != nil {
		return err
	}
	totalusers := data.Totalusers

	user.Username = strings.ToLower(user.Username)
	// to do sanitize string

	for i = 0; i < len(data.Users); i++ {
		//find if user exist
		if user.Username == data.Users[i].Username {
			newUser = false
			break
		}
	}
	if newUser {
		if user.Userid < 0 {
			user.Userid = totalusers + 1
			totalusers++
		}
		data.Users = append(data.Users, user)
		data.Totalusers = totalusers
	} else {
		save = data.Users[i]
		if saveMask.Type {
			save.Type = user.Type
		}
		if saveMask.Username {
			save.Username = user.Username
		}
		// user ids persist
		/*if saveMask.Userid {
			save.Userid = user.Userid
		} */
		if saveMask.PasswordHash {
			save.PasswordHash = user.PasswordHash
		}
		if saveMask.Email {
			save.Email = user.Email
		}
		if saveMask.Access == 1 {
			save.Access = user.Access
		} else if saveMask.Access == 2 {
			save.Access = append(save.Access, user.Access...)
		}
		data.Users[i] = save
	}

	err = SaveDatabase(Databasename, &data, DatabasePassword)
	if err != nil {
		return err
	}

	return nil
}

func SaveUserType(Username string, Type string) error {
	savemask := DataMask{Type: true}
	user := Userdata{Username: Username, Type: Type}
	err := SaveUser(user, savemask)
	if err != nil {
		return err
	}
	return nil
}

func SaveUserPasswordHash(Username string, PasswordHash string) error {
	savemask := DataMask{PasswordHash: true}
	user := Userdata{Username: Username, PasswordHash: PasswordHash}
	err := SaveUser(user, savemask)
	if err != nil {
		return err
	}
	return nil
}

func SaveUserPassword(Username string, Password *string) error {
	PasswordHash, err := HashPassword(*Password)
	if err != nil {
		return err
	}
	err = SaveUserPasswordHash(Username, PasswordHash)
	if err != nil {
		return err
	}
	return nil
}

func SaveUserEmail(Username string, Email string) error {
	savemask := DataMask{Email: true}
	user := Userdata{Username: Username, Email: Email}
	err := SaveUser(user, savemask)
	if err != nil {
		return err
	}
	return nil
}

func SaveUserAccess(Username string, Access []string) error {
	savemask := DataMask{Access: 1}
	user := Userdata{Username: Username, Access: Access}
	err := SaveUser(user, savemask)
	if err != nil {
		return err
	}
	return nil
}

func SaveUserAppendAccess(Username string, Access []string) error {
	savemask := DataMask{Access: 2}
	user := Userdata{Username: Username, Access: Access}
	err := SaveUser(user, savemask)
	if err != nil {
		return err
	}
	return nil
}

func CheckUserPassword(Username string, password *string) (bool, error) {
	Databasename, DatabasePassword := getUserDatabase()
	var data dataxml

	err := LoadDatabase(Databasename, &data, DatabasePassword)
	if err != nil {
		return false, err
	}

	Username = strings.ToLower(Username)
	// to do sanitize string
	for i := 0; i < len(data.Users); i++ {
		//find if user exist
		if Username == data.Users[i].Username {
			if CheckPasswordHash(*password, data.Users[i].PasswordHash) {
				return true, nil
			}
			return false, nil
		}
	}

	return false, nil
}

func getUser(Username string, datamask DataMask) (Userdata, error) {
	Databasename, DatabasePassword := getUserDatabase()
	var data dataxml
	var user Userdata

	err := LoadDatabase(Databasename, &data, DatabasePassword)
	if err != nil {
		return Userdata{}, err
	}

	Username = strings.ToLower(Username)
	user.Username = Username
	// to do sanitize string
	for i := 0; i < len(data.Users); i++ {
		//find if user exist
		if Username == data.Users[i].Username {
			if datamask.Type {
				user.Type = data.Users[i].Type
			}
			if datamask.Userid {
				user.Userid = data.Users[i].Userid
			}
			if datamask.Email {
				user.Email = data.Users[i].Email
			}
			if datamask.Access > 0 {
				user.Access = data.Users[i].Access
			}
			return user, nil
		}
	}

	return Userdata{}, nil
}

func GetUser(Username string) (Userdata, error) {
	datamask := DataMask{Type: true, Userid: true, Email: true, Access: 1}
	user, err := getUser(Username, datamask)
	if err != nil {
		return Userdata{}, err
	}
	return user, nil
}

func GetUserType(Username string) (string, error) {
	datamask := DataMask{Type: true}
	user, err := getUser(Username, datamask)
	if err != nil {
		return "", err
	}
	return user.Type, nil
}

func GetUserUserid(Username string) (int, error) {
	datamask := DataMask{Userid: true}
	user, err := getUser(Username, datamask)
	if err != nil {
		return -1, err
	}
	return user.Userid, nil
}

func GetUserEmail(Username string) (string, error) {
	datamask := DataMask{Email: true}
	user, err := getUser(Username, datamask)
	if err != nil {
		return "", err
	}
	return user.Email, nil
}

func GetUserAccess(Username string) ([]string, error) {
	datamask := DataMask{Access: 1}
	user, err := getUser(Username, datamask)
	if err != nil {
		return []string{""}, err
	}
	return user.Access, nil
}

func NewUser(Username string, password *string, Email string, Access []string) error {
	PasswordHash, err := HashPassword(*password)
	if err != nil {
		return err
	}
	uniqueid := -1 // request new id

	//user data
	user := Userdata{Type: "Normal-v1",
		Username:     Username,
		Userid:       uniqueid,
		PasswordHash: PasswordHash,
		Email:        Email,
		Access:       Access}

	//save mask for new user
	saveMask := DataMask{Type: true,
		Username:     true,
		Userid:       true,
		PasswordHash: true,
		Email:        true,
		Access:       1}

	//pass new user data and complete mask to SaveUser
	err = SaveUser(user, saveMask)
	if err != nil {
		return err
	}
	return nil
}
