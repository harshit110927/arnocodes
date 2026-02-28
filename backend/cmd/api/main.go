package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/harshit110927/arnocodes/backend/config"
	"github.com/harshit110927/arnocodes/backend/internal/assessment"
	"github.com/harshit110927/arnocodes/backend/internal/course"
	"github.com/harshit110927/arnocodes/backend/internal/dashboard"
	"github.com/harshit110927/arnocodes/backend/internal/database"
	"github.com/harshit110927/arnocodes/backend/internal/handlers"
	"github.com/harshit110927/arnocodes/backend/internal/ide"
	"github.com/harshit110927/arnocodes/backend/internal/learning/activity"
	"github.com/harshit110927/arnocodes/backend/internal/middleware"
)

type assessmentCourseStatusAdapter struct {
	repo *assessment.Repository
}

func (a assessmentCourseStatusAdapter) GetUserStatus(ctx context.Context, userID string) (course.DiagnosticUserStatus, error) {
	status, err := a.repo.GetUserStatus(ctx, userID)
	if err != nil {
		return course.DiagnosticUserStatus{}, err
	}
	return course.DiagnosticUserStatus{DiagnosticCompleted: status.DiagnosticCompleted}, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()
	ctx := context.Background()

	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(ctx, db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	if err := database.RunSeed(ctx, db); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	assessmentRepo := assessment.NewRepository(db.Pool())
	courseRepo := course.NewCourseRepository(db.Pool())
	learningActivityRepo := activity.NewActivityRepository(db.Pool())
	dashboardRepo := dashboard.NewRepository(db.Pool())
	ideRepo := ide.NewRepository(db.Pool())
	ideService := ide.NewService(ideRepo, ide.NewDockerEvaluator(), learningActivityRepo)

	authMiddleware, err := middleware.NewJWKSAuthMiddleware(cfg.SupabaseURL, cfg.SupabaseAudience)
	if err != nil {
		log.Fatalf("auth middleware init failed: %v", err)
	}
	h := handlers.NewHandler(cfg, assessmentRepo, courseRepo, assessmentCourseStatusAdapter{repo: assessmentRepo}, dashboardRepo, ideService, authMiddleware)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go assessment.StartCodingEvaluationWorker(workerCtx, assessment.NewService(assessmentRepo), 10*time.Second, 20)
	go ide.StartIDEWorker(workerCtx, db.Pool(), ideRepo, ideRepo, ide.NewDockerEvaluator(), learningActivityRepo)

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	workerCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
