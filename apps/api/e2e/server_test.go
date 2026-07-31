package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/bcrypt"
	adinmemory "ddd-second-hand-marketplace/internal/classified-ad/adapter/driven/inmemory"
	adconsumer "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/consumer"
	adhttp "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/http"
	adpublisher "ddd-second-hand-marketplace/internal/classified-ad/adapter/driving/publisher"
	adcommand "ddd-second-hand-marketplace/internal/classified-ad/application/command"
	adquery "ddd-second-hand-marketplace/internal/classified-ad/application/query"
	modinmemory "ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	modconsumer "ddd-second-hand-marketplace/internal/moderation/adapter/driving/consumer"
	modhttp "ddd-second-hand-marketplace/internal/moderation/adapter/driving/http"
	modpublisher "ddd-second-hand-marketplace/internal/moderation/adapter/driving/publisher"
	modcommand "ddd-second-hand-marketplace/internal/moderation/application/command"
	modquery "ddd-second-hand-marketplace/internal/moderation/application/query"
	moddomain "ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
	mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"
)

// mutableClock is a Clock (structurally satisfying both bounded contexts'
// domain.Clock ports) whose current time can be advanced during a test, so
// expiration (a fixed 90-day lifetime) can be exercised without waiting for
// the real ticker.
type mutableClock struct {
	mu  sync.Mutex
	now time.Time
}

func newMutableClock(now time.Time) *mutableClock {
	return &mutableClock{now: now}
}

