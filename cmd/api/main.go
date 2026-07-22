package main

import (
	"log"
	"net/http"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	classifiedadhttp "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/http"
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

func main() {
	eventBus := eventbus.NewAsyncInMemoryEventBus()

	classifiedAdRepo := inmemory.NewInMemoryClassifiedAdRepository()
	postClassifiedAdCommand := command.BuildPostClassifiedAdCommand(classifiedAdRepo, eventBus)

	classifiedAdHandler := classifiedadhttp.NewHandler(postClassifiedAdCommand)

	mux := http.NewServeMux()
	mux.HandleFunc("/classified-ads", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			classifiedAdHandler.PostClassifiedAd(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
