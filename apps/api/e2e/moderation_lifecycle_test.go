package e2e

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modhttp "ddd-second-hand-marketplace/internal/moderation/adapter/driving/http"
)

// TestModeration_HappyPathApproval covers the happy path: submitting an ad
// creates a moderation task; once claimed and accepted, the ad is approved
// then published (chained) and becomes publicly visible.
func TestModeration_HappyPathApproval(t *testing.T) {
	srv := newTestServer(t)

	// Submission: the ad awaits moderation and is not publicly visible.
	adID := srv.submitAd(t, "seller@example.com", "supersecret")
	srv.assertPubliclyInvisible(t, adID)

	// A task was created in the shared moderation queue.
	task := srv.soleTask(t)
	assert.Equal(t, "pending", task.Status)
	assert.Equal(t, "Vélo hollandais", task.ClassifiedAdTitle)
	assert.Empty(t, task.ClaimedBy)

	detail := srv.taskDetail(t, task.ID)
	assert.Equal(t, adID, detail.ClassifiedAdID)

	// Alice claims the task: it stays in the queue, marked as hers.
	srv.claimTask(t, task.ID, srv.moderatorAlice)
	claimed := srv.soleTask(t)
	assert.Equal(t, "claimed", claimed.Status)
	assert.Equal(t, "Alice Martin", claimed.ClaimedBy)

	// The ad is still not publicly visible while under review.
	srv.assertPubliclyInvisible(t, adID)

	// Alice accepts: approved → published is chained synchronously.
	srv.acceptTask(t, task.ID, srv.moderatorAlice)

	// The completed task is physically deleted.
	assert.Empty(t, srv.listTasks(t).Tasks)

	// The ad is now publicly visible.
	srv.assertPubliclyVisible(t, adID)

	// The history recorded the full chain, in order, with the moderator and
	// the submission snapshot.
	assert.Equal(t, []string{"submitted", "approved", "published"}, srv.historyActions(t, adID))
	entries := srv.historyEntries(t, adID)
	require.NotNil(t, entries[0].Snapshot())
	assert.Equal(t, "Vélo hollandais", entries[0].Snapshot().Title)
	require.NotNil(t, entries[1].ModeratorID())
	assert.Equal(t, srv.moderatorAlice, *entries[1].ModeratorID())
}

// TestModeration_Rejection covers the rejection path: a rejected ad is
// automatically deleted (reason "rejected") and never becomes publicly
// visible.
func TestModeration_Rejection(t *testing.T) {
	srv := newTestServer(t)

	adID := srv.submitAd(t, "seller@example.com", "supersecret")
	task := srv.soleTask(t)

	srv.claimTask(t, task.ID, srv.moderatorAlice)
	srv.rejectTask(t, task.ID, srv.moderatorAlice, "suspect_price")

	// The completed task is gone and the ad never surfaced publicly.
	assert.Empty(t, srv.listTasks(t).Tasks)
	srv.assertPubliclyInvisible(t, adID)

	// History: submitted → rejected (with moderator and reason) → deleted
	// (automatic, reason "rejected").
	assert.Equal(t, []string{"submitted", "rejected", "deleted"}, srv.historyActions(t, adID))
	entries := srv.historyEntries(t, adID)
	require.NotNil(t, entries[1].ModeratorID())
	assert.Equal(t, srv.moderatorAlice, *entries[1].ModeratorID())
	require.NotNil(t, entries[1].Reason())
	assert.Equal(t, "suspect_price", *entries[1].Reason())
	require.NotNil(t, entries[2].Reason())
	assert.Equal(t, "rejected", *entries[2].Reason())
}

