package controller

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dnp314/khayal/internal/entities"
	"github.com/dnp314/khayal/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeQuestionService struct {
	createdUserID     int
	createdClientID   uuid.UUID
	createdCiphertext []byte
	questions         []entities.Question
}

func (f *fakeQuestionService) GetQuestion(ctx context.Context, userID, id int) (entities.Question, error) {
	return entities.Question{ID: id, UserID: userID}, nil
}

func (f *fakeQuestionService) GetAllQuestions(ctx context.Context, userID int) ([]entities.Question, error) {
	return f.questions, nil
}

func (f *fakeQuestionService) CreateQuestion(ctx context.Context, userID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error {
	f.createdUserID = userID
	f.createdClientID = clientID
	f.createdCiphertext = ciphertext
	return nil
}

type fakeKhayalService struct {
	createdUserID     int
	createdQuestionID int
	createdClientID   uuid.UUID
	createdCiphertext []byte
}

func (f *fakeKhayalService) GetAllKhayals(ctx context.Context, userID, questionID int) ([]entities.Khayal, error) {
	return nil, nil
}

func (f *fakeKhayalService) CreateKhayal(ctx context.Context, userID, questionID int, clientID uuid.UUID, ciphertext []byte, encryptionVersion int) error {
	f.createdUserID = userID
	f.createdQuestionID = questionID
	f.createdClientID = clientID
	f.createdCiphertext = ciphertext
	return nil
}

func authContextRouter(userID int, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/questions", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, userID)
	}, handler)
	return router
}

func TestCreateQuestionRequiresUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	questionService := &fakeQuestionService{}
	ctrl := NewQuestionController(questionService, &fakeKhayalService{})
	router := gin.New()
	router.POST("/questions", ctrl.CreateQuestion)

	req := httptest.NewRequest(http.MethodPost, "/questions", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateQuestionStoresOpaqueCiphertext(t *testing.T) {
	questionService := &fakeQuestionService{}
	ctrl := NewQuestionController(questionService, &fakeKhayalService{})
	router := authContextRouter(42, ctrl.CreateQuestion)

	clientID := uuid.New()
	ciphertext := []byte("0123456789abcdefghijklmnopqrstuv")
	form := url.Values{
		"client_id":           {clientID.String()},
		"question_ciphertext": {base64.RawURLEncoding.EncodeToString(ciphertext)},
		"encryption_version":  {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/questions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if questionService.createdUserID != 42 {
		t.Fatalf("expected user 42, got %d", questionService.createdUserID)
	}
	if questionService.createdClientID != clientID {
		t.Fatal("client id did not reach the service")
	}
	if string(questionService.createdCiphertext) != string(ciphertext) {
		t.Fatal("ciphertext did not reach the service unchanged")
	}
}

func TestCreateQuestionRejectsPlaintextOnlyRequest(t *testing.T) {
	questionService := &fakeQuestionService{}
	ctrl := NewQuestionController(questionService, &fakeKhayalService{})
	router := authContextRouter(42, ctrl.CreateQuestion)

	form := url.Values{"question": {"server-visible plaintext"}}
	req := httptest.NewRequest(http.MethodPost, "/questions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if len(questionService.createdCiphertext) != 0 {
		t.Fatal("server must not store plaintext from the legacy form field")
	}
}

func TestCreateKhayalUsesAuthenticatedUser(t *testing.T) {
	khayalService := &fakeKhayalService{}
	questionService := &fakeQuestionService{}
	ctrl := NewKhayalController(khayalService, questionService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/questions/:questionID/khayals", func(c *gin.Context) {
		c.Set(middleware.UserIDContextKey, 7)
	}, ctrl.CreateKhayal)

	clientID := uuid.New()
	ciphertext := []byte("abcdefghijklmnopqrstuvwxyz012345")
	form := url.Values{
		"client_id":          {clientID.String()},
		"khayal_ciphertext":  {base64.RawURLEncoding.EncodeToString(ciphertext)},
		"encryption_version": {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/questions/9/khayals", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if khayalService.createdUserID != 7 || khayalService.createdQuestionID != 9 {
		t.Fatalf("unexpected ids: user=%d question=%d", khayalService.createdUserID, khayalService.createdQuestionID)
	}
	if string(khayalService.createdCiphertext) != string(ciphertext) {
		t.Fatal("ciphertext did not reach the service unchanged")
	}
}
