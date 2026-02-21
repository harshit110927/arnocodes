package course

import (
	"context"
	"testing"
)

type mockStatusProvider struct {
	status DiagnosticUserStatus
	err    error
}

func (m mockStatusProvider) GetUserStatus(ctx context.Context, userID string) (DiagnosticUserStatus, error) {
	_ = ctx
	_ = userID
	return m.status, m.err
}

func TestBuildCourseTopicsRootUnlocked(t *testing.T) {
	topics := []TopicRow{{TopicID: "t1", Name: "Arrays", CompletionStatus: "not_started", SubtopicCount: 1}}
	out := BuildCourseTopics(topics, nil)
	if out[0].UnlockStatus != "unlocked" {
		t.Fatalf("expected unlocked, got %s", out[0].UnlockStatus)
	}
}

func TestBuildCourseTopicsChildLockedWhenParentBelowThreshold(t *testing.T) {
	parentMastery := 79.0
	topics := []TopicRow{
		{TopicID: "p", Name: "Parent", MasteryScore: &parentMastery, CompletionStatus: "in_progress", SubtopicCount: 1},
		{TopicID: "c", Name: "Child", CompletionStatus: "not_started", SubtopicCount: 1},
	}
	prereqs := []TopicPrerequisite{{TopicID: "c", PrerequisiteID: "p"}}
	out := BuildCourseTopics(topics, prereqs)
	var child CourseTopic
	for _, trow := range out {
		if trow.ID == "c" {
			child = trow
		}
	}
	if child.UnlockStatus != "locked" {
		t.Fatalf("expected locked child, got %s", child.UnlockStatus)
	}
}

func TestBuildCourseTopicsChildUnlockedWhenParentAboveThreshold(t *testing.T) {
	parentMastery := 80.0
	topics := []TopicRow{
		{TopicID: "p", Name: "Parent", MasteryScore: &parentMastery, CompletionStatus: "completed", SubtopicCount: 1},
		{TopicID: "c", Name: "Child", CompletionStatus: "not_started", SubtopicCount: 1},
	}
	prereqs := []TopicPrerequisite{{TopicID: "c", PrerequisiteID: "p"}}
	out := BuildCourseTopics(topics, prereqs)
	var child CourseTopic
	for _, trow := range out {
		if trow.ID == "c" {
			child = trow
		}
	}
	if child.UnlockStatus != "unlocked" {
		t.Fatalf("expected unlocked child, got %s", child.UnlockStatus)
	}
}

func TestBuildCourseTopicsMultiPrerequisiteCase(t *testing.T) {
	m1 := 92.0
	m2 := 81.0
	topics := []TopicRow{
		{TopicID: "a", Name: "A", MasteryScore: &m1, CompletionStatus: "completed", SubtopicCount: 1},
		{TopicID: "b", Name: "B", MasteryScore: &m2, CompletionStatus: "completed", SubtopicCount: 1},
		{TopicID: "c", Name: "C", CompletionStatus: "not_started", SubtopicCount: 1},
	}
	prereqs := []TopicPrerequisite{{TopicID: "c", PrerequisiteID: "a"}, {TopicID: "c", PrerequisiteID: "b"}}
	out := BuildCourseTopics(topics, prereqs)
	var child CourseTopic
	for _, trow := range out {
		if trow.ID == "c" {
			child = trow
		}
	}
	if child.UnlockStatus != "unlocked" {
		t.Fatalf("expected unlocked with all prerequisites mastered, got %s", child.UnlockStatus)
	}
}

func TestBuildCourseTopicsMultiPrerequisiteLockedWhenAnyBelowThreshold(t *testing.T) {
	m1 := 90.0
	m2 := 79.0
	topics := []TopicRow{
		{TopicID: "a", Name: "A", MasteryScore: &m1, CompletionStatus: "completed", SubtopicCount: 1},
		{TopicID: "b", Name: "B", MasteryScore: &m2, CompletionStatus: "in_progress", SubtopicCount: 1},
		{TopicID: "c", Name: "C", CompletionStatus: "not_started", SubtopicCount: 1},
	}
	prereqs := []TopicPrerequisite{{TopicID: "c", PrerequisiteID: "a"}, {TopicID: "c", PrerequisiteID: "b"}}
	out := BuildCourseTopics(topics, prereqs)
	var child CourseTopic
	for _, trow := range out {
		if trow.ID == "c" {
			child = trow
		}
	}
	if child.UnlockStatus != "locked" {
		t.Fatalf("expected locked with one prerequisite below threshold, got %s", child.UnlockStatus)
	}
}

func TestBuildCourseTopicsCompletedOverridesUnlockState(t *testing.T) {
	mastery := 99.0
	topics := []TopicRow{{TopicID: "t1", Name: "Done", MasteryScore: &mastery, CompletionStatus: "in_progress", SubtopicCount: 1}}
	out := BuildCourseTopics(topics, nil)
	if out[0].UnlockStatus != "completed" {
		t.Fatalf("expected completed unlock status, got %s", out[0].UnlockStatus)
	}
}

func TestEnsureDiagnosticNotSubmittedDenied(t *testing.T) {
	svc := NewCourseService(&CourseRepository{}, mockStatusProvider{status: DiagnosticUserStatus{DiagnosticCompleted: false}})
	if err := svc.ensureDiagnosticSubmitted(context.Background(), "u1"); err == nil {
		t.Fatalf("expected diagnostic gating error")
	}
}

func TestUnlockReflectsUpdatedMasteryScore(t *testing.T) {
	m := 79.0
	topics := []TopicRow{
		{TopicID: "p", Name: "Parent", MasteryScore: &m, CompletionStatus: "in_progress", SubtopicCount: 1},
		{TopicID: "c", Name: "Child", CompletionStatus: "not_started", SubtopicCount: 1},
	}
	prereqs := []TopicPrerequisite{{TopicID: "c", PrerequisiteID: "p"}}
	out := BuildCourseTopics(topics, prereqs)
	if out[1].UnlockStatus != "locked" {
		t.Fatalf("expected locked before external mastery update")
	}
	m = 85.0
	out = BuildCourseTopics(topics, prereqs)
	if out[1].UnlockStatus != "unlocked" {
		t.Fatalf("expected unlocked after mastery update, got %s", out[1].UnlockStatus)
	}
}
