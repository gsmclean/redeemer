package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	helix "github.com/nicklaw5/helix/v2"
)

type eventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

func eventsub(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()
	// verify that the notification came from twitch using the secret.
	if !helix.VerifyEventSubNotification(sc.Webhook_Secret, r.Header, string(body)) {
		log.Println("no valid signature on subscription")
		return
	} else {
		log.Println("verified signature for subscription")
	}
	var vals eventSubNotification
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&vals)
	if err != nil {
		log.Println(err)
		return
	}
	// if there's a challenge in the request, respond with only the challenge to verify your eventsub.
	if vals.Challenge != "" {
		w.Write([]byte(vals.Challenge))
		return
	}
	var followEvent helix.EventSubChannelPointsCustomRewardRedemptionEvent
	err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&followEvent)
	if err != nil {
		stdLog.Printf("Error decoding json for sub event")
	}

	log.Printf("got follow webhook: %s redeemed %s\n", followEvent.UserName, followEvent.Reward.Title)
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func followHandle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()
	// verify that the notification came from twitch using the secret.
	if !helix.VerifyEventSubNotification(sc.Webhook_Secret, r.Header, string(body)) {
		log.Println("no valid signature on subscription")
		return
	} else {
		log.Println("verified signature for subscription")
	}
	var vals eventSubNotification
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&vals)
	if err != nil {
		log.Println(err)
		return
	}
	// if there's a challenge in the request, respond with only the challenge to verify your eventsub.
	if vals.Challenge != "" {
		w.Write([]byte(vals.Challenge))
		return
	}
	var followEvent helix.EventSubChannelFollowEvent
	err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&followEvent)
	if err != nil {
		panic(err.Error())
	}

	log.Printf("got poll webhook: %s was followed by %s\n", followEvent.BroadcasterUserName, followEvent.UserName)
	err = sdb.AddFollowEvent(followEvent)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func redeemHandle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()
	// verify that the notification came from twitch using the secret.
	if !helix.VerifyEventSubNotification(sc.Webhook_Secret, r.Header, string(body)) {
		log.Println("no valid signature on subscription")
		return
	} else {
		log.Println("verified signature for subscription")
	}
	var vals eventSubNotification
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&vals)
	if err != nil {
		log.Println(err)
		return
	}
	// if there's a challenge in the request, respond with only the challenge to verify your eventsub.
	if vals.Challenge != "" {
		w.Write([]byte(vals.Challenge))
		return
	}
	var redeemEvent helix.EventSubChannelPointsCustomRewardRedemptionEvent
	err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&redeemEvent)
	if err != nil {
		panic(err.Error())
	}
	var shtId = 488613652
	var spshtId = "1P76tbPV1vYZCwI6KNuFnFJmFXKQZZ1VIDXSnTJ89kE4"
	err = PushToSheet(int64(shtId), spshtId, []string{redeemEvent.BroadcasterUserName, redeemEvent.Reward.Title, redeemEvent.UserLogin, redeemEvent.UserName, redeemEvent.RedeemedAt.Time.Local().String()})
	if err != nil {
		stdLog.Printf("Error pushing to sheets: %v\n", err.Error())
	}

	log.Printf("got poll webhook: %s had a reward (%s) redmption by %s\n", redeemEvent.BroadcasterUserName, redeemEvent.Reward.Title, redeemEvent.UserName)
	err = sdb.AddRedeemEvent(redeemEvent)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func closeSubs(w http.ResponseWriter, r *http.Request) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     sc.Client_ID,
		ClientSecret: sc.Client_Secret,
	})
	if err != nil {
		stdLog.Printf("Error creating client for closing subs: %v\n", err.Error())
	}
	aatresp, err := client.RequestAppAccessToken([]string{"channel:read:redemptions"})
	if err != nil {
		panic("error on app token request")
	}
	client.SetAppAccessToken(aatresp.Data.AccessToken)

	subs, err := GetSubs("")
	if err != nil {
		panic("Failed on getsubs")
	}
	for _, v := range subs {
		cresp, err := client.RemoveEventSubSubscription(v.ID)
		if err != nil {
			panic("failed to remove")
		}
		fmt.Printf("%+v\n", cresp)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func GetSubs(tid string) ([]helix.EventSubSubscription, error) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     sc.Client_ID,
		ClientSecret: sc.Client_Secret,
	})
	if err != nil {
		return nil, err
	}
	aatresp, err := client.RequestAppAccessToken([]string{"channel:read:redemptions"})
	if err != nil {
		return nil, err
	}
	client.SetAppAccessToken(aatresp.Data.AccessToken)
	var esp helix.EventSubSubscriptionsParams
	if tid != "" {
		esp.UserID = tid
	}

	resp, err := client.GetEventSubSubscriptions(&esp)
	if err != nil {
		return nil, err
	}
	return resp.Data.EventSubSubscriptions, nil
}

func AddFollowSub(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic("error parsing sub")
	}
	client, err := helix.NewClient(&helix.Options{
		ClientID:     sc.Client_ID,
		ClientSecret: sc.Client_Secret,
	})
	if err != nil {
		panic("Creating sub client")
	}
	aatresp, err := client.RequestAppAccessToken([]string{"moderator:read:followers"})
	if err != nil {
		return
	}
	client.SetAppAccessToken(aatresp.Data.AccessToken)

	tid := r.PostForm.Get("tid")
	fmt.Println(tid)
	resp, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    "channel.follow",
		Version: "2",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: tid,
			ModeratorUserID:   tid,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%v/followevent", sc.BaseURL),
			Secret:   sc.Webhook_Secret,
		},
	})
	if err != nil {
		panic("Creating Sub")
	}

	fmt.Printf("%+v\n", resp)
	http.Redirect(w, r, "/profile", http.StatusFound)

}

func AddRedeemSub(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic("error parsing sub")
	}
	client, err := helix.NewClient(&helix.Options{
		ClientID:     sc.Client_ID,
		ClientSecret: sc.Client_Secret,
	})
	if err != nil {
		panic("Creating sub client")
	}
	aatresp, err := client.RequestAppAccessToken([]string{"moderator:read:followers"})
	if err != nil {
		return
	}
	client.SetAppAccessToken(aatresp.Data.AccessToken)

	tid := r.PostForm.Get("tid")
	fmt.Println(tid)
	resp, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    "channel.channel_points_custom_reward_redemption.add",
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: tid,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%v/redeemevent", sc.BaseURL),
			Secret:   sc.Webhook_Secret,
		},
	})
	if err != nil {
		panic("Creating Sub")
	}

	fmt.Printf("%+v\n", resp)
	http.Redirect(w, r, "/profile", http.StatusFound)

}
