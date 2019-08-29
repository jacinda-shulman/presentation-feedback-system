/*
CMPT 315 (Winter 2019)
Assign. 1: Presentation Feedback System (Backend)
Author: Jacinda Shulman

This file implements the data access functions
*/

package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var NumQuestions int

// Database wraps a sqlx.DB type
type Database struct {
	*sqlx.DB
}

// Question holds data representing a question
type Question struct {
	XMLName struct{} `json:"-" xml:"question"`
	ID      int      `db:"q_id" json:"qID" xml:"qID"`
	QType   string   `db:"q_type" json:"qType" xml:"qType"`
	QNum    int      `db:"q_num" json:"qNum" xml:"qNum"`
	QText   string   `db:"q_text" json:"qText" xml:"qText"`
}

// Account holds data representing a user account
type Account struct {
	XMLName   struct{} `json:"-" xml:"account"`
	ID        int      `db:"account_id" json:"accountID" xml:"accountID"`
	Token     int      `db:"token" json:"token" xml:"token"`
	FirstName string   `db:"first_name" json:"firstName" xml:"firstName"`
	LastName  string   `db:"last_name" json:"lastName" xml:"lastName"`
}

// SimpleUser holds the first name and id of a user
type SimpleUser struct {
	XMLName   struct{} `json:"-" xml:"account"`
	FirstName string   `db:"first_name" json:"firstName" xml:"firstName"`
	ID        int      `db:"account_id" json:"accountID" xml:"accountID"`
}

// Presentation holds details of a presentation
type Presentation struct {
	XMLName     struct{} `json:"-" xml:"presentation"`
	PresenterID int      `db:"presenter_id" json:"presenterID" xml:"presenterID"`
	Title       string   `db:"title" json:"title" xml:"title"`
	Date        string   `db:"slot_date" json:"slotDate" xml:"slotDate"`
	Time        string   `db:"slot_time" json:"slotTime" xml:"slotTime"`
}

// Presenter contains a presenter's name and presentation title
type Presenter struct {
	XMLName   struct{} `json:"-" xml:"presenter"`
	ID        int      `db:"account_id" json:"accountID" xml:"accountID"`
	FirstName string   `db:"first_name" json:"firstName" xml:"firstName"`
	LastName  string   `db:"last_name" json:"lastName" xml:"lastName"`
	Title     string   `db:"title" json:"title" xml:"title"`
}

// Form contains the IDs associated with a feedback form submission
type Form struct {
	XMLName     struct{} `json:"-" xml:"form"`
	ID          int      `db:"form_id" json:"formID" xml:"formID"`
	PresenterID int      `db:"presenter_id" json:"presenterID" xml:"presenterID"`
	EvaluatorID int      `db:"evaluator_id" json:"evaluatorID" xml:"evaluatorID"`
}

// WrappedForm is a struct containing a form struct and the
// answerIDs that go with it
type WrappedForm struct {
	XMLName   struct{} `json:"-" xml:"wrappedForm"`
	Form      Form     `json:"form" xml:"form"`
	AnswerIDs []int    `json:"answerIDs" xml:"answerIDs"`
}

// Answer represents a response to a question on a feedback form
type Answer struct {
	XMLName  struct{} `json:"-" xml:"answer"`
	AnswerID int      `db:"answer_id" json:"answerID" xml:"answerID"`
	FormID   int      `db:"form_id" json:"formID" xml:"formID"`
	QID      int      `db:"q_id" json:"qID" xml:"qID"`
	Value    string   `db:"a_value" json:"answerValue" xml:"answerValue"`
}

// GetQuestionSet obtains the set of questions for the feedback form
// - Returns a slice of Question structs
func (db *Database) GetQuestionSet() ([]Question, error) {
	var results []Question

	q := `SELECT * FROM question`
	if err := db.Select(&results, q); err != nil {
		log.Printf("-- error getting from database: %s\n", q)
		return []Question{}, err
	}
	return results, nil
}

// GetPresenters obtains the set of presenters (all details)
// - Returns a slice of Presenters
func (db *Database) GetPresenters() (results []Presenter, err error) {
	q := `SELECT account_id, first_name, last_name, title
			FROM account, presentation
			WHERE account_id = presenter_id
			ORDER BY first_name ASC`
	if err = db.Select(&results, q); err != nil {
		log.Printf("-- error getting presenters from database: %s\n", q)
		return
	}
	return
}