// TestModeration_ChallengeThenFixThenAccept covers the correction loop: a
// challenged ad is fixed by the seller (PUT), re-submitted through a brand
// new task, and finally accepted and published.
func TestModeration_ChallengeThenFixThenAccept(t *testing.T) {
	srv := newTestServer(t)

	sellerEmail := "seller@example.com"
	sellerPassword := "supersecret"
	adID := srv.submitAd(t, sellerEmail, sellerPassword)
	firstTask := srv.soleTask(t)

	// Alice challenges the ad.
	srv.claimTask(t, firstTask.ID, srv.moderatorAlice)
	srv.challengeTask(t, firstTask.ID, srv.moderatorAlice, "price_to_verify")

	// The task is completed (deleted) and the challenged ad stays invisible.
	assert.Empty(t, srv.listTasks(t).Tasks)
	srv.assertPubliclyInvisible(t, adID)

	// The seller was emailed about the required corrections.
	emails := srv.mailerSpy.GetSentSimpleEmails()
	require.Len(t, emails, 1)
	assert.Equal(t, sellerEmail, emails[0].To)

	// The seller corrects the ad: it goes back to submitted through a brand
	// new task (new ID), still invisible.
	srv.editAd(t, adID, sellerEmail, sellerPassword, "Vélo hollandais (prix corrigé)", 12000)
	secondTask := srv.soleTask(t)
	assert.NotEqual(t, firstTask.ID, secondTask.ID, "each re-submission creates a new task")
	assert.Equal(t, "pending", secondTask.Status)
	assert.Equal(t, "Vélo hollandais (prix corrigé)", secondTask.ClassifiedAdTitle)
	srv.assertPubliclyInvisible(t, adID)

	// Bob accepts the corrected version: the ad is published.
	srv.claimTask(t, secondTask.ID, srv.moderatorBob)
	srv.acceptTask(t, secondTask.ID, srv.moderatorBob)

	assert.Empty(t, srv.listTasks(t).Tasks)
	srv.assertPubliclyVisible(t, adID)

	// The published version carries the corrected content.
	searchBody := srv.searchAds(t)
	require.Len(t, searchBody.Items, 1)
	assert.Equal(t, "Vélo hollandais (prix corrigé)", searchBody.Items[0].Title)
	assert.Equal(t, int64(12000), searchBody.Items[0].PriceInCents)

	assert.Equal(t, []string{"submitted", "challenged", "submitted", "approved", "published"}, srv.historyActions(t, adID))
}

// TestModeration_ChallengeThenFixThenReject covers a challenge whose
// correction still does not pass moderation: the re-submitted ad is rejected
// and automatically deleted.
func TestModeration_ChallengeThenFixThenReject(t *testing.T) {
	srv := newTestServer(t)

	sellerEmail := "seller@example.com"
	sellerPassword := "supersecret"
	adID := srv.submitAd(t, sellerEmail, sellerPassword)

	firstTask := srv.soleTask(t)
	srv.claimTask(t, firstTask.ID, srv.moderatorAlice)
	srv.challengeTask(t, firstTask.ID, srv.moderatorAlice, "category_to_fix")

	srv.editAd(t, adID, sellerEmail, sellerPassword, "Vélo hollandais (corrigé)", 15000)

	secondTask := srv.soleTask(t)
	srv.claimTask(t, secondTask.ID, srv.moderatorBob)
	srv.rejectTask(t, secondTask.ID, srv.moderatorBob, "inappropriate_content")

	// The ad was deleted and never became visible.
	assert.Empty(t, srv.listTasks(t).Tasks)
	srv.assertPubliclyInvisible(t, adID)

	assert.Equal(t, []string{"submitted", "challenged", "submitted", "rejected", "deleted"}, srv.historyActions(t, adID))
	entries := srv.historyEntries(t, adID)
	require.NotNil(t, entries[3].ModeratorID())
	assert.Equal(t, srv.moderatorBob, *entries[3].ModeratorID())
	require.NotNil(t, entries[3].Reason())
	assert.Equal(t, "inappropriate_content", *entries[3].Reason())
}

