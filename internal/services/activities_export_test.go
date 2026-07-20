package services

import (
	"context"
	"errors"
	"testing"

	"github.com/melonamin/quantlete/internal/storage"
)

func TestActivityServiceStreamExportRequiresAthleteID(t *testing.T) {
	service := NewActivityService(storage.NewActivityRepository(testDB(t)), nil, nil)
	writeCalled := false

	err := service.StreamExport(context.Background(), ExportActivitiesInput{}, func([]storage.Activity) error {
		writeCalled = true
		return nil
	})

	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("StreamExport() error = %v, want ErrUnauthorized", err)
	}
	if writeCalled {
		t.Fatal("StreamExport() called the page writer without an athlete ID")
	}
}
