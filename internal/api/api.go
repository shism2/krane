package api

import (
	"net/http"
	"os"
	"time"

	"github.com/biensupernice/krane/internal/api/controllers"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/biensupernice/krane/internal/api/middlewares"
)

func Run() {
	logrus.Debugf("Starting Krane API on pid: %d", os.Getpid())
	router := mux.NewRouter()

	withBaseMiddlewares(router)
	withRoutes(router)

	srv := &http.Server{
		Handler:      router,
		Addr:         os.Getenv("LISTEN_ADDRESS"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Infof("API on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err.Error())
	}
}

func withBaseMiddlewares(router *mux.Router) {
	router.Use(middlewares.Logging)
	router.Use(handlers.RecoveryHandler())
	router.Use(handlers.CORS(
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost}),
		handlers.AllowedOrigins([]string{"*"}))) // TODO: Not allowing wild card origins (*) use envar LISTEN_ADDRESS
}

func withRoutes(router *mux.Router) {
	noAuthRouter := router.PathPrefix("/").Subrouter()
	withRoute(noAuthRouter, "/", controllers.GetServerStatus).Methods(http.MethodGet)
	withRoute(noAuthRouter, "/login", controllers.RequestLoginPhrase).Methods(http.MethodGet)
	withRoute(noAuthRouter, "/auth", controllers.AuthenticateClientJWT).Methods(http.MethodPost)
	withRoute(noAuthRouter, "/ping/{namespace}/{message}", controllers.PingController).Methods(http.MethodGet)

	authRouter := router.PathPrefix("/").Subrouter()
	// deployments
	withRoute(authRouter, "/deployments", controllers.GetAllDeployments, middlewares.AuthSessionMiddleware).Methods(http.MethodGet)
	withRoute(authRouter, "/deployments", controllers.SaveDeployment, middlewares.AuthSessionMiddleware).Methods(http.MethodPost)
	withRoute(authRouter, "/deployments/{name}", controllers.GetDeployment, middlewares.AuthSessionMiddleware).Methods(http.MethodGet)
	withRoute(authRouter, "/deployments/{name}", controllers.DeleteDeployment, middlewares.AuthSessionMiddleware).Methods(http.MethodDelete)
	// secrets
	withRoute(authRouter, "/secrets/{name}", controllers.GetSecrets, middlewares.AuthSessionMiddleware).Methods(http.MethodGet)
	withRoute(authRouter, "/secrets/{name}", controllers.CreateSecret, middlewares.AuthSessionMiddleware).Methods(http.MethodPost)
	// jobs
	withRoute(authRouter, "/jobs", controllers.GetRecentJobs, middlewares.AuthSessionMiddleware).Methods(http.MethodGet)
	withRoute(authRouter, "/jobs/{namespace}", controllers.GetJobsByNamespace, middlewares.AuthSessionMiddleware).Methods(http.MethodGet)
	withRoute(authRouter, "/jobs/{namespace}/{id}", controllers.GetJobByID, middlewares.AuthSessionMiddleware).Methods(http.MethodGet)
	// session
	withRoute(authRouter, "/sessions", controllers.GetSessions, middlewares.AuthSessionMiddleware).Methods(http.MethodGet)
}

type routeHandler func(http.ResponseWriter, *http.Request)

func withRoute(r *mux.Router, path string, handler routeHandler, middlewares ...mux.MiddlewareFunc) *mux.Route {
	for _, mw := range middlewares {
		r.Use(mw)
	}
	return r.HandleFunc(path, handler)
}
