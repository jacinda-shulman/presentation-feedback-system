package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Middleware struct contains a pointer to the database
type Middleware struct {
	db *Database
}

// Encoder is a generic encoder type - used to wrap a json or xml encoder
type Encoder interface {
	Encode(v interface{}) error
}

// setIndent takes an Encoder, determines its type,
// and uses the appropriate function to set the indent (two spaces "  ")
func setIndent(e Encoder) error {
	switch t := e.(type) {
	case *json.Encoder:
		t.SetIndent("", "  ")
		return nil
	case *xml.Encoder:
		t.Indent("", "  ")
		return nil
	default:
		log.Printf("encoder type (%v) not recognized\n", t)
		return fmt.Errorf("encoder type (%v) not recognized", t)
	}
}

type EncoderKey string
type IDKey string

// Authenticate middleware function:
// Authorizes using HTTP Authorization scheme: Bearer realm (RFC 6750)
// https://tools.ietf.org/html/rfc6750#section-3.1
// Modified from source: http://www.gorillatoolkit.org/pkg/mux
func (mw Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			user        Account
			tokenString string
			token       string
			err         error
		)
		// Check authorization header
		tokenString = r.Header.Get("Authorization")

		// No token, or not proper format: return status 401 and example only
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			w.Header().Set("WWW-Authenticate", `Bearer realm="example"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Trim 'Bearer' prefix off of token
		token = strings.TrimPrefix(tokenString, "Bearer ")

		// Invalid token provided, return status 401 and error description
		q := `SELECT account_id FROM account WHERE token = $1`
		if err = mw.db.Get(&user, q, token); err != nil {
			w.Header().Set("WWW-Authenticate",
				`Bearer realm="example, error="invalid_token",
								error_description="Invalid access token provided"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Token is valid, authenticate user, add userID to context
		var idKey IDKey = "authUser"
		ctx := context.WithValue(r.Context(), idKey, user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetEncoder sets the appropriate encoder (json or xml) through context
// according to the request's "Accept" header. Json is set at default.
// Source: inspired by code shown in class; Author: Nicholas Boers
func (mw Middleware) SetEncoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			enc    Encoder
			encKey EncoderKey = "encKey"
			err    error
		)
		// From the request, get the data entered in the "Accept" header
		acceptType := r.Header.Get("Accept")
		if acceptType == "application/xml" {
			enc = xml.NewEncoder(w)
		} else if acceptType == "application/javascript" {
			enc = json.NewEncoder(w)
		} else {
			enc = json.NewEncoder(w)
		}

		// Set the indent to "  " on the encoder
		if err = setIndent(enc); err != nil {
			fmt.Println("Error came back from setIndent")
		}
		// Add encoder to context
		ctx := context.WithValue(r.Context(), encKey, &enc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs basic info about the request to os.Stdout
// Source: inspired by code shown in class; Author: Nicholas Boers
func (mw Middleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			entries   = map[string]interface{}{}
			accountID int
			ok        bool
			idKey     IDKey = "authUser"
		)
		entries["timeStamp"] = time.Now()
		entries["requestURI"] = r.RequestURI
		entries["method"] = r.Method

		// Get accountID from context
		if accountID, ok = r.Context().Value(idKey).(int); ok {
			entries["userID"] = accountID
		}
		entries["userAuthorized"] = ok

		enc := json.NewEncoder(os.Stdout)
		enc.Encode(entries)
		next.ServeHTTP(w, r)
	})
}