// TestModeration_MultipleChallenges covers repeated correction rounds: the
// ad is challenged twice, edited twice, then finally accepted; the history
// keeps every round with its snapshot, moderator and reason.
func TestModeration_MultipleChallenges(t *testing.T) {
	srv := newTestServer(t)

	sellerEmail := "seller@example.com"
	sellerPassword := "supersecret"
	adID := srv.submitAd(t, sellerEmail, sellerPassword)

	// Round 1: Alice challenges the price.
	task1 := srv.soleTask(t)
	srv.claimTask(t, task1.ID, srv.moderatorAlice)
	srv.challengeTask(t, task1.ID, srv.moderatorAlice, "price_to_verify")
	srv.editAd(t, adID, sellerEmail, sellerPassword, "Vélo hollandais v2", 12000)

	// Round 2: Bob challenges the category.
	task2 := srv.soleTask(t)
	require.NotEqual(t, task1.ID, task2.ID)
	srv.claimTask(t, task2.ID, srv.moderatorBob)
	srv.challengeTask(t, task2.ID, srv.moderatorBob, "category_to_fix")
	srv.editAd(t, adID, sellerEmail, sellerPassword, "Vélo hollandais v3", 12000)

	// Round 3: Alice finally accepts.
	task3 := srv.soleTask(t)
	require.NotEqual(t, task2.ID, task3.ID)
	srv.claimTask(t, task3.ID, srv.moderatorAlice)
	srv.acceptTask(t, task3.ID, srv.moderatorAlice)

	srv.assertPubliclyVisible(t, adID)

	// The history contains every round, in order.
	assert.Equal(t, []string{
		"submitted",
		"challenged",
		"submitted",
		"challenged",
		"submitted",
		"approved",
		"published",
	}, srv.historyActions(t, adID))

	entries := srv.historyEntries(t, adID)

	// Each submission entry carries the snapshot of that version.
	require.NotNil(t, entries[0].Snapshot())
	assert.Equal(t, "Vélo hollandais", entries[0].Snapshot().Title)
	require.NotNil(t, entries[2].Snapshot())
	assert.Equal(t, "Vélo hollandais v2", entries[2].Snapshot().Title)
	require.NotNil(t, entries[4].Snapshot())
	assert.Equal(t, "Vélo hollandais v3", entries[4].Snapshot().Title)

	// Each challenge entry carries its moderator and reason.
	require.NotNil(t, entries[1].ModeratorID())
	assert.Equal(t, srv.moderatorAlice, *entries[1].ModeratorID())
	require.NotNil(t, entries[1].Reason())
	assert.Equal(t, "price_to_verify", *entries[1].Reason())
	require.NotNil(t, entries[3].ModeratorID())
	assert.Equal(t, srv.moderatorBob, *entries[3].ModeratorID())
	require.NotNil(t, entries[3].Reason())
	assert.Equal(t, "category_to_fix", *entries[3].Reason())

	// The seller was emailed once per challenge (the publication email sent
	// by the legacy consumer is not counted here).
	challengeEmails := 0
	for _, email := range srv.mailerSpy.GetSentSimpleEmails() {
		if strings.Contains(email.Title, "nécessite des corrections") {
			challengeEmails++
		}
	}
	assert.Equal(t, 2, challengeEmails)
}

// TestModeration_ClaimConcurrency covers the exclusive lock on tasks: when
// two moderators try to claim the same task, only the first succeeds and the
// second gets a 409 Conflict.
func TestModeration_ClaimConcurrency(t *testing.T) {
	srv := newTestServer(t)

	srv.submitAd(t, "seller@example.com", "supersecret")
	task := srv.soleTask(t)

	// Alice claims first.
	srv.claimTask(t, task.ID, srv.moderatorAlice)

	// Bob tries to claim the same task: conflict.
	resp := srv.doJSON(t, http.MethodPost, "/moderation/tasks/"+task.ID+"/claim", modhttp.ClaimModerationTaskRequest{ModeratorID: srv.moderatorBob})
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	errBody := decodeJSON[modhttp.ErrorResponse](t, resp)
	assert.NotEmpty(t, errBody.Error)

	// The task is still claimed by Alice.
	claimed := srv.soleTask(t)
	assert.Equal(t, "claimed", claimed.Status)
	assert.Equal(t, "Alice Martin", claimed.ClaimedBy)
}

// TestModeration_CompleteByNonOwner covers the ownership rule: only the
// moderator who claimed a task can complete it; anyone else gets a 403.
func TestModeration_CompleteByNonOwner(t *testing.T) {
	srv := newTestServer(t)

	adID := srv.submitAd(t, "seller@example.com", "supersecret")
	task := srv.soleTask(t)
	srv.claimTask(t, task.ID, srv.moderatorAlice)

	// Bob tries to complete Alice's task, in all three ways.
	acceptResp := srv.doJSON(t, http.MethodPost, "/moderation/tasks/"+task.ID+"/accept", modhttp.AcceptClassifiedAdRequest{ModeratorID: srv.moderatorBob})
	defer acceptResp.Body.Close()
	assert.Equal(t, http.StatusForbidden, acceptResp.StatusCode)

	rejectResp := srv.doJSON(t, http.MethodPost, "/moderation/tasks/"+task.ID+"/reject", modhttp.RejectClassifiedAdRequest{ModeratorID: srv.moderatorBob, Reason: "suspect_price"})
	defer rejectResp.Body.Close()
	assert.Equal(t, http.StatusForbidden, rejectResp.StatusCode)

	challengeResp := srv.doJSON(t, http.MethodPost, "/moderation/tasks/"+task.ID+"/challenge", modhttp.ChallengeClassifiedAdRequest{ModeratorID: srv.moderatorBob, Reason: "price_to_verify"})
	defer challengeResp.Body.Close()
	assert.Equal(t, http.StatusForbidden, challengeResp.StatusCode)

	// Nothing happened: the task is still claimed by Alice and the ad is
	// still under review.
	claimed := srv.soleTask(t)
	assert.Equal(t, "claimed", claimed.Status)
	assert.Equal(t, "Alice Martin", claimed.ClaimedBy)
	srv.assertPubliclyInvisible(t, adID)
	assert.Equal(t, []string{"submitted"}, srv.historyActions(t, adID))

	// Sanity check: the actual owner can still complete the task.
	srv.acceptTask(t, task.ID, srv.moderatorAlice)
	srv.assertPubliclyVisible(t, adID)
}

