package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/bcrypt"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/clock"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/consumer"
	httpadapter "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/http"
	"ddd-second-hand-marketplace/internal/classified-ad/application/command"
	"ddd-second-hand-marketplace/internal/classified-ad/application/query"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Infrastructure
	eventBus := eventbus.NewAsyncInMemoryEventBus()
	repo := inmemory.NewInMemoryClassifiedAdRepository()
	hasher := bcrypt.NewBcryptPasswordHasher()
	systemClock := clock.NewSystemClock()
	fakeMailer := mailer.NewFakeMailer()

	// Commands
	submitClassifiedAd := command.BuildSubmitClassifiedAdCommand(repo, hasher, systemClock, eventBus)
	makeOffer := command.BuildMakeOfferCommand(repo, systemClock, eventBus)
	deleteClassifiedAd := command.BuildDeleteClassifiedAdCommand(repo, hasher, systemClock, eventBus)
	expireOutdatedAds := command.BuildExpireOutdatedAdsCommand(repo, systemClock, eventBus)

	// Queries
	searchClassifiedAds := query.BuildSearchClassifiedAdsQuery(repo)
	getClassifiedAd := query.BuildGetClassifiedAdQuery(repo)

	// HTTP adapter
	handler := httpadapter.NewHandler(submitClassifiedAd, makeOffer, deleteClassifiedAd, searchClassifiedAds, getClassifiedAd)
	handler.RegisterRoutes(mux)

	// Event consumers
	if err := eventBus.Subscribe("ClassifiedAdPublished", consumer.NewAdPublishedEmailConsumer(fakeMailer)); err != nil {
		log.Fatalf("failed to subscribe ad published email consumer: %v", err)
	}
	if err := eventBus.Subscribe("BuyerOfferMade", consumer.NewOfferEmailConsumer(fakeMailer)); err != nil {
		log.Fatalf("failed to subscribe offer email consumer: %v", err)
	}

	// Periodic expiration of outdated ads
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			count, err := expireOutdatedAds()
			if err != nil {
				log.Printf("failed to expire outdated ads: %v", err)
				continue
			}
			log.Printf("expired %d outdated classified ad(s)", count)
		}
	}()

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Health check available at http://localhost:8080/health")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
