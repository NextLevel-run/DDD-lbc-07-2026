package httpadapter

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ddd-second-hand-marketplace/internal/moderation/adapter/driven/inmemory"
	"ddd-second-hand-marketplace/internal/moderation/application/command"
	"ddd-second-hand-marketplace/internal/moderation/application/query"
	"ddd-second-hand-marketplace/internal/moderation/domain"
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// fakeClock is a settable implementation of domain.Clock for deterministic tests.
type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

// handlerTestSetup wires the handler on top of the real commands, queries and
// in-memory repositories, exercising the full HTTP → application → domain path.
type handlerTestSetup struct {
	mux           *http.ServeMux
	taskRepo      *inmemory.InMemoryModerationTaskRepository
	moderatorRepo *inmemory.InMemoryModeratorRepository
	historyRepo   *inmemory.InMemoryClassifiedAdHistoryRepository
	clock         *fakeClock
}

func setupHandlerTest(t *testing.T) *handlerTestSetup {
	t.Helper()

	taskRepo := inmemory.NewInMemoryModerationTaskRepository()
	moderatorRepo := inmemory.NewInMemoryModeratorRepository()
	historyRepo := inmemory.NewInMemoryClassifiedAdHistoryRepository()
	clock := &fakeClock{now: time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)}
	internalBus := eventbus.NewSyncInMemoryEventBus()

	handler := NewHandler(
		command.BuildClaimModerationTaskCommand(taskRepo, moderatorRepo, clock, internalBus),
		command.BuildAcceptClassifiedAdCommand(taskRepo, clock, internalBus),
		command.BuildRejectClassifiedAdCommand(taskRepo, clock, internalBus),
		command.BuildChallengeClassifiedAdCommand(taskRepo, clock, internalBus),
		query.BuildListModerationTasksQuery(taskRepo, moderatorRepo, historyRepo),
		query.BuildGetModerationTaskDetailQuery(taskRepo, moderatorRepo, historyRepo),
	)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return &handlerTestSetup{
		mux:           mux,
		taskRepo:      taskRepo,
		moderatorRepo: moderatorRepo,
		historyRepo:   historyRepo,
		clock:         clock,
	}
}

// seedModerator creates and stores a moderator with the given full name.
func (s *handlerTestSetup) seedModerator(t *testing.T, fullName string) *domain.Moderator {
	t.Helper()

	moderator, err := domain.NewModerator(fullName)
	require.NoError(t, err)
	require.NoError(t, s.moderatorRepo.Save(moderator))
	return moderator
}

// seedTask creates and stores an unclaimed moderation task for the given ad.
func (s *handlerTestSetup) seedTask(t *testing.T, classifiedAdID string) *domain.ModerationTask {
	t.Helper()

	task, err := domain.NewModerationTask(classifiedAdID, s.clock.Now())
	require.NoError(t, err)
	require.NoError(t, s.taskRepo.Save(task))
	return task
}

// seedClaimedTask creates and stores a task already claimed by the given moderator.
func (s *handlerTestSetup) seedClaimedTask(t *testing.T, classifiedAdID string, moderator *domain.Moderator) *domain.ModerationTask {
	t.Helper()

	task := s.seedTask(t, classifiedAdID)
	require.NoError(t, task.Claim(moderator.ID(), s.clock.Now()))
	require.NoError(t, s.taskRepo.Save(task))
	return task
}

// seedHistory stores a history for the given ad with a submitted entry carrying
// the given title in its snapshot.
func (s *handlerTestSetup) seedHistory(t *testing.T, classifiedAdID, title string) {
	t.Helper()

	history, err := domain.NewClassifiedAdHistory(classifiedAdID)
	require.NoError(t, err)
	entry, err := domain.NewHistoryEntry(
		s.clock.Now(),
		domain.HistoryActionSubmitted,
		nil,
		nil,
		&domain.ClassifiedAdSnapshot{Title: title, Description: "desc", PriceInCents: 1000},
	)
	require.NoError(t, err)
	history.Append(entry)
	require.NoError(t, s.historyRepo.Save(history))
}

// do performs a request against the handler's mux and returns the recorder.
func (s *handlerTestSetup) do(t *testing.T, method, target string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewBuffer(jsonBody)
	} else {
		reader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec
}

// --- GET /moderation/tasks ---