func (c *mutableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *mutableClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// testServer bundles a running HTTP server (both bounded contexts on one mux)
// with the pieces of the app needed to drive test-only actions that have no
// HTTP route: advancing the clock, running the expiration sweep, asserting on
// the moderation history and on sent emails.
type testServer struct {
	*httptest.Server
	clock             *mutableClock
	expireOutdatedAds adcommand.ExpireOutdatedAdsCommand
	historyRepo       *modinmemory.InMemoryClassifiedAdHistoryRepository
	mailerSpy         *mailertesting.MailerSpy
	moderatorAlice    string // seeded moderator "Alice Martin"
	moderatorBob      string // seeded moderator "Bob Dupont"
}

// newTestServer wires both bounded contexts exactly like cmd/api/main.go, but
// with SYNC in-memory buses everywhere so every event chain (internal bus →
// publisher → public bus → consumers → internal bus → …) completes before the
// triggering HTTP call returns: no sleeps, deterministic assertions.
func newTestServer(t *testing.T) *testServer {
	t.Helper()

	testClock := newMutableClock(time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC))
	mailerSpy := mailertesting.NewMailerSpy()

	// Each bounded context keeps its own internal bus; the public bus only
	// carries the integration DTOs from internal/shared.
	classifiedAdBus := eventbus.NewSyncInMemoryEventBus()
	moderationBus := eventbus.NewSyncInMemoryEventBus()
	publicBus := eventbus.NewSyncInMemoryEventBus()

	mux := http.NewServeMux()

	// --- ClassifiedAd bounded context ---
	adRepo := adinmemory.NewInMemoryClassifiedAdRepository()
	hasher := bcrypt.NewBcryptPasswordHasher()

	submit := adcommand.BuildSubmitClassifiedAdCommand(adRepo, hasher, testClock, classifiedAdBus)
	makeOffer := adcommand.BuildMakeOfferCommand(adRepo, testClock, classifiedAdBus)
	deleteAd := adcommand.BuildDeleteClassifiedAdCommand(adRepo, hasher, testClock, classifiedAdBus)
	edit := adcommand.BuildEditClassifiedAdCommand(adRepo, hasher, testClock, classifiedAdBus)
	approve := adcommand.BuildApproveClassifiedAdCommand(adRepo, testClock, classifiedAdBus)
	publish := adcommand.BuildPublishClassifiedAdCommand(adRepo, testClock, classifiedAdBus)
	reject := adcommand.BuildRejectClassifiedAdCommand(adRepo, testClock, classifiedAdBus)
	challenge := adcommand.BuildChallengeClassifiedAdCommand(adRepo, testClock, classifiedAdBus)
	expireOutdatedAds := adcommand.BuildExpireOutdatedAdsCommand(adRepo, testClock, classifiedAdBus)
	search := adquery.BuildSearchClassifiedAdsQuery(adRepo)
	get := adquery.BuildGetClassifiedAdQuery(adRepo)

	adHandler := adhttp.NewHandler(submit, makeOffer, deleteAd, edit, search, get)
	adHandler.RegisterRoutes(mux)

	// Publisher: bridges internal ClassifiedAd events to the public bus.
	require.NoError(t, adpublisher.RegisterPublishers(classifiedAdBus, publicBus))

	// --- Moderation bounded context ---
	taskRepo := modinmemory.NewInMemoryModerationTaskRepository()
	moderatorRepo := modinmemory.NewInMemoryModeratorRepository()
	historyRepo := modinmemory.NewInMemoryClassifiedAdHistoryRepository()

	claim := modcommand.BuildClaimModerationTaskCommand(taskRepo, moderatorRepo, testClock, moderationBus)
	accept := modcommand.BuildAcceptClassifiedAdCommand(taskRepo, testClock, moderationBus)
	modReject := modcommand.BuildRejectClassifiedAdCommand(taskRepo, testClock, moderationBus)
	modChallenge := modcommand.BuildChallengeClassifiedAdCommand(taskRepo, testClock, moderationBus)
	createTask := modcommand.BuildCreateModerationTaskCommand(taskRepo, testClock)
	appendHistory := modcommand.BuildAppendHistoryEntryCommand(historyRepo)

	listTasks := modquery.BuildListModerationTasksQuery(taskRepo, moderatorRepo, historyRepo)
	getDetail := modquery.BuildGetModerationTaskDetailQuery(taskRepo, moderatorRepo, historyRepo)

	modHandler := modhttp.NewHandler(claim, accept, modReject, modChallenge, listTasks, getDetail)
	modHandler.RegisterRoutes(mux)

	// Publisher: bridges internal Moderation events to the public bus.
	require.NoError(t, modpublisher.RegisterPublishers(moderationBus, publicBus))

	// Moderation history consumers are subscribed FIRST: on a sync bus,
	// subscription order is invocation order, so the moderation decision entry
	// (approved/rejected/challenged) lands in the history before the
	// ClassifiedAd side reacts and appends the chained follow-up entries
	// (published/deleted).
	require.NoError(t, modconsumer.NewClassifiedAdSubmittedConsumer(publicBus, createTask, appendHistory))
	require.NoError(t, modconsumer.NewClassifiedAdEditedConsumer(publicBus, createTask, appendHistory))
	require.NoError(t, modconsumer.NewClassifiedAdPublishedConsumer(publicBus, appendHistory))
	require.NoError(t, modconsumer.NewClassifiedAdDeletedConsumer(publicBus, appendHistory))
	require.NoError(t, modconsumer.NewClassifiedAdExpiredConsumer(publicBus, appendHistory))
	require.NoError(t, modconsumer.NewClassifiedAdApprovedConsumer(publicBus, appendHistory))
	require.NoError(t, modconsumer.NewClassifiedAdRejectedConsumer(publicBus, appendHistory))
	require.NoError(t, modconsumer.NewClassifiedAdChallengedConsumer(publicBus, appendHistory))

	// ClassifiedAd consumers reacting to the public Moderation events.
	require.NoError(t, adconsumer.NewClassifiedAdApprovedConsumer(publicBus, approve))
	// Chained internally: as soon as an ad is approved, it is published.
	require.NoError(t, adconsumer.NewClassifiedAdApprovedInternalConsumer(classifiedAdBus, publish))
	require.NoError(t, adconsumer.NewClassifiedAdRejectedConsumer(publicBus, reject))
	require.NoError(t, adconsumer.NewClassifiedAdChallengedConsumer(publicBus, challenge, adRepo, mailerSpy))

	// Legacy email consumers on the internal bus.
	require.NoError(t, classifiedAdBus.Subscribe("ClassifiedAdPublished", adconsumer.NewAdPublishedEmailConsumer(mailerSpy)))
	require.NoError(t, classifiedAdBus.Subscribe("BuyerOfferMade", adconsumer.NewOfferEmailConsumer(mailerSpy)))

	// Seed two moderators with fixed IDs, like main.go does.
	alice := seedModerator(t, moderatorRepo, "11111111-1111-1111-1111-111111111111", "Alice Martin")
	bob := seedModerator(t, moderatorRepo, "22222222-2222-2222-2222-222222222222", "Bob Dupont")

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return &testServer{
		Server:            srv,
		clock:             testClock,
		expireOutdatedAds: expireOutdatedAds,
		historyRepo:       historyRepo,
		mailerSpy:         mailerSpy,
		moderatorAlice:    alice,
		moderatorBob:      bob,
	}
}

