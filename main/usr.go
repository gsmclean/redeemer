package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type UserAdd struct {
	*BasicData
	Token   string
	URL     string
	Message string
}
type UserData struct {
	ID      int16
	Name    string
	Chan_ID int16
	Login   bool
	Invite  bool
	Admin   bool
}

type GenericData struct {
	*BasicData
	Message string
}
type SignupData struct {
	*BasicData
	Message string
	Token   string
}

type UserManage struct {
	*BasicData
	Message string
	Users   []UserData
}

func handleUserAdd(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	uid := session.Values["id"].(int)
	var err error
	userData, err := sdb.GetUserData(uid)
	if err != nil {
		fileLog.Printf("Error Getting user data on useradd get: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/userAdd.html",
	}
	data := GenericData{
		BasicData: &userData,
		Message:   "",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleAddPost(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	uid := session.Values["id"].(int)
	var err error
	var data UserAdd
	bd, err := sdb.GetUserData(uid)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	bd.IsLoggedIn = session.Values["Authenticated"].(bool)
	token, err := sdb.NewUserInvite()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data = UserAdd{
		BasicData: &bd,
		Token:     token,
		Message:   "",
		URL:       sc.BaseURL,
	}

	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/userInvite.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleSignUp(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	var userData BasicData
	if session.IsNew {
		userData = BasicData{
			UserID:        0,
			UserName:      "",
			TwitchAccount: -1,
			IsLoggedIn:    false,
			CanInvite:     false,
			IsAdmin:       false,
		}
	}
	query := r.URL.Query()
	token := query["token"]
	if token == nil || token[0] == "" {
		fileLog.Printf("Error Getting user data on useradd get")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/userSignUp.html",
	}
	data := SignupData{
		BasicData: &userData,
		Message:   "",
		Token:     token[0],
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleUserSignup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get the username and password from the form
	username := strings.TrimSpace(r.Form.Get("username"))
	password := r.Form.Get("password")
	cpw := r.Form.Get("confirm")
	token := r.Form.Get("token")

	test, err := sdb.CheckToken(token)
	if err != nil {
		stdLog.Printf("Error checking token: %v\r", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var msg string
	var retry bool
	if test < 0 {
		msg = "Invalid Token"
		retry = true
	}
	if password != cpw {
		msg = "Mismatch Password"
		retry = true
	}
	if retry {
		userData := BasicData{
			UserID:        0,
			UserName:      "",
			TwitchAccount: -1,
			IsLoggedIn:    false,
			CanInvite:     false,
			IsAdmin:       false,
		}
		files := []string{
			"./templates/base.html",
			"./templates/title.html",
			"./templates/nav.html",
			"./templates/userSignUp.html",
		}
		data := SignupData{
			BasicData: &userData,
			Message:   msg,
			Token:     token,
		}
		ts, err := template.ParseFiles(files...)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		err = ts.ExecuteTemplate(w, "base", data)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	err = sdb.insertUser(username, password)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = sdb.DeleteToken(test)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func handleManage(w http.ResponseWriter, r *http.Request) {
	var data UserManage
	session, _ := store.Get(r, "session.id")
	uid := session.Values["id"].(int)
	var err error
	userData, err := sdb.GetUserData(uid)
	if err != nil {
		fileLog.Printf("Error Getting user data on useradd get: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	data.BasicData = &userData
	data.Users, err = sdb.GetUsers()
	if err != nil {
		fileLog.Printf("Error Getting user data on useradd get: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	data.Message = ""

	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/userManage.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handlePerms(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	stdLog.Printf("Post vars %v", r.Form)
	id, err := strconv.ParseInt(r.Form.Get("id"), 10, 32)
	if err != nil {
		stdLog.Printf("Error parsing %v", r.Form)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	var perms int16 = 0
	if r.Form.Get("login") != "" {
		perms += 1
	}
	if r.Form.Get("invite") != "" {
		perms += 2
	}
	if r.Form.Get("admin") != "" {
		perms += 4
	}
	stdLog.Printf("Permissions: %v\n", perms)
	err = sdb.UpdatePerms(id, perms)
	if err != nil {
		stdLog.Printf("Error on SQL Perms: %v Error: %v\n", perms, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

func handleUserDel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars["id"] == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	uid := vars["id"]
	id, err := strconv.ParseInt(uid, 10, 32)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = sdb.DeleteUser(id)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

type PwRes struct {
	*BasicData
	Token string
}

func handlePWRes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if vars["id"] == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	uid := vars["id"]
	id, err := strconv.ParseInt(uid, 10, 32)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	token, err := sdb.PasswordReset(id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := PwRes{Token: token}
	ts, err := template.ParseFiles("./templates/pwres.html")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.Execute(w, data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func handlePWChange(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	uid := session.Values["id"].(int)
	var err error
	bd, err := sdb.GetUserData(uid)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	bd.IsLoggedIn = session.Values["Authenticated"].(bool)
	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/chpwd.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.ExecuteTemplate(w, "base", bd)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handlePwdPost(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	uid := int64(session.Values["id"].(int))
	err := r.ParseForm()
	if err != nil {
		stdLog.Printf("Error parsing form %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	id, err := strconv.ParseInt(r.Form.Get("id"), 10, 32)
	if err != nil {
		stdLog.Printf("Error parsing id %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if uid != id {
		stdLog.Printf("ID's don't match, something is sketchy: form id %v, user id: %v\n", id, uid)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	opw := r.Form.Get("opd")
	npw := r.Form.Get("npd")
	cpw := r.Form.Get("cpd")
	if npw != cpw {
		stdLog.Printf("Pw Don't match\n")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = sdb.UpdatePassword(uid, opw, npw)
	if err != nil {
		stdLog.Printf("Error from update function: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// func tempReset(w http.ResponseWriter, r *http.Request) {
// 	t, err := sdb.PasswordReset(5)
// 	if err != nil {
// 		stdLog.Printf("Error: %v\n", err)
// 	}
// 	stdLog.Printf("token: %v\n", t)
// }
