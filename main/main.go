package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Data structure for the greeting template
type GreetingData struct {
	Name string
}

// Data structure for the login template
type LoginData struct {
	Message string
}

type SiteConfig struct {
	BaseURL        string
	DB_URL         string
	Client_ID      string
	Client_Secret  string
	Scope_String   string
	Webhook_Secret string
	Port           string
}

var (
	sdb     SiteDB
	sc      SiteConfig
	store   *sessions.CookieStore
	stdLog  *log.Logger
	fileLog *log.Logger
)

func init() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	stdLog = log.New(os.Stdout, "Warn: ", log.Ldate|log.Ltime)
	fileLog = log.New(file, "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	store = sessions.NewCookieStore([]byte("lkjlkjjljks"))
	err = godotenv.Load("./.env")
	if err != nil {
		stdLog.Printf("Error loading env: %v\n", err.Error())
		return
	}
	sc.BaseURL = os.Getenv("REDEEM_URL")
	sc.DB_URL = fmt.Sprintf("%v?sslmode=disable", os.Getenv("REDEEM_DB"))
	sc.Client_ID = os.Getenv("REDEEM_ID")
	sc.Client_Secret = os.Getenv("REDEEM_SECRET")
	sc.Scope_String = os.Getenv("REDEEM_SCOPE")
	sc.Webhook_Secret = os.Getenv("REDEEM_EVENT_SECRET")
	sc.Port = os.Getenv("REDEEM_PORT")
	if sc.BaseURL == "" || sc.DB_URL == "" || sc.Client_ID == "" || sc.Client_Secret == "" || sc.Scope_String == "" || sc.Webhook_Secret == "" {
		stdLog.Println("Error reading env variables")
		fileLog.Printf("Error reading env variables, %v", sc)
		return
	}

	if sc.Port == "" {
		sc.Port = "8083"
	}

	// Initialize the database

	err = sdb.Init(sc.DB_URL)
	if err != nil {
		stdLog.Println("Error initializing database:", err.Error())
		fileLog.Println("Error initializing database:", err.Error())
		return
	}
	defer sdb.DB.Close()
}

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Define a handler function for the greeting page

	// Register the handlers for the home, greet, and login pages
	router.HandleFunc("/", indexHandler).Methods(http.MethodGet)
	router.HandleFunc("/profile", ProfileHandler).Methods(http.MethodGet)
	router.HandleFunc("/logout", LogoutHandler)
	router.HandleFunc("/login", logonHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/handle", oAuthCallback)
	router.HandleFunc("/eventsub", eventsub)
	router.HandleFunc("/follow", followHandle)
	router.HandleFunc("/oauth", requestOAuth)
	router.HandleFunc("/followsub", AddFollowSub).Methods(http.MethodPost)
	router.HandleFunc("/followevent", followHandle)
	router.HandleFunc("/close", closeSubs)
	router.HandleFunc("/follows", followPage)
	router.HandleFunc("/followtable", followTable)
	router.HandleFunc("/redeemsub", AddRedeemSub).Methods(http.MethodPost)
	router.HandleFunc("/redeemevent", redeemHandle)
	router.HandleFunc("/redeems", redeemPage)
	router.HandleFunc("/redeemtable", redeemTable)
	router.HandleFunc("/follow", followStateEdit).Methods(http.MethodPost)
	router.HandleFunc("/follow/{id:[0-9]+}", followDeleteBtn)
	router.HandleFunc("/redeem", redeemStateEdit).Methods(http.MethodPost)
	router.HandleFunc("/redeem/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}", redeemDeleteBtn)

	// Serve static files from the "static" directory
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start the HTTP server on port 8080 using the router
	fmt.Println("Server is listening on :8083...")

	http.ListenAndServe(fmt.Sprintf(":%v", sc.Port), router)
}
