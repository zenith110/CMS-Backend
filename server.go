package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/zenith110/portfilo/graph"
	"github.com/zenith110/portfilo/graph/generated"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("GRAPHQLPORT")
	domain := os.Getenv("DOMAIN")
	environment := os.Getenv("ENV")
	if port == "" {
		port = defaultPort
	}
	router := chi.NewRouter()
	if environment == "PROD" {
		router.Use(cors.New(cors.Options{
			AllowedOrigins:   []string{"https://*abrahannevarez.dev"},
			AllowedMethods:   []string{http.MethodGet, http.MethodPost},
			AllowCredentials: true,
			Debug:            true,
		}).Handler)
	} else if environment == "LOCAL" {
		router.Use(cors.New(cors.Options{
			AllowedOrigins:   []string{"http://*"},
			AllowedMethods:   []string{http.MethodGet, http.MethodPost},
			AllowCredentials: true,
			Debug:            true,
		}).Handler)
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Check against your desired domains here
				return r.Host == domain
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, router))
}
