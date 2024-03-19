package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"

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
	login := IsLoggedIn(session)
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

// func getUserID(w http.ResponseWriter, r *http.Request) {
// 	client, err := helix.NewClient(&helix.Options{
// 		ClientID:        sc.Client_ID,
// 		UserAccessToken: "vriiopv8k4l6yfj7vq2drpuzz3m8p3",
// 	})
// 	if err != nil {
// 		// handle error
// 	}

// 	resp, err := client.GetUsers(&helix.UsersParams{
// 		IDs:    []string{"26301881", "18074328"},
// 		Logins: []string{"summit1g", "lirik"},
// 	})
// 	if err != nil {
// 		// handle error
// 	}

// 	fmt.Printf("%+v\n", resp)
// }

func testFlow(w http.ResponseWriter, r *http.Request) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     sc.Client_ID,
		ClientSecret: sc.Client_Secret,
		RedirectURI:  fmt.Sprintf("%v/handle", sc.BaseURL),
	})
	if err != nil {
		panic("on client create")
	}

	resp, err := client.RequestAppAccessToken([]string{"channel:read:redemptions"})
	if err != nil {
		panic("on app token request")
	}

	fmt.Printf("%+v\n", resp)

	// Set the access token on the client
	client.SetAppAccessToken(resp.Data.AccessToken)
	id := "151737789"

	sresp, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: id,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%v/eventsub", sc.BaseURL),
			Secret:   sc.Webhook_Secret,
		},
	})
	if err != nil {
		panic("creating sub")
	}

	fmt.Printf("%+v\n", sresp)

}

func IsLoggedIn(c *sessions.Session) bool {
	return c.Values["Authenticated"].(bool)
}
