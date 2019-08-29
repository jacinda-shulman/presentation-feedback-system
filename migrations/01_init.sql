-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TYPE question_type as enum('M/C', 'Open');

CREATE TABLE question(
	q_id	SERIAL,
	q_type	question_type NOT NULL,	
	q_num	int		NOT NULL,
	q_text	text	NOT NULL,
	UNIQUE(q_id),
	PRIMARY KEY(q_id)
);

CREATE TABLE account(
	account_id		SERIAL,
	token			text	NOT NULL,
	first_name		text	NOT NULL,
	last_name		text	NOT NULL,
	PRIMARY KEY(account_id)
);

CREATE TABLE presentation(
	presenter_id	int 	REFERENCES account(account_id) NOT NULL,
	title		text	NOT NULL,
	slot_date	text,
	slot_time	text,
	PRIMARY KEY(presenter_id) --Each presenter can only have one presentation
);

CREATE TABLE form(
	form_id		SERIAL UNIQUE,
	presenter_id	int	NOT NULL	REFERENCES presentation(presenter_id),
	evaluator_id	int	NOT NULL	REFERENCES account(account_id),
	CHECK(presenter_id != evaluator_id),
	PRIMARY KEY(presenter_id, evaluator_id)
);

CREATE TABLE answer(
	answer_id	SERIAL,	
	form_id		int	NOT NULL REFERENCES form(form_id),
	q_id		int	NOT NULL REFERENCES question(q_id),
	a_value		text,
	PRIMARY KEY(answer_id)
);


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE answer;
DROP TABLE question;
DROP TABLE form;
DROP TABLE presentation;
DROP TABLE account;
DROP TYPE question_type;