func TestListModerationTasks_EmptyQueue(t *testing.T) {
	setup := setupHandlerTest(t)

	rec := setup.do(t, http.MethodGet, "/moderation/tasks", nil)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response ListModerationTasksResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	assert.Empty(t, response.Tasks)
}

func TestListModerationTasks_ReturnsPendingAndClaimedTasks(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	pending := setup.seedTask(t, "ad-1")
	setup.seedHistory(t, "ad-1", "Vintage bike")
	setup.clock.now = setup.clock.now.Add(time.Minute)
	claimed := setup.seedClaimedTask(t, "ad-2", moderator)

	rec := setup.do(t, http.MethodGet, "/moderation/tasks", nil)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response ListModerationTasksResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	require.Len(t, response.Tasks, 2)

	assert.Equal(t, pending.ID().String(), response.Tasks[0].ID)
	assert.Equal(t, "Vintage bike", response.Tasks[0].ClassifiedAdTitle)
	assert.Equal(t, "pending", response.Tasks[0].Status)
	assert.Empty(t, response.Tasks[0].ClaimedBy)

	assert.Equal(t, claimed.ID().String(), response.Tasks[1].ID)
	assert.Equal(t, "claimed", response.Tasks[1].Status)
	assert.Equal(t, "Jane Doe", response.Tasks[1].ClaimedBy)
}

// --- GET /moderation/tasks/{id} ---

func TestGetModerationTaskDetail_Success(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	task := setup.seedClaimedTask(t, "ad-1", moderator)
	setup.seedHistory(t, "ad-1", "Vintage bike")

	rec := setup.do(t, http.MethodGet, "/moderation/tasks/"+task.ID().String(), nil)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response ModerationTaskDetailResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))

	assert.Equal(t, task.ID().String(), response.ID)
	assert.Equal(t, "ad-1", response.ClassifiedAdID)
	assert.Equal(t, "claimed", response.Status)
	assert.Equal(t, "Jane Doe", response.ClaimedBy)
	assert.Equal(t, moderator.ID().String(), response.ModeratorID)
	require.NotNil(t, response.ClaimedAt)

	require.Len(t, response.History, 1)
	assert.Equal(t, "submitted", response.History[0].Action)
	require.NotNil(t, response.History[0].Snapshot)
	assert.Equal(t, "Vintage bike", response.History[0].Snapshot.Title)

	require.NotNil(t, response.LastSnapshot)
	assert.Equal(t, "Vintage bike", response.LastSnapshot.Title)
}

func TestGetModerationTaskDetail_PendingTaskWithoutHistory(t *testing.T) {
	setup := setupHandlerTest(t)
	task := setup.seedTask(t, "ad-1")

	rec := setup.do(t, http.MethodGet, "/moderation/tasks/"+task.ID().String(), nil)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response ModerationTaskDetailResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	assert.Equal(t, "pending", response.Status)
	assert.Empty(t, response.ClaimedBy)
	assert.Nil(t, response.ClaimedAt)
	assert.Empty(t, response.History)
	assert.Nil(t, response.LastSnapshot)
}