func seedModerator(t *testing.T, repo moddomain.ModeratorRepository, id, fullName string) string {
	t.Helper()

	moderator, err := moddomain.RehydrateModerator(uuid.MustParse(id), fullName)
	require.NoError(t, err)
	require.NoError(t, repo.Save(moderator))
	return moderator.ID().String()
}

// --- Generic HTTP helpers ---

func (s *testServer) get(t *testing.T, path string) *http.Response {
	t.Helper()

	resp, err := http.Get(s.URL + path)
	require.NoError(t, err)
	return resp
}

func (s *testServer) doJSON(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(method, s.URL+path, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()

	defer resp.Body.Close()
	var body T
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

// --- ClassifiedAd helpers ---

func (s *testServer) submitAd(t *testing.T, email, password string) string {
	t.Helper()

	body := adhttp.SubmitClassifiedAdRequest{
		Title:          "Vélo hollandais",
		Description:    "Très bon état",
		PriceInCents:   15000,
		SellerEmail:    email,
		SellerPseudo:   "seller1",
		SellerPassword: password,
		ImageURLs:      []string{"http://example.com/img1.jpg"},
		Category:       "consumer_goods",
		ZipCode:        "75001",
		CityName:       "Paris",
	}

	resp := s.doJSON(t, http.MethodPost, "/classified-ads", body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	respBody := decodeJSON[adhttp.SubmitClassifiedAdResponse](t, resp)
	require.NotEmpty(t, respBody.ID)
	return respBody.ID
}

// editAd corrects a challenged ad (all fields are editable; the test varies
// the title and price to track snapshots across re-submissions).
func (s *testServer) editAd(t *testing.T, id, email, password, title string, priceInCents int64) {
	t.Helper()

	resp := s.doJSON(t, http.MethodPut, "/classified-ads/"+id, adhttp.EditClassifiedAdRequest{
		Email:        email,
		Password:     password,
		Title:        title,
		Description:  "Très bon état (corrigé)",
		PriceInCents: priceInCents,
		ImageURLs:    []string{"http://example.com/img1.jpg"},
		Category:     "consumer_goods",
		ZipCode:      "75001",
		CityName:     "Paris",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func (s *testServer) searchAds(t *testing.T) adhttp.SearchClassifiedAdsResponse {
	t.Helper()

	resp := s.get(t, "/classified-ads?category=consumer_goods")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return decodeJSON[adhttp.SearchClassifiedAdsResponse](t, resp)
}

// assertPubliclyVisible checks the ad appears both in the public search
// listing and on the public detail endpoint.
func (s *testServer) assertPubliclyVisible(t *testing.T, adID string) {
	t.Helper()

	searchBody := s.searchAds(t)
	found := false
	for _, item := range searchBody.Items {
		if item.ID == adID {
			found = true
		}
	}
	assert.True(t, found, "ad %s should appear in the public search listing", adID)

	getResp := s.get(t, "/classified-ads/"+adID)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode, "ad %s should be fetchable by id", adID)
}

// assertPubliclyInvisible checks a non-published ad (submitted, approved,
// challenged, rejected, deleted or expired) leaks neither through the public
// search listing nor through the public detail endpoint.
func (s *testServer) assertPubliclyInvisible(t *testing.T, adID string) {
	t.Helper()

	searchBody := s.searchAds(t)
	for _, item := range searchBody.Items {
		assert.NotEqual(t, adID, item.ID, "ad %s should not appear in the public search listing", adID)
	}

	getResp := s.get(t, "/classified-ads/"+adID)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode, "ad %s should not be fetchable by id", adID)
}

// --- Moderation helpers ---

func (s *testServer) listTasks(t *testing.T) modhttp.ListModerationTasksResponse {
	t.Helper()

	resp := s.get(t, "/moderation/tasks")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return decodeJSON[modhttp.ListModerationTasksResponse](t, resp)
}

// soleTask returns the single active moderation task, failing if the queue
// does not contain exactly one.
func (s *testServer) soleTask(t *testing.T) modhttp.ModerationTaskListItemResponse {
	t.Helper()

	tasks := s.listTasks(t).Tasks
	require.Len(t, tasks, 1, "expected exactly one active moderation task")
	return tasks[0]
}

func (s *testServer) taskDetail(t *testing.T, taskID string) modhttp.ModerationTaskDetailResponse {
	t.Helper()

	resp := s.get(t, "/moderation/tasks/"+taskID)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return decodeJSON[modhttp.ModerationTaskDetailResponse](t, resp)
}

func (s *testServer) claimTask(t *testing.T, taskID, moderatorID string) {
	t.Helper()

	resp := s.doJSON(t, http.MethodPost, "/moderation/tasks/"+taskID+"/claim", modhttp.ClaimModerationTaskRequest{ModeratorID: moderatorID})
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func (s *testServer) acceptTask(t *testing.T, taskID, moderatorID string) {
	t.Helper()

	resp := s.doJSON(t, http.MethodPost, "/moderation/tasks/"+taskID+"/accept", modhttp.AcceptClassifiedAdRequest{ModeratorID: moderatorID})
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func (s *testServer) rejectTask(t *testing.T, taskID, moderatorID, reason string) {
	t.Helper()

	resp := s.doJSON(t, http.MethodPost, "/moderation/tasks/"+taskID+"/reject", modhttp.RejectClassifiedAdRequest{ModeratorID: moderatorID, Reason: reason})
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func (s *testServer) challengeTask(t *testing.T, taskID, moderatorID, reason string) {
	t.Helper()

	resp := s.doJSON(t, http.MethodPost, "/moderation/tasks/"+taskID+"/challenge", modhttp.ChallengeClassifiedAdRequest{ModeratorID: moderatorID, Reason: reason})
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// approveAd runs the whole moderation happy path for the current sole task:
// claim by Alice, then accept. Thanks to the sync buses the ad is approved AND
// published by the time this returns.
func (s *testServer) approveAd(t *testing.T, adID string) {
	t.Helper()

	task := s.soleTask(t)
	detail := s.taskDetail(t, task.ID)
	require.Equal(t, adID, detail.ClassifiedAdID, "the sole active task should target the ad under test")

	s.claimTask(t, task.ID, s.moderatorAlice)
	s.acceptTask(t, task.ID, s.moderatorAlice)
}

// --- History helpers (direct repository access: no HTTP route exposes the
// history once the task is completed and deleted) ---

func (s *testServer) historyEntries(t *testing.T, adID string) []moddomain.HistoryEntry {
	t.Helper()

	history, err := s.historyRepo.FindByClassifiedAdID(adID)
	require.NoError(t, err)
	return history.Entries()
}

// historyActions returns the chronological list of recorded actions for an ad.
func (s *testServer) historyActions(t *testing.T, adID string) []string {
	t.Helper()

	entries := s.historyEntries(t, adID)
	actions := make([]string, len(entries))
	for i, entry := range entries {
		actions[i] = string(entry.Action())
	}
	return actions
}
