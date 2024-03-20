package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	helix "github.com/nicklaw5/helix/v2"
)

type oauthData struct {
	Url        string
	IsLoggedIn bool
}

func oAuthCallback(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()
	c := qp.Get("code")
	s := qp.Get("state")
	for k, v := range qp {
		fmt.Println(k, " => ", v)
	}
	client, err := helix.NewClient(&helix.Options{
		ClientID:     sc.Client_ID,
		ClientSecret: sc.Client_Secret,
		RedirectURI:  fmt.Sprintf("%v/handle", sc.BaseURL),
	})

	if err != nil {
		panic("Error on client for token")
	}
	state := string(s)
	code := string(c)

	resp, err := client.RequestUserAccessToken(code)
	if err != nil {
		panic("error on request token")
	}
	fmt.Printf("%+v\n", resp)

	//Set the access token on the client
	client.SetUserAccessToken(resp.Data.AccessToken)
	aatresp, err := client.RequestAppAccessToken([]string{sc.Scope_String})
	if err != nil {
		panic("error on app token request")
	}

	fmt.Printf("%+v\n", aatresp)
	client.SetAppAccessToken(aatresp.Data.AccessToken)
	uresp, err := client.GetUsers(&helix.UsersParams{})
	if err != nil {
		panic("error on get user")
	}
	id := uresp.Data.Users[0].ID
	name := uresp.Data.Users[0].Login
	err = sdb.HandleOauth(state, id, name)
	if err != nil {
		fmt.Println(err.Error())
		panic("error on Handle Oauth function")
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func requestOAuth(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session.id")
	login := session.Values["Authenticated"].(bool)
	client, err := helix.NewClient(&helix.Options{
		ClientID:    sc.Client_ID,
		RedirectURI: fmt.Sprintf("%v/handle", sc.BaseURL),
	})
	if err != nil {
		panic("failed on client init")
	}
	id := uuid.New().String()
	uid := session.Values["id"].(int)
	url := client.GetAuthorizationURL(&helix.AuthorizationURLParams{
		ResponseType: "code", // or "token"
		Scopes:       strings.Split(sc.Scope_String, "+"),
		State:        id,
		ForceVerify:  false,
	})
	sdb.AddState(id, uid)
	files := []string{
		"./templates/base.html",
		"./templates/title.html",
		"./templates/nav.html",
		"./templates/oauth.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = ts.ExecuteTemplate(w, "base", oauthData{Url: url, IsLoggedIn: login})
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
