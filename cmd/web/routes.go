package main

import (
	"net/http"

	"github.com/Gaussgeek/rentals/internal/config"
	"github.com/Gaussgeek/rentals/internal/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about/", handlers.Repo.About)

	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/user/signup", handlers.Repo.SignUp)
	mux.Post("/user/signup", handlers.Repo.PostSignUp)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		//mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/add-new-property", handlers.Repo.AdminAddNewProperty)
		mux.Post("/add-new-property", handlers.Repo.AdminPostAddNewProperty)

		mux.Get("/all-properties", handlers.Repo.AdminAllPropertiesByID)
		mux.Get("/all-properties/{id}/show", handlers.Repo.AdminShowPropertyByPropertyID)

		mux.Get("/all-properties/{id}/add-unit", handlers.Repo.AdminAddUnitToProperty)
		mux.Post("/all-properties/{id}/add-unit", handlers.Repo.AdminPostAddUnitToProperty)

		mux.Get("/all-properties/{id}/view-units", handlers.Repo.AdminShowUnitsByPropertyID)

		mux.Get("/unit-details/{id}/show", handlers.Repo.AdminShowUnitDetails)
	})

	return mux
}
