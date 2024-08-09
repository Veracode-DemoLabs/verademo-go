package controllers

import (
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"verademo-go/src-app/shared/db"
	session "verademo-go/src-app/shared/session"
	"verademo-go/src-app/shared/view"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

type User struct {
	Username     string
	PasswordHint string
	CreatedAt    time.Time
	LastLogin    time.Time
	RealName     string
	BlabName     string
}
type Account struct {
	Error    string
	Username string
}
type Output struct {
	username string
	Error    string
}

// var db *sql.DB

func getMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func ShowLogin(w http.ResponseWriter, req *http.Request) {
	target := req.URL.Query().Get("target")
	username := req.URL.Query().Get("username")

	session, err := req.Cookie("session_username")
	if err == nil && session.Value != "" {
		log.Println("User is already logged in - redirecting...")
		if target != "" {
			http.Redirect(w, req, target, http.StatusFound)
		} else {
			http.Redirect(w, req, "feed.html", http.StatusFound)
		}
		return
	}

	user, err := createFromRequest(req)

	if err != nil {
		log.Println("Error creating user from request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if user != nil {
		http.SetCookie(w, &http.Cookie{Name: "session_username", Value: user.Username})
		log.Println("User is remembered - redirecting...")
		if target != "" {
			http.Redirect(w, req, target, http.StatusFound)
		} else {
			http.Redirect(w, req, "feed.html", http.StatusFound)
		}
		return
	} else {
		log.Println("User is not remembered")
	}

	if username == "" {
		username = ""
	}

	if target == "" {
		target = ""
	}
	log.Println("Entering showLogin with username " + username + " and target " + target)

	view.Render(w, "login.html", nil)
}

func ProcessLogin(w http.ResponseWriter, req *http.Request) {
	log.Println("Entering processLogin")

	// Form data check
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")
	remember := req.FormValue("remember")
	target := req.FormValue("target")

	var nextView string
	if target != "" {
		nextView = target
	} else {
		nextView = "/feed"
	}

	// Constructing SQL Query, have to figure out hashing
	sqlQuery := fmt.Sprintf("select username, password, password_hint, created_at, last_login, real_name, blab_name from users where username='", username, getMD5(password))
	result := struct {
		Username     string
		PasswordHint string
		CreatedAt    string
		LastLogin    string
		RealName     string
		BlabName     string
	}{}

	err = db.Db.QueryRow(sqlQuery, username, getMD5(password)).Scan(
		&result.Username,
		&result.PasswordHint,
		&result.CreatedAt,
		&result.LastLogin,
		&result.RealName,
		&result.BlabName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			http.Error(w, "Login failed. Please try again.", http.StatusUnauthorized)
			return
		}
		log.Println(err)
		http.Error(w, "An error has occured", http.StatusInternalServerError)
	}
	log.Println("User found. Redirecting...")

	http.SetCookie(w, &http.Cookie{Name: "username", Value: result.Username})

	// Handling the "remember me"
	if remember == "" {
		// Store details in session
		session, _ := store.Get(req, "session-name")
		session.Values["username"] = result.Username
		session.Values["password_hint"] = result.PasswordHint
		session.Values["created_at"] = result.CreatedAt
		session.Values["last_login"] = result.LastLogin
		session.Values["real_name"] = result.RealName
		session.Values["blab_name"] = result.BlabName
		session.Save(req, w)
	}
	// Updating last login time
	_, err = db.Db.Exec("UPDATE users SET last_login = NOW() WHERE username = ?", result.Username)
	if err != nil {
		log.Println()
		http.Error(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	// TOTP Handling
	if len(username) >= 4 && username[len(username)-4:] == "totp" {
		log.Println("User " + username + " Has TOTP Enabled!")
		session, _ := store.Get(req, "session-name")
		session.Values["totp_username"] = result.Username
		session.Save(req, w)
		nextView = "/totp"
	} else {
		log.Println("Setting session username to: " + username)
		session, _ := store.Get(req, "session-name")
		session.Values["username"] = result.Username
		session.Save(req, w)
	}

	log.Println("Redirecting to view: " + nextView)
	http.Redirect(w, req, nextView, http.StatusSeeOther)
}

func processLogout(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering processLogout")

	// Get the session
	session, _ := store.Get(r, "session-name")

	// Set the username to null (clear it)
	session.Values["username"] = nil

	// Save the session
	err := session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
	}

	// Optionally update response
	/*if err := updateInResponse(session.Values["username"], w); err != nil {
		log.Println("Error updating response:", err)
	}*/

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func ShowPasswordHint(w http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get("username")
	log.Printf("Entering password-hint with username: %s", username)

	if username == "" {
		http.Error(w, "No username provided, please type in your username first", http.StatusBadRequest)
		return
	}

	// Prepare the SQL query
	sqlQuery := "SELECT password_hint FROM users WHERE username = ?"
	log.Println(sqlQuery)

	var passwordHint string
	err := db.Db.QueryRow(sqlQuery, username).Scan(&passwordHint)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No password found for "+username, http.StatusNotFound)
		} else {
			log.Println("Error querying database:", err)
			http.Error(w, "ERROR!", http.StatusInternalServerError)
		}
		return
	}

	if len(passwordHint) > 0 {
		formatString := fmt.Sprintf("Username '%s' has password: %s%s", username, passwordHint[:2], strings.Repeat("*", len(passwordHint)-2))
		log.Println(formatString)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`"%s"`, formatString)))
	} else {
		http.Error(w, "No password found for "+username, http.StatusNotFound)
	}

}

func createFromRequest(req *http.Request) (*User, error) {
	cookie, err := req.Cookie("user")
	if err != nil {
		log.Println("No user cookie.")
		return nil, nil
	}

	cookieValue := cookie.Value
	decoded, err := base64.StdEncoding.DecodeString(cookieValue)
	if err != nil {
		log.Println("Error decoding cookie:", err)
		return nil, err
	}

	var user User
	if err := json.Unmarshal(decoded, &user); err != nil {
		log.Println("Error unmarshaling user from cookie:", err)
		return nil, err
	}

	log.Println("Username is:", user.Username)
	return &user, nil
}

func updateInResponse(currentUser *User, w http.ResponseWriter) error {
	userJSON, err := json.Marshal(currentUser)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(userJSON)
	http.SetCookie(w, &http.Cookie{
		Name:  "user",
		Value: encoded,
		Path:  "/",
	})
	return nil
}

func ShowRegister(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entering ShowRegister")
	p := &Account{Error: "Bad"}
	view.Render(w, "register.html", p)
}
func ProcessRegister(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Entering ProcessRegister")
	username := r.FormValue("username")
	output := Output{username: username}
	// This might be an error due to incorrect pointer logic
	if username == "" {
		output.Error = "No username provided, please type in your username first"
		view.Render(w, "register.html", output)
		return
	}
	fmt.Println()

	fmt.Println("Creating session")
	current_session := session.Instance(r)
	current_session.Values["username"] = r.FormValue("username")
	err := current_session.Save(r, w)
	if err != nil {
		log.Println("session error")
	}
	fmt.Println(current_session.Values["username"].(string))

	view.Render(w, "register.html", &output)
}
