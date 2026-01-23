package services

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/melonamin/quantlete/internal/storage"
)

// testDB creates an in-memory database with schema for testing.
func testDB(t *testing.T) *storage.DB {
	t.Helper()

	db, err := storage.OpenInMemory()
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Run migrations to create schema
	if err := db.Migrate(); err != nil {
		_ = db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// createTestAthlete creates a test athlete record for foreign key constraints.
func createTestAthlete(t *testing.T, db *storage.DB, athleteID int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO athletes (id, username, firstname, lastname)
		VALUES (?, 'testuser', 'Test', 'User')
	`, athleteID)
	if err != nil {
		t.Fatalf("failed to create test athlete: %v", err)
	}
}

// createTestGear creates a test gear record.
func createTestGear(t *testing.T, db *storage.DB, gearID string, athleteID int64, name string, retired bool) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO gear (id, athlete_id, name, is_primary, retired, distance, brand_name, model_name, description, source, created_at, updated_at)
		VALUES (?, ?, ?, 0, ?, 0, '', '', '', 'strava', datetime('now'), datetime('now'))
	`, gearID, athleteID, name, retired)
	if err != nil {
		t.Fatalf("failed to create test gear: %v", err)
	}
}

// createTestActivity creates a test activity referencing gear.
func createTestActivity(t *testing.T, db *storage.DB, activityID, athleteID int64, gearID string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO activities (id, athlete_id, name, sport_type, start_date, start_date_local, distance, moving_time, elapsed_time, gear_id)
		VALUES (?, ?, 'Test Activity', 'Ride', datetime('now'), datetime('now'), 1000, 100, 120, ?)
	`, activityID, athleteID, gearID)
	if err != nil {
		t.Fatalf("failed to create test activity: %v", err)
	}
}

func TestGearService_List(t *testing.T) {
	db := testDB(t)
	repo := storage.NewGearRepository(db)
	svc := NewGearService(repo)
	ctx := context.Background()

	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)

	// Create test gear
	createTestGear(t, db, "g1", athleteID, "Active Gear", false)
	createTestGear(t, db, "g2", athleteID, "Retired Gear", true)
	createTestGear(t, db, "g3", athleteID, "Another Active", false)

	tests := []struct {
		name           string
		input          ListGearInput
		wantCount      int
		wantTotalPages int
	}{
		{
			name: "active only",
			input: ListGearInput{
				AthleteID:      athleteID,
				IncludeRetired: false,
			},
			wantCount:      2,
			wantTotalPages: 1,
		},
		{
			name: "include retired",
			input: ListGearInput{
				AthleteID:      athleteID,
				IncludeRetired: true,
			},
			wantCount:      3,
			wantTotalPages: 1,
		},
		{
			name: "with pagination",
			input: ListGearInput{
				AthleteID:      athleteID,
				IncludeRetired: true,
				Page:           1,
				PerPage:        2,
			},
			wantCount:      2,
			wantTotalPages: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.List(ctx, tt.input)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if len(result.Data) != tt.wantCount {
				t.Errorf("List() count = %d, want %d", len(result.Data), tt.wantCount)
			}
			if result.TotalPages != tt.wantTotalPages {
				t.Errorf("List() TotalPages = %d, want %d", result.TotalPages, tt.wantTotalPages)
			}
		})
	}
}

func TestGearService_GetByID(t *testing.T) {
	db := testDB(t)
	repo := storage.NewGearRepository(db)
	svc := NewGearService(repo)
	ctx := context.Background()

	athleteID := int64(12345)
	otherAthleteID := int64(99999)
	createTestAthlete(t, db, athleteID)
	createTestAthlete(t, db, otherAthleteID)

	createTestGear(t, db, "g1", athleteID, "My Gear", false)
	createTestGear(t, db, "g2", otherAthleteID, "Other Gear", false)

	tests := []struct {
		name     string
		input    GetGearInput
		wantErr  error
		wantName string
	}{
		{
			name: "found",
			input: GetGearInput{
				AthleteID: athleteID,
				GearID:    "g1",
			},
			wantName: "My Gear",
		},
		{
			name: "not found",
			input: GetGearInput{
				AthleteID: athleteID,
				GearID:    "nonexistent",
			},
			wantErr: ErrNotFound,
		},
		{
			name: "forbidden - other athlete",
			input: GetGearInput{
				AthleteID: athleteID,
				GearID:    "g2",
			},
			wantErr: ErrForbidden,
		},
		{
			name: "bad request - empty ID",
			input: GetGearInput{
				AthleteID: athleteID,
				GearID:    "",
			},
			wantErr: ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.GetByID(ctx, tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("GetByID() expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetByID() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByID() unexpected error = %v", err)
			}
			if result.Name != tt.wantName {
				t.Errorf("GetByID() name = %q, want %q", result.Name, tt.wantName)
			}
		})
	}
}

func TestGearService_GetByID_WithActivityCount(t *testing.T) {
	db := testDB(t)
	repo := storage.NewGearRepository(db)
	svc := NewGearService(repo)
	ctx := context.Background()

	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)
	createTestGear(t, db, "g1", athleteID, "Bike", false)

	// Create activities referencing the gear
	createTestActivity(t, db, 1, athleteID, "g1")
	createTestActivity(t, db, 2, athleteID, "g1")
	createTestActivity(t, db, 3, athleteID, "g1")

	result, err := svc.GetByID(ctx, GetGearInput{
		AthleteID: athleteID,
		GearID:    "g1",
	})
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if result.ActivityCount != 3 {
		t.Errorf("GetByID() ActivityCount = %d, want 3", result.ActivityCount)
	}
}

func TestGearService_MonthlyUsage(t *testing.T) {
	db := testDB(t)
	repo := storage.NewGearRepository(db)
	svc := NewGearService(repo)
	ctx := context.Background()

	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)
	createTestGear(t, db, "g1", athleteID, "Bike", false)
	createTestActivity(t, db, 1, athleteID, "g1")

	result, err := svc.MonthlyUsage(ctx, MonthlyUsageInput{
		AthleteID:      athleteID,
		IncludeRetired: false,
	})
	if err != nil {
		t.Fatalf("MonthlyUsage() error = %v", err)
	}

	// Should have at least one month of usage data since we created an activity
	if len(result) == 0 {
		t.Error("MonthlyUsage() returned empty, expected at least one month of data with the created activity")
	}
}

func TestGearService_Validation(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		wantErr  error
	}{
		{
			name: "valid gear name",
			testFunc: func() error {
				return validateGearName("My Bike")
			},
			wantErr: nil,
		},
		{
			name: "empty gear name",
			testFunc: func() error {
				return validateGearName("")
			},
			wantErr: ErrBadRequest,
		},
		{
			name: "gear name too long",
			testFunc: func() error {
				longName := make([]byte, MaxGearNameLength+1)
				for i := range longName {
					longName[i] = 'a'
				}
				return validateGearName(string(longName))
			},
			wantErr: ErrBadRequest,
		},
		{
			name: "valid hashtag",
			testFunc: func() error {
				return validateHashtag("#myhashtag")
			},
			wantErr: nil,
		},
		{
			name: "valid hashtag without prefix",
			testFunc: func() error {
				return validateHashtag("myhashtag")
			},
			wantErr: nil,
		},
		{
			name: "empty hashtag",
			testFunc: func() error {
				return validateHashtag("")
			},
			wantErr: ErrBadRequest,
		},
		{
			name: "valid currency",
			testFunc: func() error {
				return validatePurchaseCurrency("USD")
			},
			wantErr: nil,
		},
		{
			name: "empty currency (optional)",
			testFunc: func() error {
				return validatePurchaseCurrency("")
			},
			wantErr: nil,
		},
		{
			name: "invalid currency length",
			testFunc: func() error {
				return validatePurchaseCurrency("US")
			},
			wantErr: ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error = %v", err)
			}
		})
	}
}

func TestGearService_ErrorTypes(t *testing.T) {
	// Verify error types can be checked with errors.Is
	err := NotFound("gear")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("NotFound() should wrap ErrNotFound")
	}

	err = BadRequest("invalid input")
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("BadRequest() should wrap ErrBadRequest")
	}

	err = Conflict("duplicate hashtag")
	if !errors.Is(err, ErrConflict) {
		t.Errorf("Conflict() should wrap ErrConflict")
	}
}

func TestGearService_UpdateGearPrice(t *testing.T) {
	db := testDB(t)
	repo := storage.NewGearRepository(db)
	svc := NewGearService(repo)
	ctx := context.Background()

	athleteID := int64(12345)
	otherAthleteID := int64(99999)
	createTestAthlete(t, db, athleteID)
	createTestAthlete(t, db, otherAthleteID)

	// Create test gear for this athlete and another athlete
	createTestGear(t, db, "g1", athleteID, "My Bike", false)
	createTestGear(t, db, "g2", otherAthleteID, "Other Bike", false)

	t.Run("successful update", func(t *testing.T) {
		price := 499.99
		currency := "USD"
		result, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:        athleteID,
			GearID:           "g1",
			PurchasePrice:    &price,
			PurchaseCurrency: &currency,
		})
		if err != nil {
			t.Fatalf("UpdateGearPrice() error = %v", err)
		}
		if result.PurchasePrice == nil || *result.PurchasePrice != 499.99 {
			t.Errorf("UpdateGearPrice() price = %v, want 499.99", result.PurchasePrice)
		}
		if result.PurchaseCurrency != "USD" {
			t.Errorf("UpdateGearPrice() currency = %q, want USD", result.PurchaseCurrency)
		}
	})

	t.Run("currency normalization to uppercase", func(t *testing.T) {
		price := 299.99
		currency := "eur" // lowercase
		result, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:        athleteID,
			GearID:           "g1",
			PurchasePrice:    &price,
			PurchaseCurrency: &currency,
		})
		if err != nil {
			t.Fatalf("UpdateGearPrice() error = %v", err)
		}
		if result.PurchaseCurrency != "EUR" {
			t.Errorf("UpdateGearPrice() currency = %q, want EUR (should normalize to uppercase)", result.PurchaseCurrency)
		}
	})

	t.Run("clear price with nil", func(t *testing.T) {
		result, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:        athleteID,
			GearID:           "g1",
			PurchasePrice:    nil,
			PurchaseCurrency: nil,
		})
		if err != nil {
			t.Fatalf("UpdateGearPrice() error = %v", err)
		}
		if result.PurchasePrice != nil {
			t.Errorf("UpdateGearPrice() price = %v, want nil", result.PurchasePrice)
		}
	})

	t.Run("authorization - other athlete gear returns not found", func(t *testing.T) {
		price := 100.0
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:     athleteID,
			GearID:        "g2", // belongs to otherAthleteID
			PurchasePrice: &price,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for other athlete's gear")
		}
		// Should return NotFound (not Forbidden) to avoid information disclosure
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("validation - negative price", func(t *testing.T) {
		price := -10.0
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:     athleteID,
			GearID:        "g1",
			PurchasePrice: &price,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for negative price")
		}
		if !errors.Is(err, ErrBadRequest) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrBadRequest", err)
		}
	})

	t.Run("validation - price exceeds max", func(t *testing.T) {
		price := 1000000.00 // exceeds MaxGearPrice (999999.99)
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:     athleteID,
			GearID:        "g1",
			PurchasePrice: &price,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for price exceeding max")
		}
		if !errors.Is(err, ErrBadRequest) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrBadRequest", err)
		}
	})

	t.Run("validation - infinity price rejected", func(t *testing.T) {
		price := math.Inf(1)
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:     athleteID,
			GearID:        "g1",
			PurchasePrice: &price,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for infinity price")
		}
		if !errors.Is(err, ErrBadRequest) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrBadRequest", err)
		}
	})

	t.Run("validation - NaN price rejected", func(t *testing.T) {
		price := math.NaN()
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:     athleteID,
			GearID:        "g1",
			PurchasePrice: &price,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for NaN price")
		}
		if !errors.Is(err, ErrBadRequest) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrBadRequest", err)
		}
	})

	t.Run("validation - invalid currency length", func(t *testing.T) {
		price := 100.0
		currency := "US" // only 2 chars, not 3
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:        athleteID,
			GearID:           "g1",
			PurchasePrice:    &price,
			PurchaseCurrency: &currency,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for invalid currency")
		}
		if !errors.Is(err, ErrBadRequest) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrBadRequest", err)
		}
	})

	t.Run("validation - empty gear ID", func(t *testing.T) {
		price := 100.0
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:     athleteID,
			GearID:        "",
			PurchasePrice: &price,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for empty gear ID")
		}
		if !errors.Is(err, ErrBadRequest) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrBadRequest", err)
		}
	})

	t.Run("not found - nonexistent gear", func(t *testing.T) {
		price := 100.0
		_, err := svc.UpdateGearPrice(ctx, UpdateGearPriceInput{
			AthleteID:     athleteID,
			GearID:        "nonexistent",
			PurchasePrice: &price,
		})
		if err == nil {
			t.Fatal("UpdateGearPrice() expected error for nonexistent gear")
		}
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("UpdateGearPrice() error = %v, want ErrNotFound", err)
		}
	})
}
