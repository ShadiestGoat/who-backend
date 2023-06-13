package api

import (
	"fmt"
	"strings"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/who/db"
)

// Notice about question IDs:
// There are 2 types of questions: user created ones (the manual ones) and the special ones (automatic)
//
// Normal questions are the ones created by the user, ie. the ones that go into the questions table
// Their ID is generated using {id}{section:1|2|3}
//
// There are 2 special questions per quiz: 3d question of section 2 & 3.
// 3-2 (q3s2): What is another name for {nickname}
// Accepted answers are {deadname}, {deadname} {deadlastname}, {chosenname} {chosenlastname}
// (note: {deadname} and {chosenname} are arrays, combinations apply to all items in the array)
// If {chosenname} is in the answer, the response should be the 'redirect', otherwise go to 1-3
// 3-3: What is another name for {chosenname}
// Accepted answers are {deadname}, {deadname} {deadlastname}, {nickname}
// Special ID format: sp-{2|3}-{quizID}
// (2 = "3-2", 3 = "3-3")

type Question struct {
	ID string `json:"id"`

	Content          string   `json:"content"`
	IsMultipleChoice bool     `json:"isMultipleChoice"`
	Answers          []string `json:"answers,omitempty"`
}

type FullQuestion struct {
	Question

	CorrectAnswer int `json:"correctAnswer"`
}

func (q *Question) Sanitize() error {
	var err error

	if err = cleanString(&q.Content, 2, 65, "questions.content"); err != nil {
		return err
	}
	if q.Answers, err = cleanStringArr(q.Answers, 2, 33, "questions.answers"); err != nil {
		return err
	}

	answers := map[string]bool{}

	newAns := []string{}

	for _, ans := range q.Answers {
		ans = strings.ToLower(ans)

		if answers[ans] {
			continue
		}
		answers[ans] = true

		newAns = append(newAns, ans)
	}

	q.Answers = newAns

	if len(q.Answers) == 0 || len(q.Answers) > 4 {
		return &HTTPError{
			Msg:    "Too many answers",
			Status: 400,
		}
	}

	return nil
}

func (q *FullQuestion) Sanitize() error {
	if err := q.Question.Sanitize(); err != nil {
		return err
	}

	if q.CorrectAnswer >= len(q.Answers) || q.CorrectAnswer < 0 {
		return ErrBadBody
	}

	return nil
}

func genSpecialQuestion(ogID string) (*Question, error) {
	id := ogID[3:]

	specialTime := id[0]
	quizID := id[2:]

	switch specialTime {
	case '2':
		nickname := ""

		err := db.QueryRowID(`SELECT nickname FROM quiz WHERE id = $1`, quizID, &nickname)
		if err != nil {
			if db.NoRows(err) {
				return nil, ErrNotFound
			}

			return nil, ErrServerErr
		}

		return &Question{
			ID:      ogID,
			Content: "What is another name for " + Capitalize(nickname) + "?",
		}, nil
	case '3':
		chosenName := ""

		err := db.QueryRowID(`SELECT chosenname[1] FROM quiz WHERE id = $1`, quizID, &chosenName)
		if err != nil {
			if db.NoRows(err) {
				return nil, ErrNotFound
			}

			return nil, ErrServerErr
		}

		return &Question{
			ID:      ogID,
			Content: "Who the fuck is " + Capitalize(chosenName) + "??",
		}, nil
	}

	return nil, ErrNotFound
}

// Get a question based of off it's position in the quiz, section and question being from 1-3 (inclusive)
func GetQuestionUsingPosition(section, question int, quizID string) (*Question, error) {
	if question == 3 && section != 1 {
		return genSpecialQuestion("sp-" + fmt.Sprint(section) + "-" + quizID)
	}

	qID := ""
	if section == 1 {

		err := db.QueryRowID(`SELECT order[1] FROM quiz WHERE id = $1`, quizID, &qID)

		if err != nil {
			return nil, ErrDBHandle(err)
		}
	} else {
		order := []string{}
		exception := 0

		err := db.QueryRowID(`SELECT order, drop_question FROM quiz WHERE id = $1`, quizID, &order, &exception)

		if err != nil {
			return nil, ErrDBHandle(err)
		}

		if exception == 0 {
			qID = order[1]
		} else {
			qID = order[0]
		}
	}

	return GetQuestion(qID + fmt.Sprint(section))
}

