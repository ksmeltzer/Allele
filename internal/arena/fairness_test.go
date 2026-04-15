package arena

import (
	"testing"
	"time"
)

func TestArena_CalculateFitness_NoSidelining(t *testing.T) {
	arena := NewArena()

	// Mock time
	currentTime := time.Now()
	arena.now = func() time.Time { return currentTime }

	arena.AddOrganism("org1", "strategy", "all", nil)

	// Simulate 10 seconds of active time
	currentTime = currentTime.Add(10 * time.Second)

	score, err := arena.CalculateFitness("org1", 100.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expecting 100.0 / 10.0 = 10.0
	if score != 10.0 {
		t.Errorf("expected score 10.0, got %f", score)
	}
}

func TestArena_SidelineOrganism_CalculateFitness(t *testing.T) {
	arena := NewArena()

	currentTime := time.Now()
	arena.now = func() time.Time { return currentTime }

	arena.AddOrganism("org2", "strategy", "all", nil)

	// Active for 5 seconds
	currentTime = currentTime.Add(5 * time.Second)

	// Sideline
	err := arena.SidelineOrganism("org2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Sidelined for 15 seconds
	currentTime = currentTime.Add(15 * time.Second)

	// Calculate fitness while still sidelined
	score, err := arena.CalculateFitness("org2", 50.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total time 20 seconds, sidelined 15 seconds, active time 5 seconds.
	// 50.0 / 5.0 = 10.0
	if score != 10.0 {
		t.Errorf("expected score 10.0, got %f", score)
	}
}

func TestArena_ReactivateOrganism(t *testing.T) {
	arena := NewArena()

	currentTime := time.Now()
	arena.now = func() time.Time { return currentTime }

	arena.AddOrganism("org3", "strategy", "all", nil)

	// Active for 2 seconds
	currentTime = currentTime.Add(2 * time.Second)
	arena.SidelineOrganism("org3")

	// Sidelined for 10 seconds
	currentTime = currentTime.Add(10 * time.Second)
	arena.ReactivateOrganism("org3")

	// Active for 8 seconds
	currentTime = currentTime.Add(8 * time.Second)

	score, err := arena.CalculateFitness("org3", 20.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total active time: 2 + 8 = 10 seconds.
	// 20.0 / 10.0 = 2.0
	if score != 2.0 {
		t.Errorf("expected score 2.0, got %f", score)
	}
}

func TestArena_Errors(t *testing.T) {
	arena := NewArena()

	err := arena.SidelineOrganism("missing")
	if err != ErrOrganismNotFound {
		t.Errorf("expected %v, got %v", ErrOrganismNotFound, err)
	}

	err = arena.ReactivateOrganism("missing")
	if err != ErrOrganismNotFound {
		t.Errorf("expected %v, got %v", ErrOrganismNotFound, err)
	}

	_, err = arena.CalculateFitness("missing", 10.0)
	if err != ErrOrganismNotFound {
		t.Errorf("expected %v, got %v", ErrOrganismNotFound, err)
	}
}
