package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/nicklaw5/helix/v2"
)

type BasicData struct {
	UserID        int
	UserName      string
	TwitchAccount int
	IsLoggedIn    bool
	CanLogin      bool
	CanInvite     bool
	IsAdmin       bool
}
type TwitchAccount struct {
	ChannelName string
	ChannelID   string
}
type ProfileData struct {
	*BasicData
	Subs  []helix.EventSubSubscription
	Chans []TwitchAccount
}
type FollowsData struct {
	*BasicData
	Channels []TwitchAccount
}
type FollowTable struct {
	Items  []FollowEvent
	States []EventStates
}
type EventStates struct {
	ID   int
	Name string
}
type RedeemTable struct {
	Items  []RedeemEvent
	States []EventStates
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/main.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/base.html",
	}
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
	} else {
		uid := session.Values["id"].(int)
		var err error
		userData, err = sdb.GetUserData(uid)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		userData.IsLoggedIn = session.Values["Authenticated"].(bool)
		session.Save(r, w)
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = ts.ExecuteTemplate(w, "base", userData)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func logonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Get the username and password from the form
		username := strings.TrimSpace(r.Form.Get("username"))
		password := r.Form.Get("password")
		session, _ := store.Get(r, "session.id")
		// Verify login credentials
		test, uid := sdb.verifyLogin(username, password)
		stdLog.Printf("User is logged in? %v\n", test)
		if test {
			// Redirect to the greeting page with the username as a query parameter
			session.Values["Authenticated"] = true
			session.Values["id"] = uid
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Display an error message on the login page
		files := []string{
			"./templates/base.html",
			"./templates/title.html",
			"./templates/nav.html",
			"./templates/login.html",
		}
		tmpl, err := template.ParseFiles(files...)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Execute the template with the error message
		err = tmpl.ExecuteTemplate(w, "base", LoginData{Message: "Invalid username or password"})
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}
	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/login.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if (session.Values["authenticated"] != nil) || session.Values["authenticated"] == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if session.Values["id"] == nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	uid := session.Values["id"].(int)

	if uid < 0 {
		http.Error(w, "Server errror", http.StatusInternalServerError)
		return
	}
	bd, err := sdb.GetUserData(uid)
	bd.IsLoggedIn = true
	if err != nil {
		http.Error(w, "Server errror", http.StatusInternalServerError)
		return
	}
	pd := ProfileData{
		BasicData: &bd,
	}
	chans, err := sdb.GetUserChannels(bd.UserID)
	pd.Chans = chans
	if err != nil {
		fmt.Printf("Error getting channels for user %v", bd.UserName)
	}

	for _, v := range pd.Chans {
		s, err := GetSubs(v.ChannelID)
		if err != nil {
			fmt.Printf("Error getting subs for channel %v", v.ChannelName)
		}
		pd.Subs = append(pd.Subs, s...)
	}
	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/profile.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = ts.ExecuteTemplate(w, "base", pd)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	session.Values["Authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func followPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if (session.Values["authenticated"] != nil) && session.Values["authenticated"] == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	uid := session.Values["id"].(int)
	if uid < 0 {
		http.Error(w, "Server errror", http.StatusInternalServerError)
		return
	}
	bd, err := sdb.GetUserData(uid)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	bd.IsLoggedIn = true
	fd := FollowsData{
		BasicData: &bd,
	}
	fd.Channels, err = sdb.GetUserChannels(bd.UserID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/follows.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = ts.ExecuteTemplate(w, "base", fd)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func followTable(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if (session.Values["authenticated"] != nil) && session.Values["authenticated"] == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	query := r.URL.Query()
	bcid := query["bcid"]
	if bcid == nil || bcid[0] == "" {
		Channels, err := sdb.GetUserChannels(session.Values["id"].(int))
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		for _, v := range Channels {
			bcid = append(bcid, v.ChannelID)
		}

	}
	var ftd FollowTable
	var err error
	ftd.Items, err = sdb.GetFollowEvents(bcid)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	ftd.States, err = sdb.GetStates()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	ts, err := template.ParseFiles("./templates/follow-table.html")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.Execute(w, ftd)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func redeemPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if (session.Values["authenticated"] != nil) && session.Values["authenticated"] == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	uid := session.Values["id"].(int)
	if uid < 0 {
		http.Error(w, "Server errror", http.StatusInternalServerError)
		return
	}
	bd, err := sdb.GetUserData(uid)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	bd.IsLoggedIn = true
	fd := FollowsData{
		BasicData: &bd,
	}
	fd.Channels, err = sdb.GetUserChannels(bd.UserID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/redeems.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = ts.ExecuteTemplate(w, "base", fd)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func redeemTable(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	if (session.Values["authenticated"] != nil) && session.Values["authenticated"] == false {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	query := r.URL.Query()
	bcid := query["bcid"]
	if bcid == nil || bcid[0] == "" {
		Channels, err := sdb.GetUserChannels(session.Values["id"].(int))
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		for _, v := range Channels {
			bcid = append(bcid, v.ChannelID)
		}

	}
	var rtd RedeemTable
	var err error
	rtd.Items, err = sdb.GetRedeemEvents(bcid)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	rtd.States, err = sdb.GetStates()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	ts, err := template.ParseFiles("./templates/redeem-table.html")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	err = ts.Execute(w, rtd)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}
