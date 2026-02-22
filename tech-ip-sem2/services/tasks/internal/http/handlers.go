package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/shared/middleware"
	"github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/tasks/internal/client/authclient"
	"github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/tasks/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
	authClient  *authclient.Client
	logger      *logrus.Logger
}

func NewTaskHandler(ts *service.TaskService, ac *authclient.Client, logger *logrus.Logger) *TaskHandler {
	return &TaskHandler{
		taskService: ts,
		authClient:  ac,
		logger:      logger,
	}
}

func (h *TaskHandler) verifyToken(w http.ResponseWriter, r *http.Request) bool {
	requestID := middleware.GetRequestID(r.Context())
	logEntry := h.logger.WithFields(logrus.Fields{
		"component":  "http_handler",
		"request_id": requestID,
	})

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logEntry.Warn("missing authorization header")
		http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
		return false
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		logEntry.WithField("auth_header", authHeader).Warn("invalid authorization header format")
		http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
		return false
	}
	token := parts[1]

	valid, _, err := h.authClient.VerifyToken(r.Context(), token)
	if err != nil {
		logEntry.WithError(err).Error("authentication service unavailable")
		http.Error(w, `{"error":"authentication service unavailable"}`, http.StatusServiceUnavailable)
		return false
	}
	if !valid {
		logEntry.WithField("token_present", token != "").Warn("invalid token")
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return false
	}

	logEntry.Debug("token verified successfully")
	return true
}

// ... остальные методы с добавлением логирования ошибок

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logEntry := h.logger.WithFields(logrus.Fields{
		"component":  "http_handler",
		"handler":    "CreateTask",
		"request_id": requestID,
	})

	if !h.verifyToken(w, r) {
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logEntry.WithError(err).Warn("invalid request body")
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		logEntry.Warn("title is required")
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	task := service.Task{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Done:        false,
	}
	created := h.taskService.Create(task)

	logEntry.WithField("task_id", created.ID).Info("task created successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toTaskResponse(created))
}

// Аналогично обновить остальные методы:
// - ListTasks
// - GetTask (логировать 404)
// - UpdateTask (логировать успех/ошибку)
// - DeleteTask (логировать успех/ошибку)
