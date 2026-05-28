package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/chrisapos3/mmo-rpg/internal/api/handler"
	"github.com/chrisapos3/mmo-rpg/internal/api/middleware"
	"github.com/chrisapos3/mmo-rpg/internal/service"
)

func NewRouter(authSvc *service.AuthService, onboardingSvc *service.OnboardingService, githubSvc *service.GitHubService, signalSvc *service.SignalService, profileH *handler.ProfileHandler, frontendURL string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors)

	auth := handler.NewAuthHandler(authSvc)
	onboarding := handler.NewOnboardingHandler(onboardingSvc)
	github := handler.NewGitHubHandler(githubSvc, frontendURL)
	signal := handler.NewSignalHandler(signalSvc)
	authMW := middleware.Auth(authSvc)


	r.Get("/health", handler.Health)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register)
			r.Post("/login", auth.Login)

			r.Group(func(r chi.Router) {
				r.Use(authMW)
				r.Get("/me", auth.Me)
			})
		})

		r.Route("/onboarding", func(r chi.Router) {
			r.Use(authMW)
			r.Post("/cv", onboarding.UploadCV)
			r.Get("/cv/status", onboarding.CVStatus)
			r.Post("/build", onboarding.GenerateBuild)
			r.Get("/build", onboarding.GetBuild)
		})

		r.Route("/signal", func(r chi.Router) {
			r.Use(authMW)
			r.Get("/scores", signal.GetScores)
			r.Get("/evidence", signal.GetEvidence)
		})

		// Public profile — no auth
		r.Get("/p/{userID}", profileH.GetPublic)

		r.Route("/profile", func(r chi.Router) {
			r.Use(authMW)
			r.Post("/publish", profileH.Publish)
		})

		r.Route("/github", func(r chi.Router) {
			// Callback is unauthenticated — browser redirect from GitHub
			r.Get("/callback", github.Callback)

			r.Group(func(r chi.Router) {
				r.Use(authMW)
				r.Get("/authorize", github.Authorize)
				r.Get("/status", github.Status)
				r.Post("/sync", github.Sync)
			})
		})
	})

	return r
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
