CREATE TABLE IF NOT EXISTS answers (
id bigserial PRIMARY KEY,
question_id int NOT NULL REFERENCES questions(id),
answer text NOT NULL,
created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);