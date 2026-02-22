package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/shared/logger"
	"github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/shared/middleware"
	"github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/tasks/internal/client/authclient"
	handlers "github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/tasks/internal/http"
	"github.com/sun1tar/MIREA-TIP-Practice-18/tech-ip-sem2/tasks/internal/service"
)

func main() {
	// Инициализация структурированного логгера
	logrusLogger := logger.Init("tasks")

	tasksPort := os.Getenv("TASKS_PORT")
	if tasksPort == "" {
		tasksPort = "8082"
	}
	authGrpcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGrpcAddr == "" {
		authGrpcAddr = "localhost:50051"
	}

	authClient, err := authclient.NewClient(authGrpcAddr, 2*time.Second, logrusLogger)
	if err != nil {
		logrusLogger.WithError(err).Fatal("Failed to create auth client")
	}
	defer authClient.Close()

	taskService := service.NewTaskService()
	taskHandler := handlers.NewTaskHandler(taskService, authClient, logrusLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /v1/tasks", taskHandler.ListTasks)
	mux.HandleFunc("GET /v1/tasks/{id}", taskHandler.GetTask)
	mux.HandleFunc("PATCH /v1/tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /v1/tasks/{id}", taskHandler.DeleteTask)

	// RequestIDMiddleware должен идти первым
	handler := middleware.RequestIDMiddleware(middleware.LoggingMiddleware(mux))

	addr := fmt.Sprintf(":%s", tasksPort)
	logrusLogger.WithField("port", tasksPort).Info("Tasks service starting")
	if err := http.ListenAndServe(addr, handler); err != nil {
		logrusLogger.WithError(err).Fatal("server failed")
	}
}