// TestModeration_FullHistory covers the history read model end to end: after
// a full journey (submit, challenge, edit, accept), GetModerationTaskDetail
// exposes every entry in chronological order with snapshots, moderator IDs
// and reasons, and the last snapshot reflects the latest edition.
func TestModeration_FullHistory(t *testing.T) {
	srv := newTestServer(t)

	sellerEmail := "seller@example.com"
	sellerPassword := "supersecret"

	adID := srv.submitAd(t, sellerEmail, sellerPassword)
	srv.clock.Advance(time.Minute)

	task1 := srv.soleTask(t)
	srv.claimTask(t, task1.ID, srv.moderatorAlice)
	srv.challengeTask(t, task1.ID, srv.moderatorAlice, "price_to_verify")
	srv.clock.Advance(time.Minute)

	srv.editAd(t, adID, sellerEmail, sellerPassword, "Vélo hollandais (prix corrigé)", 12000)
	srv.clock.Advance(time.Minute)

	task2 := srv.soleTask(t)
	srv.claimTask(t, task2.ID, srv.moderatorBob)

	// While the second task is claimed, its detail view exposes the full
	// history so far and the latest content snapshot.
	detail := srv.taskDetail(t, task2.ID)
	assert.Equal(t, adID, detail.ClassifiedAdID)
	assert.Equal(t, "claimed", detail.Status)
	assert.Equal(t, "Bob Dupont", detail.ClaimedBy)
	require.NotNil(t, detail.ClaimedAt)

	require.Len(t, detail.History, 3)
	assert.Equal(t, "submitted", detail.History[0].Action)
	require.NotNil(t, detail.History[0].Snapshot)
	assert.Equal(t, "Vélo hollandais", detail.History[0].Snapshot.Title)
	assert.Equal(t, int64(15000), detail.History[0].Snapshot.PriceInCents)
	assert.Equal(t, "seller@example.com", detail.History[0].Snapshot.SellerEmail)

	assert.Equal(t, "challenged", detail.History[1].Action)
	require.NotNil(t, detail.History[1].ModeratorID)
	assert.Equal(t, srv.moderatorAlice, *detail.History[1].ModeratorID)
	require.NotNil(t, detail.History[1].Reason)
	assert.Equal(t, "price_to_verify", *detail.History[1].Reason)

	assert.Equal(t, "submitted", detail.History[2].Action)
	require.NotNil(t, detail.History[2].Snapshot)
	assert.Equal(t, "Vélo hollandais (prix corrigé)", detail.History[2].Snapshot.Title)

	require.NotNil(t, detail.LastSnapshot)
	assert.Equal(t, "Vélo hollandais (prix corrigé)", detail.LastSnapshot.Title)
	assert.Equal(t, int64(12000), detail.LastSnapshot.PriceInCents)

	// Bob accepts: the task disappears, its detail endpoint returns 404.
	srv.clock.Advance(time.Minute)
	srv.acceptTask(t, task2.ID, srv.moderatorBob)

	detailAfterResp := srv.get(t, "/moderation/tasks/"+task2.ID)
	defer detailAfterResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, detailAfterResp.StatusCode)

	// The full history survives task deletion, in chronological order.
	assert.Equal(t, []string{"submitted", "challenged", "submitted", "approved", "published"}, srv.historyActions(t, adID))
	entries := srv.historyEntries(t, adID)
	for i := 1; i < len(entries); i++ {
		assert.False(t, entries[i].OccurredAt().Before(entries[i-1].OccurredAt()),
			"history entries should be in chronological order (entry %d before entry %d)", i, i-1)
	}
	require.NotNil(t, entries[3].ModeratorID())
	assert.Equal(t, srv.moderatorBob, *entries[3].ModeratorID())
}
