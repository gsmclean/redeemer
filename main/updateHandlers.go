package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func followStateEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fid := r.Form.Get("id")
		fmt.Printf("Saw fid: %v\n", fid)
		sid, err := strconv.ParseInt(r.Form.Get("state"), 10, 32)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = sdb.UpdateFollowState(fid, int(sid))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(200)
		return
	}
}

func followDeleteBtn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fid, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = sdb.DeleteFollow(int(fid))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)

}

func redeemStateEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		fid := r.Form.Get("id")
		fmt.Printf("Saw fid: %v\n", fid)
		sid, err := strconv.ParseInt(r.Form.Get("state"), 10, 32)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = sdb.UpdateRedeemState(fid, int(sid))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(200)
		return
	}
}

func redeemDeleteBtn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fid := vars["id"]
	err := sdb.DeleteRedeem(fid)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)

}
