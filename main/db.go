package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	pg "github.com/lib/pq"
	helix "github.com/nicklaw5/helix/v2"
	"golang.org/x/crypto/bcrypt"
)

type SiteDB struct {
	DB      *sql.DB
	ConnStr string
}

type FollowEvent struct {
	ID              int
	DT              string
	UserId          string
	UserName        string
	BroadcasterId   string
	BroadcasterName string
	Status          string
	StatusID        int
}

type RedeemEvent struct {
	ID              string
	DT              string
	UserId          string
	UserName        string
	BroadcasterId   string
	BroadcasterName string
	UserInput       string
	Status          string
	RewardId        string
	RewardTitle     string
	StateID         int
	StateName       string
}

func (sdb *SiteDB) Init(connStr string) error {
	var err error
	sdb.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	_, err = sdb.DB.Exec(`
CREATE SEQUENCE IF NOT EXISTS auth_pending_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."auth_pending" (
    "id" integer DEFAULT nextval('auth_pending_id_seq') NOT NULL,
    "state" text NOT NULL,
    "user_id" integer NOT NULL,
    CONSTRAINT "auth_pending_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


CREATE SEQUENCE IF NOT EXISTS "channelInvites_id_seq" INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."channel_invites" (
    "id" integer DEFAULT nextval('"channelInvites_id_seq"') NOT NULL,
    "twitch_id" text NOT NULL,
    "user_id" integer NOT NULL,
    "token" text NOT NULL,
    CONSTRAINT "channelInvites_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


CREATE SEQUENCE IF NOT EXISTS follow_events_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."follow_events" (
    "date" timestamp NOT NULL,
    "user_id" text NOT NULL,
    "user_name" text NOT NULL,
    "broadcaster_id" text NOT NULL,
    "id" integer DEFAULT nextval('follow_events_id_seq') NOT NULL,
    "state_id" integer DEFAULT '1' NOT NULL,
    CONSTRAINT "follow_events_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


CREATE SEQUENCE IF NOT EXISTS invite_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."invite" (
    "id" integer DEFAULT nextval('invite_id_seq') NOT NULL,
    "token" text NOT NULL,
    CONSTRAINT "invite_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


CREATE TABLE IF NOT EXISTS "public"."redeem_events" (
    "id" text NOT NULL,
    "broadcaster_id" text NOT NULL,
    "user_id" text NOT NULL,
    "user_name" text,
    "user_input" text,
    "status" text,
    "reward_id" text,
    "reward_title" text,
    "time_stamp" timestamp,
    "state_id" integer DEFAULT '1' NOT NULL,
    CONSTRAINT "redeem_events_id" PRIMARY KEY ("id")
) WITH (oids = false);


CREATE SEQUENCE IF NOT EXISTS states_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."states" (
    "id" integer DEFAULT nextval('states_id_seq') NOT NULL,
    "name" text NOT NULL,
    CONSTRAINT "states_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


CREATE SEQUENCE IF NOT EXISTS "twitchAccounts_id_seq" INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."twitch_accounts" (
    "id" integer DEFAULT nextval('"twitchAccounts_id_seq"') NOT NULL,
    "channel_name" text NOT NULL,
    "channel_id" text NOT NULL,
    CONSTRAINT "twitchAccounts_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "twitch_accounts_channel_id" UNIQUE ("channel_id")
) WITH (oids = false);


CREATE SEQUENCE IF NOT EXISTS "twitchRelations_id_seq" INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."twitch_relations" (
    "id" integer DEFAULT nextval('"twitchRelations_id_seq"') NOT NULL,
    "user_id" integer NOT NULL,
    "channel_id" integer NOT NULL,
    CONSTRAINT "twitchRelations_pkey" PRIMARY KEY ("id")
) WITH (oids = false);


CREATE SEQUENCE IF NOT EXISTS users_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."users" (
    "id" integer DEFAULT nextval('users_id_seq') NOT NULL,
    "user_name" text NOT NULL,
    "password" text NOT NULL,
    "twitch_account" integer,
    CONSTRAINT "users_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "users_username_key" UNIQUE ("user_name")
) WITH (oids = false);
	`)
	if err != nil {
		return err
	}
	sdb.DB.Exec(`
	LTER TABLE ONLY "public"."follow_events" ADD CONSTRAINT "follow_events_state_id_fkey" FOREIGN KEY (state_id) REFERENCES states(id) NOT DEFERRABLE;

ALTER TABLE ONLY "public"."redeem_events" ADD CONSTRAINT "redeem_events_state_id_fkey" FOREIGN KEY (state_id) REFERENCES states(id) NOT DEFERRABLE;

ALTER TABLE ONLY "public"."twitch_relations" ADD CONSTRAINT "twitchRelations_twitchid_fkey" FOREIGN KEY (channel_id) REFERENCES twitch_accounts(id) NOT DEFERRABLE;
ALTER TABLE ONLY "public"."twitch_relations" ADD CONSTRAINT "twitchRelations_userId_fkey" FOREIGN KEY (user_id) REFERENCES users(id) NOT DEFERRABLE;

ALTER TABLE ONLY "public"."users" ADD CONSTRAINT "users_twitchAccount_fkey" FOREIGN KEY (twitch_account) REFERENCES twitch_accounts(id) NOT DEFERRABLE;
	`)
	res, err := sdb.DB.Query("SELECT * FROM USERS")
	if err != nil {
		return err
	}
	test := res.Next()
	if !test {
		usr := os.Getenv("REDEEM_USER")
		pw := os.Getenv("REDEEM_PASS")
		if usr == "" || pw == "" {
			return fmt.Errorf("error getting fresh user/pass from env")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = sdb.DB.Exec("INSERT INTO users (user_name, password) VALUES ($1, $2)", usr, hashedPassword)
		return err
	}
	return nil
}

func (sdb SiteDB) insertUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = sdb.DB.QueryRow("SELECT password FROM users WHERE user_name = $1", username).Scan()
	if err == nil {
		stdLog.Printf("Error adding user, user exists: %v", username)
		return fmt.Errorf("user %v already exists", username)
	}
	_, err = sdb.DB.Exec("INSERT INTO users (user_name, password) VALUES ($1, $2)", username, hashedPassword)

	return err
}

