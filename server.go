/*
CMPT 315 (Winter 2019)
Assign. 1: Presentation Feedback System (Backend)
Author: Jacinda Shulman

This file implements the server and request handler functionality
*/

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Handler struct contains a pointer to the database
type Handler struct {
	*Database
}

// Handle
func (h *Handler) handleCheckToken(w http.ResponseWriter, r *http.Request) {
	var (
		accountID int
		ok        bool
		idKey     IDKey = "authUser"
		firstName string
		err       error
	)
	// Get accountID from context
	if accountID, ok = r.Context().Value(idKey).(int); !ok {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	// Get name of the user
	if firstName, err = h.Database.NameFromID(accountID); err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	u := SimpleUser{
		FirstName: firstName,
		ID:        accountID,
	}
	fmt.Println(u.FirstName, u.ID)
	encodeResults(u, w, r)
}

// Handle retrieving the question set
func (h *Handler) handleGetQuestionSet(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		results []Question
	)
	if results, err = h.Database.GetQuestionSet(); err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	encodeResults(results, w, r)
}

// Handle getting the list of presenters (all data)
func (h *Handler) handleGetPresenters(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		results []Presenter
	)
	if results, err = h.Database.GetPresenters(); err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	encodeResults(results, w, r)
}

// Handle getting the list of presentation titles
func (h *Handler) handleGetPresentations(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		results []Presentation
	)
	if results, err = h.Database.GetPresentations(); err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	encodeResults(results, w, r)
}

