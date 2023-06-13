package db

// array of [2]string{SQL statement, context}
var setup = [][2]string{
	{sql_SETUP_quiz, "creating the quiz table"},
	{sql_SETUP_questions, "creating the questions table"},
	{sql_SETUP_users, "creating the users (ppl) table"},
}

// drop_question: an index from order, 0 based
const sql_SETUP_quiz = `CREATE TABLE IF NOT EXISTS quiz (
	id PRIMARY KEY,
	author TEXT REFERANCES ppl(id),
	deadname TEXT[] NOT NULL,
	deadlastname TEXT NOT NULL,
	chosenname TEXT[] NOT NULL,
	chosenlastname TEXT NOT NULL,
	nickname TEXT NOT NULL,
	order TEXT[] NOT NULL,
	drop_question SMALLINT NOT NULL,
	redirect TEXT NOT NULL
)`

// correct_answer: for multiple choice, 0 based index for answers
const sql_SETUP_questions = `CREATE TABLE IF NOT EXISTS questions (
	id TEXT PRIMARY KEY,
	quiz TEXT REFERANCES quiz(id) ON DELETE CASCADE,

	is_multiple_choice BOOL DEFAULT 'false',
	answers TEXT[] NOT NULL,
	correct_answer SMALLINT DEFAULT '0',
	content TEXT NOT NULL
)`

const sql_SETUP_users = `CREATE TABLE IF NOT EXISTS ppl (
	id TEXT PRIMARY KEY,
	token TEXT UNIQUE NOT NULL,
	username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL
)`