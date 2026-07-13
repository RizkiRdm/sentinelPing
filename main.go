package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/RizkiRdm/sentinelPing/internal/auth"
	"github.com/RizkiRdm/sentinelPing/internal/db"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/sentinelping.db"
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	if err := run(dbPath, listenAddr); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(dbPath, listenAddr string) error {
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		return err
	}

	slog.Info("migrations complete")

	authRepo := auth.NewRepository(database)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.StripSlashes)
	r.Use(authSvc.Middleware)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Get("/me", authHandler.Me)
	})

	staticFS, err := fs.Sub(WebFS, "web/dist")
	if err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(staticFS))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			_, err := staticFS.Open(r.URL.Path[1:])
			if err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	slog.Info("server starting", "addr", listenAddr)
	return http.ListenAndServe(listenAddr, r)
}