func (sdb SiteDB) verifyLogin(username, password string) (bool, int) {
	// Retrieve the hashed password from the database
	var hashedPassword string
	var suid string
	var perms sql.NullInt16
	rows := sdb.DB.QueryRow("SELECT password, id, permissions FROM users WHERE user_name = $1", username)
	err := rows.Scan(&hashedPassword, &suid, &perms)
	if err != nil {
		stdLog.Printf("Error scanning: %v", err)
		return false, -1
	}

	if !perms.Valid {
		stdLog.Printf("Bad perms: %v", perms)
		return false, -1
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, -1
	}
	uid, err := strconv.ParseInt(suid, 10, 32)
	if err != nil {
		return false, -1
	}
	return (perms.Int16&1 > 0), int(uid)
}

func (sdb SiteDB) GetUserChannels(user int) ([]TwitchAccount, error) {
	rows, err := sdb.DB.Query(`
	SELECT
		twitch_accounts.channel_id, twitch_accounts.channel_name
	FROM
		twitch_relations
	INNER JOIN twitch_accounts
		ON twitch_relations.channel_id = twitch_accounts.id
	WHERE twitch_relations.user_id = $1

	`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []TwitchAccount
	for rows.Next() {
		var channel TwitchAccount
		if err := rows.Scan(&channel.ChannelID, &channel.ChannelName); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil

}

func (sdb SiteDB) GetFollowEvents(broadcasters []string) ([]FollowEvent, error) {
	stmt, err := sdb.DB.Prepare(`
		SELECT
			follow_events.id, follow_events.date, follow_events.user_id, follow_events.user_name, follow_events.broadcaster_id, twitch_accounts.channel_name, follow_events.state_id, states.name
		FROM
			follow_events
		INNER JOIN twitch_accounts ON follow_events.broadcaster_id = twitch_accounts.channel_id
		INNER JOIN states ON follow_events.state_id = states.id
		WHERE follow_events.broadcaster_id = ANY( $1 )
		`)
	if err != nil {
		stdLog.Printf("Failed to construct statement for Follow events for %v: %v\r", broadcasters, err.Error())
		return nil, err
	}
	rows, err := stmt.Query(pg.Array(broadcasters))

	var vals []FollowEvent
	if err != nil {
		stdLog.Printf("Failed to query database for Follow events for %v: %v\r", broadcasters, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sid, date, userId, userName, broadcasterID, broadcasterName, stateId, stateName string
		if err := rows.Scan(&sid, &date, &userId, &userName, &broadcasterID, &broadcasterName, &stateId, &stateName); err != nil {
			stdLog.Printf("Failed to scan values while parsing Follow events for %v: %v\r", broadcasters, err.Error())
			return nil, err
		}
		fid, err := strconv.ParseInt(sid, 10, 32)
		if err != nil {
			stdLog.Printf("Failed to parse event ID in Follow events for %v: %v\r", sid, err.Error())
			return nil, err
		}
		stateInt, err := strconv.ParseInt(stateId, 10, 32)
		if err != nil {
			stdLog.Printf("Failed to parse state ID in Follow events for %v: %v\r", stateId, err.Error())
			return nil, err
		}
		vals = append(vals, FollowEvent{
			ID:              int(fid),
			DT:              date,
			UserId:          userId,
			UserName:        userName,
			BroadcasterId:   broadcasterID,
			BroadcasterName: broadcasterName,
			StatusID:        int(stateInt),
			Status:          stateName,
		})

	}
	return vals, nil

}

func (sdb SiteDB) GetRedeemEvents(broadcasters []string) ([]RedeemEvent, error) {

	stmt, err := sdb.DB.Prepare(`
	SELECT
		redeem_events.id, redeem_events.time_stamp, redeem_events.user_id, redeem_events.user_name, redeem_events.broadcaster_id, redeem_events.user_input, redeem_events.status, redeem_events.reward_id, redeem_events.reward_title, twitch_accounts.channel_name, redeem_events.state_id, states.name
	FROM
		redeem_events
	INNER JOIN twitch_accounts ON redeem_events.broadcaster_id = twitch_accounts.channel_id
	INNER JOIN states ON redeem_events.state_id = states.id
	WHERE broadcaster_id = ANY ( $1 )
	`)
	if err != nil {
		stdLog.Printf("Failed to prepare statement for redeem events for %v: %v\r", broadcasters, err.Error())
		return nil, err
	}
	rows, err := stmt.Query(pg.Array(broadcasters))

	if err != nil {
		stdLog.Printf("Failed to query database for redeem events for %v: %v\r", broadcasters, err.Error())
		return nil, err
	}
	defer rows.Close()

	var vals []RedeemEvent
	for rows.Next() {
		var rid, date, userId, userName, broadcasterID, broadcaster, userInput, rewStatus, rewId, rewTitle, stateid, stateName string
		if err := rows.Scan(&rid, &date, &userId, &userName, &broadcasterID, &userInput, &rewStatus, &rewId, &rewTitle, &broadcaster, &stateid, &stateName); err != nil {
			return nil, err
		}
		sid, err := strconv.ParseInt(stateid, 10, 32)
		if err != nil {
			return nil, err
		}
		vals = append(vals, RedeemEvent{
			ID:              rid,
			DT:              date,
			UserId:          userId,
			UserName:        userName,
			BroadcasterId:   broadcasterID,
			BroadcasterName: broadcaster,
			UserInput:       userInput,
			Status:          rewStatus,
			RewardId:        rewId,
			RewardTitle:     rewTitle,
			StateName:       stateName,
			StateID:         int(sid),
		})
	}

	return vals, nil

}

func (sdb SiteDB) AddChannel(userId int32, ta TwitchAccount) error {
	r := sdb.DB.QueryRow("INSERT INTO twitch_accounts (channel_name, channel_id) VALUES ($1, $2) RETURNING id", ta.ChannelName, ta.ChannelID)
	var rid int
	err := r.Scan(&rid)
	if err != nil {
		return err
	}
	_, err = sdb.DB.Exec("INSERT INTO twitch_relations (user_id, channel_id) VALUES ($1, $2) RETURNING id", userId, rid)
	if err != nil {
		return err
	}
	_, err = sdb.DB.Exec("UPDATE users SET twitch_account = $1 WHERE id = $2", rid, userId)
	if err != nil {
		return err
	}
	return nil

}

func (sdb SiteDB) AddRelation(userId string, channelId string) error {
	_, err := sdb.DB.Exec("INSERT INTO twitch_relations (user_id, twitch_id) VALUES ($1, $2)", userId, channelId)
	return err
}

func (sdb SiteDB) AddFollowEvent(fe helix.EventSubChannelFollowEvent) error {
	_, err := sdb.DB.Exec("INSERT INTO follow_events (date, user_id, user_name, broadcaster_id) VALUES ($1, $2, $3, $4)", fe.FollowedAt.Time, fe.UserID, fe.UserName, fe.BroadcasterUserID)
	return err
}

func (sdb SiteDB) AddRedeemEvent(re helix.EventSubChannelPointsCustomRewardRedemptionEvent) error {
	_, err := sdb.DB.Exec(`
	INSERT INTO
		redeem_events (id, broadcaster_id, user_id, user_name, user_input, status, reward_id, reward_title, time_stamp)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, re.ID, re.BroadcasterUserID, re.UserID, re.UserName, re.UserInput, re.Status, re.Reward.ID, re.Reward.Title, re.RedeemedAt.Time)
	return err
}

func (sdb SiteDB) GetUserData(uid int) (BasicData, error) {
	var bd BasicData
	var ta sql.NullInt32
	var p sql.NullInt16
	err := sdb.DB.QueryRow("SELECT id, user_name, twitch_account, permissions FROM users WHERE id = $1", uid).Scan(&bd.UserID, &bd.UserName, &ta, &p)
	if ta.Valid {
		bd.TwitchAccount = int(ta.Int32)
	} else {
		bd.TwitchAccount = -1
	}
	if p.Valid {
		bd.IsAdmin = (p.Int16 & 4) > 0
		bd.CanInvite = (p.Int16 & 2) > 0
		bd.CanLogin = (p.Int16 & 1) > 0
	} else {
		bd.IsAdmin = false
		bd.CanInvite = false
		bd.CanLogin = false
	}
	bd.UserID = uid
	return bd, err
}

func (sdb SiteDB) InsertOauthState(uid int, state string) error {
	_, err := sdb.DB.Exec(`
	INSERT INTO
		auth_pending (state, user_id)
	VALUES ($1, $2)
	`, state, uid)
	return err
}

func (sdb SiteDB) HandleOauth(state string, tid string, name string) error {
	var suid sql.NullInt32
	err := sdb.DB.QueryRow("SELECT user_id FROM auth_pending WHERE state = $1", state).Scan(&suid)
	if err != nil {
		return err
	}
	if !suid.Valid {
		return err
	}
	uid := suid.Int32
	err = sdb.AddChannel(uid, TwitchAccount{ChannelID: tid, ChannelName: name})
	if err != nil {
		return err
	}
	_, err = sdb.DB.Exec("DELETE FROM auth_pending WHERE state = $1", state)
	return err

}

func (sdb SiteDB) AddState(state string, uid int) error {
	_, err := sdb.DB.Exec("INSERT INTO auth_pending (state, user_id) VALUES ($1, $2)", state, uid)
	return err

}

func (sdb SiteDB) GetStates() ([]EventStates, error) {
	rows, err := sdb.DB.Query("SELECT id, name FROM states")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ess []EventStates
	for rows.Next() {
		var es EventStates
		if err := rows.Scan(&es.ID, &es.Name); err != nil {
			return nil, err
		}
		ess = append(ess, es)
	}

	return ess, nil

}

func (sdb SiteDB) UpdateFollowState(fid string, sid int) error {
	_, err := sdb.DB.Exec("UPDATE follow_events SET state_id = $1 WHERE id = $2", sid, fid)
	return err
}

func (sdb SiteDB) DeleteFollow(fid int) error {
	_, err := sdb.DB.Exec("DELETE FROM follow_events WHERE id = $1", fid)
	return err
}

func (sdb SiteDB) UpdateRedeemState(fid string, sid int) error {
	_, err := sdb.DB.Exec("UPDATE redeem_events SET state_id = $1 WHERE id = $2", sid, fid)
	return err
}

func (sdb SiteDB) DeleteRedeem(fid string) error {
	_, err := sdb.DB.Exec("DELETE FROM redeem_events WHERE id = $1", fid)
	return err
}

func (sdb SiteDB) NewUserInvite() (string, error) {
	token := uuid.New().String()
	_, err := sdb.DB.Exec("INSERT INTO invite (token) VALUES ($1)", token)
	return token, err
}

func (sdb SiteDB) CheckToken(token string) (int16, error) {
	res, err := sdb.DB.Query("SELECT id FROM USERS")
	if err != nil {
		return -1, err
	}
	test := res.Next()
	if !test {
		stdLog.Printf("Token Failure\n")
		return -1, nil
	}
	var id sql.NullInt16
	err = res.Scan(&id)
	if err != nil {
		return -1, err
	}
	if !id.Valid {
		return -1, err
	}
	return id.Int16, nil

}
