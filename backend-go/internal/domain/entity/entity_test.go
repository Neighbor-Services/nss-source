package entity_test

import (
	"testing"
	"time"

	"backend-go/internal/domain/entity"
)

func TestBackgroundCheck_ComputeTrialFields(t *testing.T) {
	now := time.Now()

	t.Run("clear status within 30 days is in trial", func(t *testing.T) {
		cleared10DaysAgo := now.Add(-10 * 24 * time.Hour)
		bc := &entity.BackgroundCheck{
			Status:    "clear",
			IsClear:   true,
			ClearedAt: &cleared10DaysAgo,
		}
		bc.ComputeTrialFields()

		if !bc.IsInTrial {
			t.Errorf("expected IsInTrial to be true, got false")
		}
		if bc.TrialDaysLeft <= 0 || bc.TrialDaysLeft > 21 {
			t.Errorf("expected TrialDaysLeft around 20, got %d", bc.TrialDaysLeft)
		}
		if bc.TrialEndsAt == nil {
			t.Fatalf("expected TrialEndsAt to be set, got nil")
		}
		expectedEnd := cleared10DaysAgo.Add(30 * 24 * time.Hour)
		if bc.TrialEndsAt.Sub(expectedEnd).Abs() > time.Second {
			t.Errorf("expected TrialEndsAt %v, got %v", expectedEnd, *bc.TrialEndsAt)
		}
	})

	t.Run("clear status past 30 days is not in trial", func(t *testing.T) {
		cleared40DaysAgo := now.Add(-40 * 24 * time.Hour)
		bc := &entity.BackgroundCheck{
			Status:    "clear",
			IsClear:   true,
			ClearedAt: &cleared40DaysAgo,
		}
		bc.ComputeTrialFields()

		if bc.IsInTrial {
			t.Errorf("expected IsInTrial to be false, got true")
		}
		if bc.TrialDaysLeft != 0 {
			t.Errorf("expected TrialDaysLeft to be 0, got %d", bc.TrialDaysLeft)
		}
	})

	t.Run("clear status with nil ClearedAt is not in trial", func(t *testing.T) {
		bc := &entity.BackgroundCheck{
			Status:    "clear",
			IsClear:   true,
			ClearedAt: nil,
		}
		bc.ComputeTrialFields()

		if bc.IsInTrial {
			t.Errorf("expected IsInTrial to be false, got true")
		}
		if bc.TrialDaysLeft != 0 {
			t.Errorf("expected TrialDaysLeft to be 0, got %d", bc.TrialDaysLeft)
		}
	})

	t.Run("consider status with ClearedAt is not in trial", func(t *testing.T) {
		cleared10DaysAgo := now.Add(-10 * 24 * time.Hour)
		bc := &entity.BackgroundCheck{
			Status:    "consider",
			IsClear:   false,
			ClearedAt: &cleared10DaysAgo,
		}
		bc.ComputeTrialFields()

		if bc.IsInTrial {
			t.Errorf("expected IsInTrial to be false, got true")
		}
		if bc.TrialDaysLeft != 0 {
			t.Errorf("expected TrialDaysLeft to be 0, got %d", bc.TrialDaysLeft)
		}
	})
}