// GetPresentations obtains the set of presentations
// - Returns a slice of Presenters
func (db *Database) GetPresentations() (results []Presentation, err error) {
	q := `SELECT * FROM presentation ORDER BY title ASC`
	if err = db.Select(&results, q); err != nil {
		log.Printf("-- error getting presenters from database: %s\n", q)
		return
	}
	return
}

// CreateForm inserts a new form and a set of answers into the database
// It returns a Form struct with the new form's details and a slice of ints
//	containing the set of answerIDs associated with that form
func (db *Database) CreateForm(presenterID, evaluatorID int) (form Form, listIDs []int, status int) {
	status = 500 // Will be updated before return if function is successful

	// Check if there is a form with the IDs provided
	q1 := `SELECT * FROM form WHERE presenter_id = $1 and evaluator_id = $2`
	err := db.Get(&form, q1, presenterID, evaluatorID)

	// The form already exists, set status to 409 Conflict
	if err == nil {
		status = 409 // Conflict
		// get the range of answerIDs for the relevant form
		if listIDs, err = db.GetAnswerIDs(form.ID); err != nil {
			status = 500
			return
		}
	} else {
		// if form doesn't exist, insert new one
		q2 := `INSERT INTO form(presenter_id, evaluator_id) VALUES($1, $2)`
		if _, err := db.Exec(q2, presenterID, evaluatorID); err != nil {
			// Error here means that presenterID wasn't found
			//	(evaluatorID is validated in middleware)
			status = 400 // Bad Request
			return
		}
		// get the form we just created
		if err := db.Get(&form, q1, presenterID, evaluatorID); err != nil {
			return
		}
		// initialize answers in answer table
		q2 = `INSERT INTO answer(form_id, q_id, a_value) 
						VALUES($1, $2, -1)`
		q3 := `INSERT INTO answer(form_id, q_id, a_value) 
				VALUES($1, $2, '')`
		for i := 1; i <= NumQuestions; i++ {
			if i <= 10 {
				if _, err = db.Exec(q2, form.ID, i); err != nil {
					fmt.Printf("-- error 4 in CreateForm, error: %v\n", err)
					return
				}
			} else {
				if _, err = db.Exec(q3, form.ID, i); err != nil {
					fmt.Printf("-- error 4 in CreateForm, error: %v\n", err)
					return
				}
			}
		}

		// get the list of answerIDs for the new form
		if listIDs, err = db.GetAnswerIDs(form.ID); err != nil {
			status = 500
			return
		}
		status = 201 // Created
	}
	return form, listIDs, status
}

// ClearForm deletes all answers with the formID
//	The accountID given must match the evaluatorID for the form
func (db *Database) ClearForm(formID, accountID int) (status int, err error) {
	status = 200
	// Try to get the evaluator of the form given
	var evaluatorID int
	if evaluatorID, err = db.GetEvaluator(formID); err != nil {
		status = 400 // Bad Request - form doesn't exist
		return
	}
	// If the accountID and evaluatorID don't match, send 403 Forbidden
	if evaluatorID != accountID {
		status = 403 // Forbidden - can't clear this form
		return
	}
	// If authorized, clear form (set all answers to null)
	q := `UPDATE answer
			SET a_value = NULL
			WHERE form_id = $1`
	if _, err = db.Exec(q, formID); err != nil {
		fmt.Printf("-- error 4 in ClearForm, error: %v\n", err)
		status = 500
	}
	return
}

// DeleteForm deletes all the answers and the form entry for the given formID
//	The accountID given must match the evaluatorID for the form
func (db *Database) DeleteForm(formID, accountID int) (status int, err error) {
	status = 200
	// Try to get the evaluator of the form given
	var evaluatorID int
	if evaluatorID, err = db.GetEvaluator(formID); err != nil {
		status = 400 // Bad Request - form doesn't exist
		return
	}
	// If the accountID and evaluatorID don't match, send 403 Forbidden
	if evaluatorID != accountID {
		status = 403 // Forbidden - can't delete this form
		return
	}
	// If authorized, delete answers and delete form
	q := `DELETE FROM answer
			WHERE form_id = $1`
	if _, err = db.Exec(q, formID); err != nil {
		fmt.Printf("-- error in DeleteForm, error: %v\n", err)
		status = 500
	}
	q = `DELETE FROM form
			WHERE form_id = $1`
	if _, err = db.Exec(q, formID); err != nil {
		fmt.Printf("-- error in DeleteForm, error: %v\n", err)
		status = 500
	}
	return
}