func TestGetModerationTaskDetail_UnknownTask(t *testing.T) {
	setup := setupHandlerTest(t)

	rec := setup.do(t, http.MethodGet, "/moderation/tasks/"+uuid.NewString(), nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetModerationTaskDetail_MalformedTaskID(t *testing.T) {
	setup := setupHandlerTest(t)

	rec := setup.do(t, http.MethodGet, "/moderation/tasks/not-a-uuid", nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- POST /moderation/tasks/{id}/claim ---

func TestClaimModerationTask_Success(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	task := setup.seedTask(t, "ad-1")

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/claim",
		ClaimModerationTaskRequest{ModeratorID: moderator.ID().String()})

	assert.Equal(t, http.StatusNoContent, rec.Code)

	stored, err := setup.taskRepo.FindByID(task.ID())
	require.NoError(t, err)
	assert.True(t, stored.IsClaimed())
	assert.Equal(t, moderator.ID(), *stored.ModeratorID())
}

func TestClaimModerationTask_AlreadyClaimedReturnsConflict(t *testing.T) {
	setup := setupHandlerTest(t)
	owner := setup.seedModerator(t, "Jane Doe")
	other := setup.seedModerator(t, "John Smith")
	task := setup.seedClaimedTask(t, "ad-1", owner)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/claim",
		ClaimModerationTaskRequest{ModeratorID: other.ID().String()})

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestClaimModerationTask_UnknownTaskReturnsNotFound(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+uuid.NewString()+"/claim",
		ClaimModerationTaskRequest{ModeratorID: moderator.ID().String()})

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestClaimModerationTask_UnknownModeratorReturnsNotFound(t *testing.T) {
	setup := setupHandlerTest(t)
	task := setup.seedTask(t, "ad-1")

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/claim",
		ClaimModerationTaskRequest{ModeratorID: uuid.NewString()})

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestClaimModerationTask_InvalidBodyReturnsBadRequest(t *testing.T) {
	setup := setupHandlerTest(t)
	task := setup.seedTask(t, "ad-1")

	req := httptest.NewRequest(http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/claim",
		bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()
	setup.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	assert.Contains(t, response.Error, "invalid request body")
}

// --- POST /moderation/tasks/{id}/accept ---

func TestAcceptClassifiedAd_SuccessDeletesTask(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	task := setup.seedClaimedTask(t, "ad-1", moderator)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/accept",
		AcceptClassifiedAdRequest{ModeratorID: moderator.ID().String()})

	assert.Equal(t, http.StatusNoContent, rec.Code)

	_, err := setup.taskRepo.FindByID(task.ID())
	assert.ErrorIs(t, err, domain.ErrModerationTaskNotFound)
}

func TestAcceptClassifiedAd_NotOwnerReturnsForbidden(t *testing.T) {
	setup := setupHandlerTest(t)
	owner := setup.seedModerator(t, "Jane Doe")
	other := setup.seedModerator(t, "John Smith")
	task := setup.seedClaimedTask(t, "ad-1", owner)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/accept",
		AcceptClassifiedAdRequest{ModeratorID: other.ID().String()})

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAcceptClassifiedAd_UnknownTaskReturnsNotFound(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+uuid.NewString()+"/accept",
		AcceptClassifiedAdRequest{ModeratorID: moderator.ID().String()})

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- POST /moderation/tasks/{id}/reject ---

func TestRejectClassifiedAd_SuccessDeletesTask(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	task := setup.seedClaimedTask(t, "ad-1", moderator)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/reject",
		RejectClassifiedAdRequest{ModeratorID: moderator.ID().String(), Reason: "suspect_price"})

	assert.Equal(t, http.StatusNoContent, rec.Code)

	_, err := setup.taskRepo.FindByID(task.ID())
	assert.ErrorIs(t, err, domain.ErrModerationTaskNotFound)
}

func TestRejectClassifiedAd_InvalidReasonReturnsBadRequest(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	task := setup.seedClaimedTask(t, "ad-1", moderator)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/reject",
		RejectClassifiedAdRequest{ModeratorID: moderator.ID().String(), Reason: "bad_vibes"})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRejectClassifiedAd_NotOwnerReturnsForbidden(t *testing.T) {
	setup := setupHandlerTest(t)
	owner := setup.seedModerator(t, "Jane Doe")
	other := setup.seedModerator(t, "John Smith")
	task := setup.seedClaimedTask(t, "ad-1", owner)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/reject",
		RejectClassifiedAdRequest{ModeratorID: other.ID().String(), Reason: "suspect_price"})

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// --- POST /moderation/tasks/{id}/challenge ---

func TestChallengeClassifiedAd_SuccessDeletesTask(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	task := setup.seedClaimedTask(t, "ad-1", moderator)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/challenge",
		ChallengeClassifiedAdRequest{ModeratorID: moderator.ID().String(), Reason: "price_to_verify"})

	assert.Equal(t, http.StatusNoContent, rec.Code)

	_, err := setup.taskRepo.FindByID(task.ID())
	assert.ErrorIs(t, err, domain.ErrModerationTaskNotFound)
}

func TestChallengeClassifiedAd_InvalidReasonReturnsBadRequest(t *testing.T) {
	setup := setupHandlerTest(t)
	moderator := setup.seedModerator(t, "Jane Doe")
	task := setup.seedClaimedTask(t, "ad-1", moderator)

	rec := setup.do(t, http.MethodPost, "/moderation/tasks/"+task.ID().String()+"/challenge",
		ChallengeClassifiedAdRequest{ModeratorID: moderator.ID().String(), Reason: "suspect_price"})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
