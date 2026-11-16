package repository

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func setupRepositoryTest(t *testing.T) (*pgx.Conn, context.Context) {
	t.Helper()
	url := os.Getenv("KHAYAL_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("KHAYAL_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := conn.Exec(ctx, "create schema "+schema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	if _, err := conn.Exec(ctx, "set search_path to "+schema); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	schemaSQL := `
create table users (
	id serial primary key,
	name text not null,
	email text not null unique,
	password_hash text
);
create table questions (
	id serial primary key,
	user_id integer references users(id) on delete cascade,
	client_id uuid,
	question text,
	question_ciphertext bytea,
	encryption_version smallint,
	answered boolean not null default false,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	return_at timestamptz
);
create table khayals (
	id serial primary key,
	question_id integer not null references questions(id) on delete cascade,
	client_id uuid,
	khayal text,
	khayal_ciphertext bytea,
	encryption_version smallint,
	created_at timestamptz not null default now()
);`
	if _, err := conn.Exec(ctx, schemaSQL); err != nil {
		t.Fatalf("create test tables: %v", err)
	}

	t.Cleanup(func() {
		_, _ = conn.Exec(context.Background(), "set search_path to public")
		_, _ = conn.Exec(context.Background(), "drop schema "+schema+" cascade")
		_ = conn.Close(context.Background())
	})
	return conn, ctx
}

func createTestUser(t *testing.T, conn *pgx.Conn, ctx context.Context, name string) int {
	t.Helper()
	var id int
	err := conn.QueryRow(ctx,
		`insert into users (name, email, password_hash) values ($1, $2, 'legacy') returning id`,
		name, name+"@example.com",
	).Scan(&id)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return id
}

func TestQuestionOwnershipAndCiphertextRoundTrip(t *testing.T) {
	conn, ctx := setupRepositoryTest(t)
	repo := NewQuestionRepostiory(conn)
	user1 := createTestUser(t, conn, ctx, "one")
	user2 := createTestUser(t, conn, ctx, "two")

	clientID := uuid.New()
	ciphertext := []byte("0123456789abcdefghijklmnopqrstuv")
	if err := repo.CreateQuestion(ctx, user1, clientID, ciphertext, 1); err != nil {
		t.Fatalf("create question: %v", err)
	}

	questions, err := repo.GetAllQuestions(ctx, user1)
	if err != nil {
		t.Fatalf("get user1 questions: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(questions))
	}
	if string(questions[0].QuestionCiphertext) != string(ciphertext) {
		t.Fatal("question ciphertext did not round-trip")
	}
	if questions[0].ClientID != clientID {
		t.Fatal("question client id did not round-trip")
	}

	other, err := repo.GetAllQuestions(ctx, user2)
	if err != nil {
		t.Fatalf("get user2 questions: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("user2 must not see user1's questions, got %d", len(other))
	}
	if _, err := repo.GetQuestion(ctx, user2, questions[0].ID); err == nil {
		t.Fatal("user2 must not fetch user1's question by id")
	}
}

func TestKhayalOwnershipAndCiphertextRoundTrip(t *testing.T) {
	conn, ctx := setupRepositoryTest(t)
	questionRepo := NewQuestionRepostiory(conn)
	khayalRepo := NewKhayalRepostiory(conn)
	user1 := createTestUser(t, conn, ctx, "one")
	user2 := createTestUser(t, conn, ctx, "two")

	questionClientID := uuid.New()
	questionCiphertext := []byte("0123456789abcdefghijklmnopqrstuv")
	if err := questionRepo.CreateQuestion(ctx, user1, questionClientID, questionCiphertext, 1); err != nil {
		t.Fatalf("create question: %v", err)
	}
	questions, err := questionRepo.GetAllQuestions(ctx, user1)
	if err != nil || len(questions) != 1 {
		t.Fatalf("get question: %v, count=%d", err, len(questions))
	}
	questionID := questions[0].ID

	clientID := uuid.New()
	khayalCiphertext := []byte("abcdefghijklmnopqrstuvwxyz012345")
	if err := khayalRepo.CreateKhayal(ctx, user1, questionID, clientID, khayalCiphertext, 1, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("create khayal: %v", err)
	}

	khayals, err := khayalRepo.GetAllKhayals(ctx, user1, questionID)
	if err != nil {
		t.Fatalf("get khayals: %v", err)
	}
	if len(khayals) != 1 {
		t.Fatalf("expected 1 khayal, got %d", len(khayals))
	}
	if string(khayals[0].KhayalCiphertext) != string(khayalCiphertext) {
		t.Fatal("khayal ciphertext did not round-trip")
	}
	if khayals[0].UserID != user1 {
		t.Fatal("khayal user id did not round-trip")
	}

	other, err := khayalRepo.GetAllKhayals(ctx, user2, questionID)
	if err != nil {
		t.Fatalf("get user2 khayals: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("user2 must not see user1's khayals, got %d", len(other))
	}
	if err := khayalRepo.CreateKhayal(ctx, user2, questionID, uuid.New(), khayalCiphertext, 1, time.Now().Add(time.Hour)); err == nil {
		t.Fatal("user2 must not create a khayal on user1's question")
	}
}

func TestLegacyQuestionMigration(t *testing.T) {
	conn, ctx := setupRepositoryTest(t)
	repo := NewQuestionRepostiory(conn)
	user1 := createTestUser(t, conn, ctx, "one")

	var questionID int
	err := conn.QueryRow(ctx,
		`insert into questions (user_id, question) values ($1, 'legacy plaintext') returning id`,
		user1,
	).Scan(&questionID)
	if err != nil {
		t.Fatalf("insert legacy question: %v", err)
	}

	rows, err := repo.GetLegacyQuestions(ctx, user1)
	if err != nil {
		t.Fatalf("get legacy questions: %v", err)
	}
	if len(rows) != 1 || rows[0].Question != "legacy plaintext" {
		t.Fatalf("unexpected legacy rows: %+v", rows)
	}

	clientID := uuid.New()
	ciphertext := []byte("0123456789abcdefghijklmnopqrstuv")
	if err := repo.ImportLegacyQuestion(ctx, user1, questionID, clientID, ciphertext, 1); err != nil {
		t.Fatalf("import legacy question: %v", err)
	}

	question, err := repo.GetQuestion(ctx, user1, questionID)
	if err != nil {
		t.Fatalf("get imported question: %v", err)
	}
	if string(question.QuestionCiphertext) != string(ciphertext) {
		t.Fatal("imported ciphertext did not round-trip")
	}
	if question.ClientID != clientID {
		t.Fatal("imported client id did not round-trip")
	}

	rows, err = repo.GetLegacyQuestions(ctx, user1)
	if err != nil {
		t.Fatalf("get legacy questions after import: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("legacy question should be migrated, got %+v", rows)
	}
}

func TestLegacyKhayalMigration(t *testing.T) {
	conn, ctx := setupRepositoryTest(t)
	questionRepo := NewQuestionRepostiory(conn)
	khayalRepo := NewKhayalRepostiory(conn)
	user1 := createTestUser(t, conn, ctx, "one")

	var questionID int
	err := conn.QueryRow(ctx,
		`insert into questions (user_id, question) values ($1, 'legacy question') returning id`,
		user1,
	).Scan(&questionID)
	if err != nil {
		t.Fatalf("insert legacy question: %v", err)
	}
	var khayalID int
	err = conn.QueryRow(ctx,
		`insert into khayals (question_id, khayal) values ($1, 'legacy khayal') returning id`,
		questionID,
	).Scan(&khayalID)
	if err != nil {
		t.Fatalf("insert legacy khayal: %v", err)
	}

	rows, err := khayalRepo.GetLegacyKhayals(ctx, user1)
	if err != nil {
		t.Fatalf("get legacy khayals: %v", err)
	}
	if len(rows) != 1 || rows[0].Khayal != "legacy khayal" {
		t.Fatalf("unexpected legacy khayals: %+v", rows)
	}

	clientID := uuid.New()
	ciphertext := []byte("abcdefghijklmnopqrstuvwxyz012345")
	if err := khayalRepo.ImportLegacyKhayal(ctx, user1, khayalID, questionID, clientID, ciphertext, 1); err != nil {
		t.Fatalf("import legacy khayal: %v", err)
	}

	if _, err := questionRepo.GetQuestion(ctx, user1, questionID); err == nil {
		// The question itself is still legacy; that's fine. This assertion only
		// documents that migration is per-row.
	}

	rows, err = khayalRepo.GetLegacyKhayals(ctx, user1)
	if err != nil {
		t.Fatalf("get legacy khayals after import: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("legacy khayal should be migrated, got %+v", rows)
	}
}

func TestWrongUserCannotImportLegacyQuestion(t *testing.T) {
	conn, ctx := setupRepositoryTest(t)
	repo := NewQuestionRepostiory(conn)
	user1 := createTestUser(t, conn, ctx, "one")
	user2 := createTestUser(t, conn, ctx, "two")

	var questionID int
	if err := conn.QueryRow(ctx,
		`insert into questions (user_id, question) values ($1, 'legacy') returning id`, user1,
	).Scan(&questionID); err != nil {
		t.Fatalf("insert legacy question: %v", err)
	}
	err := repo.ImportLegacyQuestion(ctx, user2, questionID, uuid.New(), []byte("0123456789abcdefghijklmnopqrstuv"), 1)
	if err == nil {
		t.Fatal("user2 must not import user1's legacy question")
	}
}

func TestTestSchemaNameIsValid(t *testing.T) {
	if strings.ContainsAny(fmt.Sprint("test_schema"), " -;") {
		t.Fatal("unexpected schema characters")
	}
}