// UpdateAnswer changes the answer
// Parameters: answerID, answer entered (as a string), accountID calling
//	The accountID given must match the evaluatorID for the form
func (db *Database) UpdateAnswer(answerID int, answerValue string, accountID int) (status int, err error) {
	var (
		formID, evaluatorID int
	)
	status = 200
	if formID, err = db.FormFromAnswer(answerID); err != nil {
		status = 400 // Bad Request - form or question doesn't exist
		err = fmt.Errorf("Error: form and/or answer don't exist")
		return
	}
	// If the accountID and evaluatorID don't match, send 403 Forbidden
	evaluatorID, _ = db.GetEvaluator(formID)
	if evaluatorID != accountID {
		status = 403 // Forbidden - can't edit this form
		return
	}
	// If authorized update the answer
	q := `UPDATE answer
			SET a_value = $1
			WHERE answer_id = $2`
	if _, err = db.Exec(q, answerValue, answerID); err != nil {
		fmt.Printf("-- error 4 in ClearForm, error: %v\n", err)
		status = 500
	}
	return
}

// GetAnswers returns a slice of Answers for the given form
func (db *Database) GetAnswers(formID int) (list []Answer, status int, err error) {
	status = 200
	q := `SELECT * FROM answer WHERE form_id = $1`
	if err = db.Select(&list, q, formID); err != nil {
		log.Printf("-- error getting answers from database: %v\n", err)
		status = 400 // Bad Request - the form doesn't exist
		return
	}
	return list, status, nil
}

// FormFromAnswer returns a slice of integers containing the ids of
//	all the answers associated with the given form
func (db *Database) FormFromAnswer(answerID int) (list int, err error) {
	q := `SELECT form_id FROM answer WHERE answer_id = $1`
	if err = db.Get(&list, q, answerID); err != nil {
		log.Printf("-- error getting formID from database: %s\n", q)
		return
	}
	return list, nil
}

// AnswerFromForm takes a formID and the question id and returns
//	the answerID if it exists, nil and error if not
func (db *Database) AnswerFromForm(formID, qID int) (list int, err error) {
	q := `SELECT answer_id FROM answer 
			WHERE form_id = $1 and q_id = $2`
	if err = db.Get(&list, q, formID, qID); err != nil {
		log.Printf("-- error getting answerID from database: %s\n", q)
		return
	}
	return list, nil
}

// GetAnswerIDs returns a slice of integers containing the ids of
//	all the answers associated with the given form
func (db *Database) GetAnswerIDs(formID int) (list []int, err error) {
	q := `SELECT answer_id FROM answer WHERE form_id = $1`
	if err = db.Select(&list, q, formID); err != nil {
		log.Printf("-- error getting answerIDs from database: %v\n", err)
		return
	}
	return list, nil
}

// GetEvaluator returns the evaluatorID of a form
func (db *Database) GetEvaluator(formID int) (evaluatorID int, err error) {
	q := `SELECT evaluator_id FROM form WHERE form_id = $1`
	if err := db.Get(&evaluatorID, q, formID); err != nil {
		fmt.Printf("error in GetEvaluator: %v\n", err)
		return evaluatorID, err
	}
	return evaluatorID, nil
}

// UserFromToken returns the user who has the token given,
// If the token isn't found, function returns nil and error.
func (db *Database) UserFromToken(token string) (user Account, err error) {
	// Search database for the token
	q := `SELECT * FROM account WHERE token = $1`
	if err = db.Get(&user, q, token); err != nil {
		fmt.Println("couldn't find token")
		return user, err
	}
	return user, nil
}

// NameFromID returns the user's first and last name from the userID
// If the ID isn't found, function returns nil and error.
func (db *Database) NameFromID(id int) (firstName string, err error) {
	// Search database for the token
	q := `SELECT first_name FROM account WHERE account_id = $1`
	if err = db.Get(&firstName, q, id); err != nil {
		fmt.Println("couldn't find id")
		return firstName, err
	}
	return firstName, nil
}

// Source: database.go from source code given in class (lab03 & lab04)
func connectToDB(connectStr string) (*Database, error) {
	db := Database{}
	var err error

	db.DB, err = sqlx.Connect("postgres", connectStr)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to database: %v", err)
	}
	NumQuestions = 12
	return &db, nil
}
