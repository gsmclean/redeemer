package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type UserAdd struct {
	*BasicData
	Token   string
	URL     string
	Message string
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
		"./templates/userManage.html",
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
	uid := session.Values["id"].(int)
	var err error
	userData, err := sdb.GetUserData(uid)
	if err != nil {
		fileLog.Printf("Error Getting user data on useradd get: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	query := r.URL.Query()
	token := query["token"]
	if token == nil || token[0] == "" {
		fileLog.Printf("Error Getting user data on useradd get: %v\n", err)
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
	http.Redirect(w, r, "/login", http.StatusFound)

}

func handleLink(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Link Endpoint")
}
