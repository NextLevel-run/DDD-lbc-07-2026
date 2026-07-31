package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/bcrypt"
	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/clock"
	adinmemory "ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	adconsumer "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/consumer"
	adhttp "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/http"
	adpublisher "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/publisher"
	adcommand "ddd-second-hand-marketplace/internal/classified-ad/application/command"
	adquery "ddd-second-hand-marketplace/internal/classified-ad/application/query"
	moddomain "ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	"ddd-second-hand-marketplace/pkg/mailer"

	modinmemory "ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	modconsumer "ddd-second-hand-marketplace/internal/moderation/adapter/driving/consumer"
	modhttp "ddd-second-hand-marketplace/internal/moderation/adapter/driving/http"
	modpublisher "ddd-second-hand-marketplace/internal/moderation/adapter/driving/publisher"
	modcommand "ddd-second-hand-marketplace/internal/moderation/application/command"
	modquery "ddd-second-hand-marketplace/internal/moderation/application/query"
)

// Seeded moderators: no authentication is in scope, so two moderators with
// well-known fixed IDs are created at startup and logged, ready to be used
// as moderatorId in the moderation HTTP API.
var seededModerators = []struct {
	id       string
	fullName string
}{
	{"11111111-1111-1111-1111-111111111111", "Alice Martin"},
	{"22222222-2222-2222-2222-222222222222", "Bob Dupont"},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Shared infrastructure. Each bounded context keeps its own internal bus;
	// the public bus only carries the integration DTOs from internal/shared.
	classifiedAdBus := eventbus.NewAsyncInMemoryEventBus()
	moderationBus := eventbus.NewAsyncInMemoryEventBus()
	publicBus := eventbus.NewAsyncInMemoryEventBus()
	systemClock := clock.NewSystemClock()
	fakeMailer := mailer.NewFakeMailer()

	// --- ClassifiedAd bounded context ---
	adRepo := adinmemory.NewInMemoryClassifiedAdRepository()
	hasher := bcrypt.NewBcryptPasswordHasher()

	// Commands
	submitClassifiedAd := adcommand.BuildSubmitClassifiedAdCommand(adRepo, hasher, systemClock, classifiedAdBus)
	makeOffer := adcommand.BuildMakeOfferCommand(adRepo, systemClock, classifiedAdBus)
	deleteClassifiedAd := adcommand.BuildDeleteClassifiedAdCommand(adRepo, hasher, systemClock, classifiedAdBus)
	editClassifiedAd := adcommand.BuildEditClassifiedAdCommand(adRepo, hasher, systemClock, classifiedAdBus)
	approveClassifiedAd := adcommand.BuildApproveClassifiedAdCommand(adRepo, systemClock, classifiedAdBus)
	publishClassifiedAd := adcommand.BuildPublishClassifiedAdCommand(adRepo, systemClock, classifiedAdBus)
	rejectClassifiedAd := adcommand.BuildRejectClassifiedAdCommand(adRepo, systemClock, classifiedAdBus)
	challengeClassifiedAd := adcommand.BuildChallengeClassifiedAdCommand(adRepo, systemClock, classifiedAdBus)
	expireOutdatedAds := adcommand.BuildExpireOutdatedAdsCommand(adRepo, systemClock, classifiedAdBus)

	// Queries
	searchClassifiedAds := adquery.BuildSearchClassifiedAdsQuery(adRepo)
	getClassifiedAd := adquery.BuildGetClassifiedAdQuery(adRepo)

	// HTTP adapter
	adHandler := adhttp.NewHandler(submitClassifiedAd, makeOffer, deleteClassifiedAd, editClassifiedAd, searchClassifiedAds, getClassifiedAd)
	adHandler.RegisterRoutes(mux)

	// Publisher: bridges internal ClassifiedAd events to the public bus.
	if err := adpublisher.RegisterPublishers(classifiedAdBus, publicBus); err != nil {
		log.Fatalf("failed to register classified-ad publishers: %v", err)
	}

	// Consumers reacting to the public Moderation events.
	if err := adconsumer.NewClassifiedAdApprovedConsumer(publicBus, approveClassifiedAd); err != nil {
		log.Fatalf("failed to subscribe classified-ad approved consumer: %v", err)
	}
	// Chained internally: as soon as an ad is approved, it is published.
	if err := adconsumer.NewClassifiedAdApprovedInternalConsumer(classifiedAdBus, publishClassifiedAd); err != nil {
		log.Fatalf("failed to subscribe classified-ad approved internal consumer: %v", err)
	}
	if err := adconsumer.NewClassifiedAdRejectedConsumer(publicBus, rejectClassifiedAd); err != nil {
		log.Fatalf("failed to subscribe classified-ad rejected consumer: %v", err)
	}
	if err := adconsumer.NewClassifiedAdChallengedConsumer(publicBus, challengeClassifiedAd, adRepo, fakeMailer); err != nil {
		log.Fatalf("failed to subscribe classified-ad challenged consumer: %v", err)
	}

	// Legacy email consumers on the internal bus.
	if err := classifiedAdBus.Subscribe("ClassifiedAdPublished", adconsumer.NewAdPublishedEmailConsumer(fakeMailer)); err != nil {
		log.Fatalf("failed to subscribe ad published email consumer: %v", err)
	}
	if err := classifiedAdBus.Subscribe("BuyerOfferMade", adconsumer.NewOfferEmailConsumer(fakeMailer)); err != nil {
		log.Fatalf("failed to subscribe offer email consumer: %v", err)
	}

	// --- Moderation bounded context ---
	taskRepo := modinmemory.NewInMemoryModerationTaskRepository()
	moderatorRepo := modinmemory.NewInMemoryModeratorRepository()
	historyRepo := modinmemory.NewInMemoryClassifiedAdHistoryRepository()

	// Commands
	claimModerationTask := modcommand.BuildClaimModerationTaskCommand(taskRepo, moderatorRepo, systemClock, moderationBus)
	acceptClassifiedAd := modcommand.BuildAcceptClassifiedAdCommand(taskRepo, systemClock, moderationBus)
	moderationRejectClassifiedAd := modcommand.BuildRejectClassifiedAdCommand(taskRepo, systemClock, moderationBus)
	moderationChallengeClassifiedAd := modcommand.BuildChallengeClassifiedAdCommand(taskRepo, systemClock, moderationBus)
	createModerationTask := modcommand.BuildCreateModerationTaskCommand(taskRepo, systemClock)
	appendHistoryEntry := modcommand.BuildAppendHistoryEntryCommand(historyRepo)

	// Queries
	listModerationTasks := modquery.BuildListModerationTasksQuery(taskRepo, moderatorRepo, historyRepo)
	getModerationTaskDetail := modquery.BuildGetModerationTaskDetailQuery(taskRepo, moderatorRepo, historyRepo)

	// HTTP adapter
	moderationHandler := modhttp.NewHandler(claimModerationTask, acceptClassifiedAd, moderationRejectClassifiedAd, moderationChallengeClassifiedAd, listModerationTasks, getModerationTaskDetail)
	moderationHandler.RegisterRoutes(mux)

	// Publisher: bridges internal Moderation events to the public bus.
	if err := modpublisher.RegisterPublishers(moderationBus, publicBus); err != nil {
		log.Fatalf("failed to register moderation publishers: %v", err)
	}

	// Consumers reacting to the public ClassifiedAd events.
	if err := modconsumer.NewClassifiedAdSubmittedConsumer(publicBus, createModerationTask, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation submitted consumer: %v", err)
	}
	if err := modconsumer.NewClassifiedAdEditedConsumer(publicBus, createModerationTask, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation edited consumer: %v", err)
	}
	if err := modconsumer.NewClassifiedAdPublishedConsumer(publicBus, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation published consumer: %v", err)
	}
	if err := modconsumer.NewClassifiedAdDeletedConsumer(publicBus, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation deleted consumer: %v", err)
	}
	if err := modconsumer.NewClassifiedAdExpiredConsumer(publicBus, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation expired consumer: %v", err)
	}
	// Consumers reacting to Moderation's own public events (history trail).
	if err := modconsumer.NewClassifiedAdApprovedConsumer(publicBus, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation approved consumer: %v", err)
	}
	if err := modconsumer.NewClassifiedAdRejectedConsumer(publicBus, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation rejected consumer: %v", err)
	}
	if err := modconsumer.NewClassifiedAdChallengedConsumer(publicBus, appendHistoryEntry); err != nil {
		log.Fatalf("failed to subscribe moderation challenged consumer: %v", err)
	}

	// Seed the moderators.
	for _, seed := range seededModerators {
		moderator, err := moddomain.RehydrateModerator(uuid.MustParse(seed.id), seed.fullName)
		if err != nil {
			log.Fatalf("failed to build seeded moderator %q: %v", seed.fullName, err)
		}
		if err := moderatorRepo.Save(moderator); err != nil {
			log.Fatalf("failed to save seeded moderator %q: %v", seed.fullName, err)
		}
		log.Printf("seeded moderator %q with id %s", moderator.FullName(), moderator.ID())
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
