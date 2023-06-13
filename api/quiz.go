package api

import (
	"github.com/shadiestgoat/who/db"
	"github.com/shadiestgoat/who/snownode"
)

type Quiz struct {
	ID       string
	AuthorID string

	DeadNames    []string
	DeadLastName string

	ChosenNames    []string
	ChosenLastName string

	Nickname string

	Order        []string
	DropQuestion int

	Redirect string
}

// Sanitizes the quiz for step 1 (ie. creation). Does not sanitize or verify ID, AuthorID
func (q *Quiz) Sanitize1() error {
	var err error

	if err = cleanString(&q.DeadLastName, 2, 33, "dead Last Name"); err != nil {
		return err
	}
	if err = cleanString(&q.ChosenLastName, -1, 33, "chosen Last Name"); err != nil {
		return err
	}
	if err = cleanString(&q.Nickname, 2, 33, "nickname"); err != nil {
		return err
	}
	if err = cleanString(&q.DeadLastName, 0, 33, "redirect"); err != nil {
		return err
	}

	if q.DeadNames, err = cleanStringArr(q.DeadNames, 2, 33, "dead Name"); err != nil {
		return err
	}
	if q.ChosenNames, err = cleanStringArr(q.ChosenNames, 2, 33, "chosen Name"); err != nil {
		return err
	}

	if len(q.DeadNames) == 0 || len(q.DeadNames) > 4 {
		return &HTTPError{
			Msg:    "Need 1-4 dead names",
			Status: 400,
		}
	}
	if len(q.ChosenNames) == 0 || len(q.ChosenNames) > 4 {
		return &HTTPError{
			Msg:    "Need 1-4 chosen names",
			Status: 400,
		}
	}

	if q.DropQuestion > 2 || q.DropQuestion < 0 {
		return &HTTPError{
			Msg:    "Drop question out of bounds",
			Status: 400,
		}
	}

	if q.ChosenLastName == "" {
		q.ChosenLastName = q.DeadLastName
	}

	return nil
}

func NewQuiz(q *Quiz, rqs []*Question) (*Quiz, error) {
	if err := q.Sanitize1(); err != nil {
		return nil, err
	}

	if len(rqs) != 3 {
		return nil, &HTTPError{
			Msg: "Need 3 questions",
			Status: 400,
		}
	}

	q.ID = snownode.Generate()

	// `id`, `quiz`,
	// `is_multiple_choice`,
	// `answers`,
	// `content`,
	questions := [][]any{}

	for _, question := range rqs {
		if err := question.Sanitize(); err != nil {
			return nil, err
		}
		question.ID = snownode.Generate()
		questions = append(questions, []any{
			question, q.ID,
			question.IsMultipleChoice,
			question.Answers,
			question.Content,
		})
	}

	db.InsertOne(`quiz`, []string{
		`id`, `author`,
		`deadname`, `deadlastname`,
		`chosenname`, `chosenlastname`,
		`nickname`,
		`order`, `drop_question`,
		`redirect`,
	},
		q.ID, q.AuthorID,
		q.DeadNames, q.DeadLastName,
		q.ChosenNames, q.ChosenLastName,
		q.Nickname,
		q.Order, q.DropQuestion,
		q.Redirect,
	)

	db.Insert(`questions`, []string{
		`id`, `quiz`,
		`is_multiple_choice`,
		`answers`,
		`content`,
	}, questions)

	return q, nil
}

// Note: use with POST, it overrides everything!
func EditQuiz(q *Quiz) (*Quiz, error) {
	if err := q.Sanitize1(); err != nil {
		return nil, err
	}

	_, err := db.Exec(`UPDATE quiz SET deadname = $1, deadlastname = $2, chosenname = $3, chosenlastname = $4, nickname = $5, order = $6, drop_question = $7, redirect = $8 WHERE id = $9`,
		q.DeadNames, q.DeadLastName, q.ChosenNames, q.ChosenLastName, q.Nickname, q.Order, q.DropQuestion, q.Redirect, q.ID,
	)

	if err != nil {
		return nil, ErrDBHandle(err)
	}

	return q, nil
}

func DeleteQuiz(id string) (*Quiz, error) {
	q, err := GetQuiz(id)
	if err != nil {
		return nil, err
	}
	
	_, err = db.Exec(`DELETE FROM quiz WHERE id = $1`, id)
	
	if err != nil {
		return nil, ErrServerErr
	}

	return q, nil
}

func GetQuiz(id string) (*Quiz, error) {
	q := &Quiz{
		ID:             id,
		AuthorID:       "",
		DeadNames:      []string{},
		DeadLastName:   "",
		ChosenNames:    []string{},
		ChosenLastName: "",
		Nickname:       "",
		Order:          []string{},
		DropQuestion:   0,
		Redirect:       "",
	}

	err := db.QueryRowID(
		`SELECT author, deadname, deadlastname, chosenname, chosenlastname, nickname, order, drop_question, redirect WHERE id = $1`, 
		id,
		&q.AuthorID, &q.DeadNames, &q.DeadLastName, &q.ChosenNames, &q.ChosenLastName, &q.Nickname, &q.Order, &q.DropQuestion, &q.Redirect, 
	)

	if err != nil {
		return nil, ErrDBHandle(err)
	}

	return q, nil
}

func GetQuizFirstQuestion(id string) (*Question, error) {
	return GetQuestionUsingPosition(1, 1, id)
}