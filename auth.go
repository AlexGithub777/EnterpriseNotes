package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/icza/session"
	"golang.org/x/crypto/bcrypt"
)

func (a *App) registerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.ServeFile(w, r, "tmpl/register.html")
        return
    }

    // Grab user info
    username := r.FormValue("username")
    password := r.FormValue("password")

    // Check existence of user
    var existingUser User
    err := a.db.QueryRow("SELECT username FROM users WHERE username=$1", username).Scan(&existingUser.Username)
    
    // Check for errors
    if err == nil {
        // User already exists, set a cookie with the error message
        http.SetCookie(w, &http.Cookie{
            Name:  "message",
            Value: "User already exists",
            Path:  "/", // Set the path as needed
        })
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    } else if err != sql.ErrNoRows {
        // An unexpected error occurred
        http.SetCookie(w, &http.Cookie{
            Name:  "message",
            Value: "Error checking user existence",
            Path:  "/", // Set the path as needed
        })
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // User doesn't exist, proceed with registration
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    checkInternalServerError(err, w)
    // Insert the user into the database
    _, err = a.db.Exec("INSERT INTO users(username, password) VALUES($1, $2)", username, hashedPassword)
    if err != nil {
        // Registration failed, set a cookie with the error message
        http.SetCookie(w, &http.Cookie{
            Name:  "message",
            Value: "Error registering user: " + err.Error(),
            Path:  "/", // Set the path as needed
        })
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // Registration was successful, set a cookie with the success message
    http.SetCookie(w, &http.Cookie{
        Name:  "message",
        Value: "Registration was successful. Please log in.",
        Path:  "/", // Set the path as needed
    })
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}


func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("Method %s", r.Method)

    // Check for a message cookie
    cookie, err := r.Cookie("message")
    var message string
    if err == nil {
        message = cookie.Value

        // Delete the cookie
        deleteCookie := http.Cookie{Name: "message", MaxAge: -1, Path: "/"}
        http.SetCookie(w, &deleteCookie)
    }

    if r.Method != "POST" {
        // Serve the login page and include the message
        tmpl, err := template.ParseFiles("tmpl/login.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Define a data structure to hold template variables
        data := struct {
            Message string
        }{
            Message: message,
        }

        // Execute the template and pass the data
        err = tmpl.Execute(w, data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }

        return
    }

    // grab user info from the submitted form
    username := r.FormValue("usrname")
    password := r.FormValue("psw")

    // query database to get matching username
    var user User
    err = a.db.QueryRow("SELECT username, password FROM users WHERE username=$1", username).Scan(&user.Username, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows {
            // Handle the case where no user with that username was found.
            // Set an error message
            http.SetCookie(w, &http.Cookie{
                Name:  "message",
                Value: "User not found.",
                Path:  "/", // Set the path as needed
            })
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        checkInternalServerError(err, w)
        return
    }

    // password is encrypted
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        // Password does not match
        // Set an error message
        http.SetCookie(w, &http.Cookie{
            Name:  "message",
            Value: "Invalid username or password.",
            Path:  "/", // Set the path as needed
        })
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // Successful login. New session with initial constant and variable attributes
    sess := session.NewSessionOptions(&session.SessOptions{
        CAttrs: map[string]interface{}{"username": user.Username},
        Attrs:  map[string]interface{}{"count": 1},
    })
    session.Add(sess, w)
    http.Redirect(w, r, "/list", http.StatusSeeOther)
}


func (a *App) logoutHandler(w http.ResponseWriter, r *http.Request) {

	// get the current session variables
	s := session.Get(r)
	log.Printf("User %s", s.CAttr("username").(string))
	session.Remove(s, w)
	s = nil
    session.Get()
	http.Redirect(w, r, "/login", 301)
}

func (a *App) isAuthenticated(w http.ResponseWriter, r *http.Request) {
	authenticated := false

	sess := session.Get(r)

	if sess != nil {
		u := sess.CAttr("username").(string)
		c := sess.Attr("count").(int)

		//just a simple authentication check for the current user
		if c > 0 && len(u) > 0 {
			authenticated = true
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/login", 301)
	}
}


// Initialize the session manager - this is a global
// For testing purposes, we want cookies to be sent over HTTP too (not just HTTPS)
// refer to the auth.go for the authentication handlers using the sessions
func (a *App) setupAuth() {
	session.Global.Close()
	session.Global = session.NewCookieManagerOptions(session.NewInMemStore(), &session.CookieMngrOptions{AllowHTTP: true})
}
