BEGIN;

INSERT INTO quotes (text, author, source, word_count,tags) VALUES
 ('Be yourself; everyone else is already taken.', 'Oscar Wilde', '', 7, '[]'::jsonb),
 ('I''m selfish, impatient and a little insecure. I make mistakes, I am out of control and at times hard to handle. But if you can''t handle me at my worst, then you sure as hell don''t deserve me at my best.', 'Marilyn Monroe', '', 41, '["attributed-no-source","best","life","love","misattributed-marilyn-monroe","mistakes","out-of-control","truth","worst"]'::jsonb),
 ('So many books, so little time.', 'Frank Zappa', '', 6, '["books","humor"]'::jsonb),
 ('Two things are infinite: the universe and human stupidity; and I''m not sure about the universe.', 'Albert Einstein', '', 16, '["attributed-no-source","human-nature","humor","infinity","philosophy","science","stupidity","universe"]'::jsonb),
 ('A room without books is like a body without a soul.', 'Marcus Tullius Cicero', '', 11, '["attributed-no-source","books","simile","soul"]'::jsonb),
 ('Don''t let the muggles get you down.', 'J.K. Rowling', 'Harry Potter and the Prisoner of Azkaban', 7, '[]'::jsonb),
 ('To be happy, we must not be too concerned with others.', 'Albert Camus', '', 11, '[]'::jsonb);COMMIT;
