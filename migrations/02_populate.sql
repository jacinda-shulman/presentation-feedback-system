-- +goose Up
-- SQL in this section is executed when the migration is applied.

--COPY question(q_type, q_num, q_text) FROM 'questions.txt' DELIMITER ',';

INSERT INTO question(q_type, q_num, q_text) VALUES
('M/C',1,'Preparedness: the presenter was adequately prepared.'),
('M/C',2,'Organization: the presentation material was arranged logically.'),
('M/C',3,'Correctness: the presented facts were correct (to the best of your knowledge).'),
('M/C',4,'Visualization: the visual material included appropriate content/fonts/graphics.'),
('M/C',5,'General introduction: the presentation clearly introduced the broad area containing the topic.'),
('M/C',6,'Motivation: the presentation clearly motivated the specific topic in the context of the broad area.'),
('M/C',7,'Introduction: the presentation clearly introduced the specific topic.'),
('M/C',8,'Tutorial/demonstration: the tutorial/demonstration improved your understanding of the specific topic.'),
('M/C',9,'Multiple-choice questions: at least three multiple-choice questions assessed your understanding of the presented content.'),
('M/C',10,'Answers: the presenter''s answers to questions were satisfying.'),
('Open',1,'Provide any comments for the presenter.'),
('Open',2,'Provide any comments for your instructor.');

INSERT INTO account(token, first_name, last_name) VALUES
('aaaa123', 'TestF', 'TestL'),
('0000111', 'Bob', 'Baker'),
('111kkkk', 'Alice', 'Ashcroft'),
('abcdefg', 'Claire', 'Cooper'),
('vaMPir3', 'Edward', 'Cullen');

INSERT INTO presentation(presenter_id, title, slot_date, slot_time) VALUES
(1, 'The Best Presentation Ever', 'March 28', '9:30am'),
(2, 'Bobs Brilliant Boasting', 'March 28', '11:00am'),
(3, 'Web Dev Rocks','April 2', '9:30am'),
(4, 'Security for Dummies','April 4', '9:30am');

INSERT INTO form(presenter_id, evaluator_id) VALUES
(3, 2), -- Alice presents; Bob evaluates
(2, 4); -- Bob presents, Claire evaluates

INSERT INTO answer(form_id, q_id, a_value) VALUES
-- Bob evaluates Alice - gives 4 for everything
(1, 1, 4), (1, 2, 4),(1, 3, 4),(1, 4, 4),(1, 5, 4),
(1, 6, 4),(1, 7, 4),(1, 8, 4),(1, 9, 4),(1, 10, 4),
(1, 11, 'She''s ok'),(1, 12, 'You''re ok too'),
-- Claire evaluates Bob, she doesn't fill everything out
(2, 1, -1), (2, 2, -1),(2, 3, 4),(2, 4, -1),(2, 5, 4),
(2, 6, 4),(2, 7, 4),(2, 8, 4),(2, 9, 4),(2, 10, 4),
(2, 11, 'Bob is abismal'),(2, 12, '');

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DELETE FROM answer;
DELETE FROM form;
DELETE FROM question;
DELETE FROM presentation;
DELETE FROM account;