func GetQuestion(id string) (*Question, error) {
	if strings.HasPrefix(id, "sp-") {
		return genSpecialQuestion(id)
	}

	if id == "" {
		return nil, ErrNotFound
	}

	q := &Question{
		ID:               id,
		Content:          "",
		IsMultipleChoice: false,
		Answers:          []string{},
	}

	quiz := ""

	err := db.QueryRowID(
		`SELECT is_multiple_choice, answers, content, quiz FROM questions WHERE id = $1`,
		id[:len(id)-1],
		&q.IsMultipleChoice,
		&q.Answers,
		&q.Content,
		&quiz,
	)

	if err != nil {
		if db.NoRows(err) {
			return nil, ErrNotFound
		}

		return nil, ErrServerErr
	}

	if !q.IsMultipleChoice {
		q.Answers = nil
	}

	name := ""

	switch id[len(id)-1] {
	case '1':
		// deadname
		db.QueryRowID(`SELECT deadname[1] FROM quiz WHERE id = $1`, quiz, &name)
	case '2':
		// nickname
		db.QueryRowID(`SELECT nickname FROM quiz WHERE id = $1`, quiz, &name)
	case '3':
		// chosenname
		db.QueryRowID(`SELECT chosenname[1] FROM quiz WHERE id = $1`, quiz, &name)
	}

	q.Content = strings.ReplaceAll(q.Content, "{{name}}", name)

	return q, nil
}

// Admin only!
func GetQuestions(quiz string) ([3]*FullQuestion, error) {
	rows, err := db.Query(`SELECT id, is_multiple_choice, answers, correct_answer, content FROM questions WHERE quiz = $1 LIMIT 3`, quiz)

	if err != nil {
		return [3]*FullQuestion{}, ErrDBHandle(err)
	}

	q := [3]*FullQuestion{}

	i := 0

	defer rows.Close()

	for rows.Next() {
		tmpQ := &FullQuestion{
			Question: Question{
				ID:               "",
				Content:          "",
				IsMultipleChoice: false,
				Answers:          []string{},
			},
			CorrectAnswer: 0,
		}

		q[i] = tmpQ

		err := rows.Scan(&tmpQ.ID, &tmpQ.IsMultipleChoice, &tmpQ.Answers, &tmpQ.CorrectAnswer, &tmpQ.Content)

		if err != nil {
			return [3]*FullQuestion{}, ErrDBHandle(err)
		}

		i++
	}

	return q, nil
}

func EditQuestion(q *FullQuestion) (*FullQuestion, error) {
	if err := q.Sanitize(); err != nil {
		return nil, err
	}

	_, err := db.Exec(
		`UPDATE questions SET is_multiple_choice = $1, answers = $2, correct_answer = $3, content = $4 WHERE id = $5`,
		q.IsMultipleChoice, q.Answers, q.CorrectAnswer, q.Content, q.ID,
	)

	if err != nil {
		return nil, ErrDBHandle(err)
	}

	return q, nil
}

type QuestionResp struct {
	Correct bool      `json:"correct"`
	Next    *Question `json:"next,omitempty"`

	Redirect string `json:"redirect,omitempty"`
}

func genGoodQuestionResp(q *Question, err error) (*QuestionResp, error) {
	if err != nil {
		return nil, err
	}

	return &QuestionResp{
		Correct: true,
		Next:    q,
	}, nil
}

