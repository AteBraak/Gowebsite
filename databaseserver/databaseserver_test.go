package databaseserver

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("Password")
	if err != nil {
		t.Errorf("hashing failed during HashPassword call: %v", hash)
	}

	got := CheckPasswordHash("Password", hash)
	if got != true {
		t.Errorf("CheckPasswordHash(pass, hash) = %t; want true", got)
	}
}

func TestDatabase(t *testing.T) {
	var userDatabasename = "credentials_test"
	var userDatabasePassword = "password5"
	username1 := "albert4"
	password1 := "HopethisWorks1"
	username2 := "albert5"
	password2 := "HopethisWorks2"
	username3 := "albert6"
	password3 := "HopethisWorks3"
	password4 := "4HopethisWorks"

	err := SetUserDatabase(userDatabasename, userDatabasePassword)
	if err != nil {
		t.Errorf("SetUserDatabase(credentials_test, password5) = %t; want nil", err)
	}

	err = CreateUserDatabase()
	if err != nil {
		t.Errorf("CreateUserDatabase() = %t; want nil", err)
	}

	err = NewUser(username1, &password1, "al@gmail.wut", []string{"user", "admin"})
	if err != nil {
		t.Errorf("NewUser(username1, &password1, al@gmail.wut, []string{user, admin}) = %v; want nil", err)
	}

	test1, err := CheckUserPassword(username1, &password1)
	if err != nil {
		t.Errorf("CheckUserPassword(username1, &password1) = %v; want nil", err)
	}
	if !test1 {
		t.Errorf("checking for new user: CheckUserPassword(username1, &password1) = %t; want true", test1)
	}

	test2, err := CheckUserPassword(username2, &password1)
	if err != nil {
		t.Errorf("CheckUserPassword(username2, &password1) = %v; want nil", err)
	}
	if test2 {
		t.Errorf("checking for nonexistant user: CheckUserPassword(username2, &password1) = %t; want false", test2)
	}

	test3, err := CheckUserPassword(username1, &password2)
	if err != nil {
		t.Errorf("CheckUserPassword(username1, &password2) = %v; want nil", err)
	}
	if test3 {
		t.Errorf("Using incorrect password: CheckUserPassword(username1, &password2) = %t; want false", test3)
	}

	err = NewUser(username2, &password2, "al2@gmail.wut", []string{"user"})
	if err != nil {
		t.Errorf("NewUser(username2, &password2, al2@gmail.wut, []string{user}) = %v; want nil", err)
	}

	err = NewUser(username3, &password3, "al3@gmail.wut", []string{"group2"})
	if err != nil {
		t.Errorf("NewUser(username3, &password3, al3@gmail.wut, []string{group2}) = %v; want nil", err)
	}

	test4, err := CheckUserPassword(username2, &password2)
	if err != nil {
		t.Errorf("CheckUserPassword(username2, &password2) = %v; want nil", err)
	}
	if !test4 {
		t.Errorf("checking for new user: CheckUserPassword(username2, &password2) = %t; want true", test4)
	}

	test5, err := CheckUserPassword(username1, &password2)
	if err != nil {
		t.Errorf("CheckUserPassword(username1, &password2) = %v; want nil", err)
	}
	if test5 {
		t.Errorf("Using password of other user: CheckUserPassword(username1, &password2) = %t; want false", test5)
	}

	test6, err := CheckUserPassword(username2, &password1)
	if err != nil {
		t.Errorf("CheckUserPassword(username2, &password1) = %v; want nil", err)
	}
	if test6 {
		t.Errorf("Using password of other user: CheckUserPassword(username2, &password1) = %t; want false", test6)
	}

	test7, err := CheckUserPassword(username3, &password3)
	if err != nil {
		t.Errorf("CheckUserPassword(username3, &password3) = %v; want nil", err)
	}
	if !test7 {
		t.Errorf("checking for new user: CheckUserPassword(username3, &password3) = %t; want true", test7)
	}

	//GetUserType(Username string) (string, error)
	test8, err := GetUserType(username3)
	if err != nil {
		t.Errorf("GetUserType(username3) = %v; want nil", err)
	}
	if test8 != "Normal-v1" {
		t.Errorf("getting user type: GetUserType(username3) = %s; want Normal-v1", test8)
	}

	//GetUserUserid(Username string) (int, error)
	test9, err := GetUserUserid(username2)
	if err != nil {
		t.Errorf("GetUserType(username2) = %v; want nil", err)
	}
	if test9 != 2 {
		t.Errorf("getting user id: GetUserType(username2) = %d; want 2", test9)
	}

	//GetUserEmail(Username string) (string, error)
	test10, err := GetUserEmail(username1)
	if err != nil {
		t.Errorf("GetUserEmail(username1) = %v; want nil", err)
	}
	if test10 != "al@gmail.wut" {
		t.Errorf("getting user email: GetUserEmail(username1) = %s; want al@gmail.wut", test10)
	}

	//GetUserAccess(Username string) ([]string, error)
	test11, err := GetUserAccess(username1)
	if err != nil {
		t.Errorf("GetUserAccess(username1) = %v; want nil", err)
	}
	access := []string{"user", "admin"}
	if test11[0] != access[0] || test11[1] != access[1] {
		t.Errorf("getting user access: GetUserAccess(username1) = %s; want user, admin", test11)
	}

	//SaveUserType(Username string, Type string)
	err = SaveUserType(username1, "Admin")
	if err != nil {
		t.Errorf("SaveUserType(username1, Admin) = %v; want nil", err)
	}
	test12, err := GetUserType(username1)
	if test12 != "Admin" {
		t.Errorf("getting user type after save: SaveUserType(username1) = %s; want Admin", test12)
	}

	//SaveUserPasswordHash(Username string, PasswordHash string) error
	err = SaveUserPassword(username1, &password4)
	if err != nil {
		t.Errorf("SaveUserPassword(username1,password4) = %v; want nil", err)
	}
	test13, err := CheckUserPassword(username1, &password4)
	if !test13 {
		t.Errorf("checking for password update: CheckUserPassword(username1, &password4) = %t; want true", test13)
	}

	//SaveUserEmail(Username string, Email string) error
	err = SaveUserEmail(username2, "user_314@hotmail.com")
	if err != nil {
		t.Errorf("SaveUserEmail(username2, user_314@hotmail.com) = %v; want nil", err)
	}
	test14, err := GetUserEmail(username2)
	if test14 != "user_314@hotmail.com" {
		t.Errorf("getting user Email after save: GetUserEmail(username2) = %s; want user_314@hotmail.com", test14)
	}

	//SaveUserAccess(Username string, Access []string) error
	access = []string{"admin", "group2", "group1"} // previously user
	err = SaveUserAccess(username2, access)
	if err != nil {
		t.Errorf("SaveUserAccess(username2, access) = %v; want nil", err)
	}
	test15, err := GetUserAccess(username2)
	if test15[0] != access[0] || test15[1] != access[1] || test15[2] != access[2] {
		t.Errorf("getting user access after save: GetUserAccess(username2) = %s; want admin, group2, group1", test15)
	}

	//SaveUserAppendAccess(Username string, Access []string) error
	access = []string{"user", "group1"} // previously group2
	err = SaveUserAppendAccess(username3, access)
	if err != nil {
		t.Errorf("SaveUserAppendAccess(username3, access) = %v; want nil", err)
	}
	test16, err := GetUserAccess(username3)
	if test16[0] != "group2" || test16[1] != access[0] || test16[2] != access[1] {
		t.Errorf("getting user access after save: GetUserAccess(username3) = %s; want group2, user, group1", test16)
	}

	//DeleteUserDatabase() error
	err = DeleteUserDatabase(userDatabasePassword)
	if err != nil {
		t.Errorf("DeleteUserDatabase(userDatabasePassword) = %v; want nil", err)
	}
	test17, err := CheckUserPassword(username2, &password1)
	if err == nil {
		t.Errorf("DataBase Deleted: CheckUserPassword(username2, &password1) = %v; want err", err)
	}
	if test17 {
		t.Errorf("Using password of old database: CheckUserPassword(username2, &password1) = %t; want false", test17)
	}
}
