package main

import (
	"fmt"
	"net/http"
)

func handleUserAdd(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddUser Endpoint")
}

func handleAddPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Post from add user")
}

func handleSignUp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Signup Endpoint")
}

func handleUserSignup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Post from user signup")
}

func handleLink(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Link Endpoint")
}