func AnswerQuestion(id string, answer string) (*QuestionResp, error) {
	answer = strings.ToLower(answer)

	if strings.HasPrefix(id, "sp-") {
		return answerSpecial(id, answer)
	}

	if id == "" {
		return nil, ErrNotFound
	}

	section := id[len(id)-1]

	if section < '1' || section > '3' {
		return nil, ErrNotFound
	}

	qID := id[:len(id)-1]

	multipleChoice := false
	answers := []string{}
	correctAnswer := 0
	order := []string{}
	dropQuestion := 0
	quizID := ""

	err := db.QueryRowID(
		`SELECT is_multiple_choice, answers, correct_answer, quiz.order, drop_question, quiz.id FROM questions JOIN quiz ON questions.quiz = quiz.id WHERE questions.id = $1`,
		qID,
		&multipleChoice, &answers, &correctAnswer, &order, &dropQuestion, &quizID,
	)
	if err != nil {
		return nil, ErrDBHandle(err)
	}

	isCorrect := false

	if multipleChoice {
		if correctAnswer >= len(answers) || correctAnswer < 0 {
			isCorrect = true
		} else {
			isCorrect = answer == answers[correctAnswer]
		}
	} else {
		for _, a := range answers {
			if a == answer {
				isCorrect = true
				break
			}
		}
	}

	if !isCorrect {
		return &QuestionResp{
			Correct: false,
		}, nil
	}

	if section != '1' {
		tmpOrder := []string{}

		for i, o := range order {
			if i == dropQuestion {
				continue
			}

			tmpOrder = append(tmpOrder, o)
		}

		order = tmpOrder
	}

	questionIndex := -1

	for i, o := range order {
		if o == qID {
			questionIndex = i
			break
		}
	}

	if questionIndex == -1 {
		log.Warn("Crazy stuff is happening mannn")
		return nil, ErrServerErr
	}

	// last question of section 1
	if section == '1' && questionIndex == 2 {
		return genGoodQuestionResp(GetQuestionUsingPosition(2, 1, quizID))
	}

	return genGoodQuestionResp(GetQuestionUsingPosition(int(section-'0'), (questionIndex+1)+1, quizID))
}

func answerSpecial(id string, answer string) (*QuestionResp, error) {
	id = id[3:]
	specialID := id[0]
	quizID := id[2:]

	// Answer -> 1|2 (!ok -> bad answer)
	// 1 -> lead to section 3
	// 2 -> lead to redirect
	m := map[string]int{}
	redirect := ""

	switch specialID {
	case '2':
		// {deadname}
		// {deadname} {deadlastname}
		// {chosenname}
		// {chosenname} {chosenlastname}

		deadNames := []string{}
		deadLastName := ""

		chosenNames := []string{}
		chosenLastName := ""

		err := db.QueryRowID(
			`SELECT deadname, deadlastname, chosenname, chosenlastname, redirect FROM quiz WHERE id = $1`,
			quizID,

			&deadNames,
			&deadLastName,

			&chosenNames,
			&chosenLastName,

			&redirect,
		)

		if err != nil {
			if db.NoRows(err) {
				return nil, ErrNotFound
			}

			return nil, ErrServerErr
		}

		for _, n := range deadNames {
			n = strings.ToLower(n)

			m[n] = 1
			m[n+" "+deadLastName] = 1
		}

		for _, n := range chosenNames {
			n = strings.ToLower(n)

			m[n] = 2
			m[n+" "+chosenLastName] = 2
		}
	case '3':
		// {deadname}
		// {deadname} {deadlastname}
		// {nickname}

		deadNames := []string{}
		deadLastName := ""
		nickname := ""

		err := db.QueryRowID(
			`SELECT deadname, deadlastname, nickname, redirect FROM quiz WHERE id = $1`,
			quizID,

			&deadNames,
			&deadLastName,

			&nickname,
			&redirect,
		)

		if err != nil {
			if db.NoRows(err) {
				return nil, ErrNotFound
			}

			return nil, ErrServerErr
		}

		for _, n := range deadNames {
			n = strings.ToLower(n)

			m[n] = 2
			m[n+" "+deadLastName] = 2
		}

		m[nickname] = 2
	}

	if resp, ok := m[answer]; ok {
		if resp == 1 {
			return genGoodQuestionResp(GetQuestionUsingPosition(3, 1, quizID))
		} else {
			return &QuestionResp{
				Correct:  true,
				Redirect: redirect,
			}, nil
		}
	}

	return &QuestionResp{
		Correct: false,
	}, nil
}
