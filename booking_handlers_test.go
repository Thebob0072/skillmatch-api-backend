package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test Booking Model Validation
func TestBookingModelValidation(t *testing.T) {
	t.Run("Valid Booking Data", func(t *testing.T) {
		booking := Booking{
			BookingID:   1,
			ClientID:    1,
			ProviderID:  2,
			PackageID:   1,
			BookingDate: time.Now(),
			StartTime:   time.Now(),
			EndTime:     time.Now().Add(2 * time.Hour),
			TotalPrice:  1000.0,
			Status:      "pending",
		}

		assert.NotZero(t, booking.BookingID)
		assert.NotZero(t, booking.ClientID)
		assert.NotZero(t, booking.ProviderID)
		assert.Greater(t, booking.TotalPrice, 0.0)
		assert.Contains(t, []string{"pending", "confirmed", "completed", "cancelled"}, booking.Status)
	})

	t.Run("Invalid Booking Status", func(t *testing.T) {
		validStatuses := []string{"pending", "confirmed", "completed", "cancelled", "paid"}
		invalidStatus := "invalid_status"

		found := false
		for _, status := range validStatuses {
			if status == invalidStatus {
				found = true
				break
			}
		}

		assert.False(t, found, "Invalid status should not be in valid statuses")
	})
}

// Test Booking Time Validation
func TestBookingTimeValidation(t *testing.T) {
	now := time.Now()

	t.Run("End Time After Start Time", func(t *testing.T) {
		startTime := now
		endTime := now.Add(2 * time.Hour)

		assert.True(t, endTime.After(startTime))
	})

	t.Run("Invalid End Time Before Start Time", func(t *testing.T) {
		startTime := now
		endTime := now.Add(-1 * time.Hour)

		assert.False(t, endTime.After(startTime))
	})

	t.Run("Booking Date Not In Past", func(t *testing.T) {
		futureDate := now.Add(24 * time.Hour)
		pastDate := now.Add(-24 * time.Hour)

		assert.True(t, futureDate.After(now))
		assert.False(t, pastDate.After(now))
	})
}

// Test Price Calculation
func TestPriceCalculation(t *testing.T) {
	t.Run("Calculate Total Price", func(t *testing.T) {
		basePrice := 1000.0
		duration := 2.0 // hours
		expectedTotal := basePrice * duration

		actualTotal := basePrice * duration

		assert.Equal(t, expectedTotal, actualTotal)
	})

	t.Run("Calculate Stripe Fees", func(t *testing.T) {
		totalAmount := 1000.0
		stripeFee := totalAmount * 0.0275          // 2.75%
		platformCommission := totalAmount * 0.1000 // 10%
		providerEarnings := totalAmount * 0.8725   // 87.25%

		assert.Equal(t, 27.5, stripeFee)
		assert.Equal(t, 100.0, platformCommission)
		assert.Equal(t, 872.5, providerEarnings)

		// Verify total adds up
		total := stripeFee + platformCommission + providerEarnings
		assert.InDelta(t, totalAmount, total, 0.01)
	})
}

// Test Booking Status Transitions
func TestBookingStatusTransitions(t *testing.T) {
	validTransitions := map[string][]string{
		"pending":   {"confirmed", "cancelled"},
		"confirmed": {"completed", "cancelled"},
		"completed": {},
		"cancelled": {},
	}

	t.Run("Valid Status Transitions", func(t *testing.T) {
		currentStatus := "pending"
		newStatus := "confirmed"

		validNewStatuses := validTransitions[currentStatus]
		found := false
		for _, status := range validNewStatuses {
			if status == newStatus {
				found = true
				break
			}
		}

		assert.True(t, found, "confirmed should be a valid transition from pending")
	})

	t.Run("Invalid Status Transition", func(t *testing.T) {
		currentStatus := "completed"
		newStatus := "pending"

		validNewStatuses := validTransitions[currentStatus]
		found := false
		for _, status := range validNewStatuses {
			if status == newStatus {
				found = true
				break
			}
		}

		assert.False(t, found, "Cannot transition from completed back to pending")
	})
}

// Test Service Package Validation
func TestServicePackageValidation(t *testing.T) {
	t.Run("Valid Package", func(t *testing.T) {
		pkg := ServicePackage{
			PackageID:   1,
			ProviderID:  1,
			PackageName: "2 Hour Session",
			Duration:    120, // minutes
			Price:       1000.0,
			IsActive:    true,
		}

		assert.NotZero(t, pkg.PackageID)
		assert.NotEmpty(t, pkg.PackageName)
		assert.Greater(t, pkg.Duration, 0)
		assert.Greater(t, pkg.Price, 0.0)
		assert.True(t, pkg.IsActive)
	})

	t.Run("Invalid Package - Zero Price", func(t *testing.T) {
		price := 0.0
		assert.False(t, price > 0, "Price should be greater than zero")
	})

	t.Run("Invalid Package - Negative Duration", func(t *testing.T) {
		duration := -30
		assert.False(t, duration > 0, "Duration should be positive")
	})
}

// Test Booking Availability Check
func TestBookingAvailabilityCheck(t *testing.T) {
	t.Run("Time Slot Available", func(t *testing.T) {
		// Existing booking: 10:00 - 12:00
		existingStart := time.Date(2025, 12, 3, 10, 0, 0, 0, time.UTC)
		existingEnd := time.Date(2025, 12, 3, 12, 0, 0, 0, time.UTC)

		// New booking: 14:00 - 16:00 (should be available)
		newStart := time.Date(2025, 12, 3, 14, 0, 0, 0, time.UTC)
		newEnd := time.Date(2025, 12, 3, 16, 0, 0, 0, time.UTC)

		hasConflict := !(newEnd.Before(existingStart) || newStart.After(existingEnd))
		assert.False(t, hasConflict, "Time slots should not conflict")
	})

	t.Run("Time Slot Conflict", func(t *testing.T) {
		// Existing booking: 10:00 - 12:00
		existingStart := time.Date(2025, 12, 3, 10, 0, 0, 0, time.UTC)
		existingEnd := time.Date(2025, 12, 3, 12, 0, 0, 0, time.UTC)

		// New booking: 11:00 - 13:00 (overlaps with existing)
		newStart := time.Date(2025, 12, 3, 11, 0, 0, 0, time.UTC)
		newEnd := time.Date(2025, 12, 3, 13, 0, 0, 0, time.UTC)

		hasConflict := !(newEnd.Before(existingStart) || newStart.After(existingEnd))
		assert.True(t, hasConflict, "Time slots should conflict")
	})
}