// Handle POST new form
// Returns a struct representing that form and a range of answerIDs
//	associated with the form
func (h *Handler) handlePostNewForm(w http.ResponseWriter, r *http.Request) {
	var (
		body        []byte
		err         error
		status      int
		idKey       IDKey = "authUser"
		evaluatorID int
		presenterID int
		ok          bool
		form        Form
		answerIDs   []int
	)
	// Get the details from the body, return 400 if missing
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	// Read presenterID, send 400 if it can't be converted to an int
	if presenterID, err = strconv.Atoi(string(body)); err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	// Get evaluatorID from context
	// if not found, return 500 (it's an error with the context)
	if evaluatorID, ok = r.Context().Value(idKey).(int); !ok {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	// Call to create form in the database
	// CreateForm sends back http status codes according to success or failure
	form, answerIDs, status = h.Database.CreateForm(presenterID, evaluatorID)
	if status == 500 || status == 400 {
		http.Error(w, http.StatusText(status), status)
		return
	}
	// Success; send 201 Created or 409 Conflict and the structs
	w.WriteHeader(status)
	res := WrappedForm{
		Form:      form,
		AnswerIDs: answerIDs,
	}
	encodeResults(res, w, r)
}

// Handle deletions handles both clearing forms and deleting them
func (h *Handler) handleDeletions(w http.ResponseWriter, r *http.Request) {
	// Get formID from Vars
	vars := mux.Vars(r)
	formID, err := strconv.Atoi(vars["formID"])
	fmt.Printf("formID is %d\n", formID)
	if err != nil {
		// Since mux only routes correct paths, this is not client error
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	// Get accountID from context
	var (
		accountID int
		ok        bool
		idKey     IDKey = "authUser"
	)
	if accountID, ok = r.Context().Value(idKey).(int); !ok {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	// if user requested to DELETE FORM
	path := strings.Split(r.RequestURI, "/")
	if len(path) == 3 {
		// Call DeleteForm
		status, err := h.Database.DeleteForm(formID, accountID)
		if status != 200 {
			if err != nil {
				fmt.Print(err)
			}
			http.Error(w, http.StatusText(status), status)
			return
		}
		fmt.Fprintf(w, "Form %d was deleted successfully.\n", formID)

	} else if len(path) > 3 {
		// user requested CLEAR ANSWERS
		// Call ClearForm
		status, err := h.Database.ClearForm(formID, accountID)
		if status != 200 {
			log.Printf("error: %v\n", err)
			http.Error(w, http.StatusText(status), status)
			return
		}
		fmt.Fprintf(w, "Form %d was cleared successfully.\n", formID)
	}
}

// Handles answer updates by formID and qID
func (h *Handler) handlePutAnswerByForm(w http.ResponseWriter, r *http.Request) {
	var (
		formID, qID, answerID int
		err                   error
	)
	// get formID and questionID from vars
	vars := mux.Vars(r)
	formID, err = strconv.Atoi(vars["formID"])
	qID, err = strconv.Atoi(vars["qID"])
	if err != nil {
		// Since mux only routes correct paths, this is not client error
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	// Get the answerID from the form and question ids
	if answerID, err = h.Database.AnswerFromForm(formID, qID); err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	h.handleAnswerUpdate(answerID, w, r)
}

// Handles answer updates by answerID
func (h *Handler) handlePutAnswerByID(w http.ResponseWriter, r *http.Request) {
	// get answerID from vars
	vars := mux.Vars(r)
	var (
		answerID int
		err      error
	)
	if answerID, err = strconv.Atoi(vars["answerID"]); err != nil {
		// Since mux only routes correct paths, this is not client error
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	h.handleAnswerUpdate(answerID, w, r)
}

// Handle PUT new answer
func (h *Handler) handleAnswerUpdate(answerID int, w http.ResponseWriter, r *http.Request) {
	var (
		accountID int
		ok        bool
		idKey     IDKey = "authUser"
		body      []byte
		err       error
	)
	// Get accountID from context
	if accountID, ok = r.Context().Value(idKey).(int); !ok {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	// Get answer value from request body
	if body, err = ioutil.ReadAll(r.Body); err != nil || len(body) == 0 {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	value := string(body)
	// Call UpdateAnswer
	status, err := h.Database.UpdateAnswer(answerID, value, accountID)
	if status != 200 {
		log.Printf("error: %v\n", err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	fmt.Fprintf(w, "Question %d was updated successfully.\n", answerID)
}

// Handle GET all answers for a form
func (h *Handler) handleGetAnswers(w http.ResponseWriter, r *http.Request) {
	var (
		formID, status int
		err            error
		results        []Answer
	)
	// get formID from vars
	vars := mux.Vars(r)
	formID, err = strconv.Atoi(vars["formID"])
	if err != nil {
		// Since mux only routes correct paths, this is not client error
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	if results, status, _ = h.Database.GetAnswers(formID); status != 200 {
		http.Error(w, http.StatusText(status), status)
		return
	}
	encodeResults(results, w, r)
}

// Encode using the encoder sent through context
func encodeResults(results interface{}, w http.ResponseWriter, r *http.Request) {
	var (
		enc Encoder
		err error
	)
	if enc, err = getEncoderFromContext(r.Context()); err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	if err = enc.Encode(results); err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	}
}

func getEncoderFromContext(ctx context.Context) (Encoder, error) {
	var (
		encPointer *Encoder
		encKey     EncoderKey = "encKey"
		ok         bool
	)
	// Try to get pointer to Encoder; return error if it fails
	if encPointer, ok = ctx.Value(encKey).(*Encoder); !ok {
		log.Println("Error occurred when extracting encoder from context")
		return nil, fmt.Errorf("Error extracting encoder from context")
	}
	// Success; return pointer
	return *encPointer, nil
}

func registerHandlers(r *mux.Router, h Handler) {
	r.Path("/tokens").Methods("GET").HandlerFunc(h.handleCheckToken)
	r.Path("/questions").Methods("GET").HandlerFunc(h.handleGetQuestionSet)
	r.Path("/forms").Methods("POST").HandlerFunc(h.handlePostNewForm)
	r.Path("/answers/{answerID:[0-9]+}").Methods("PUT").HandlerFunc(h.handlePutAnswerByID)

	// "/presenters" & "/presentations"
	r.Path("/presenters").HandlerFunc(h.handleGetPresenters)
	r.Path("/presentations").HandlerFunc(h.handleGetPresentations)

	// accessing specific forms with formID in the path
	f := r.PathPrefix("/forms/{formID:[0-9]+}").Subrouter()
	f.Path("").Methods("DELETE").HandlerFunc(h.handleDeletions)
	f.Path("/answers").Methods("DELETE").HandlerFunc(h.handleDeletions)
	f.Path("/answers").Methods("GET").HandlerFunc(h.handleGetAnswers)
	f.Path("/questions/{qID:[0-9]+}").Methods("PUT").HandlerFunc(h.handlePutAnswerByForm)
}

// Source: The following inspired by code written in lab04
func main() {
	// Attempt to connect to the database. If it fails, quit and print error message.
	connect := "dbname=assign1 user=postgres host=localhost port=5432 sslmode=disable"
	db, err := connectToDB(connect)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer db.Close()

	// Create handler object with pointer to database
	h := Handler{
		db,
	}

	// Register handlers with gorilla mux router
	router := mux.NewRouter()
	r := router.PathPrefix("/api/v1").Subrouter()
	registerHandlers(r, h)

	// Create and initialize middleware
	mw := Middleware{db}
	r.Use(mw.Authenticate)
	r.Use(mw.SetEncoder)
	r.Use(mw.Logger)

	// File server
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("dist")))

	//Create server - will use the port number given in variable "port"
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
