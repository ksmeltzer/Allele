package arena

import (
	"errors"
	"time"
)

var (
	ErrOrganismNotFound = errors.New("organism not found")
)

// Organism represents a genetic entity in the arena.
type Organism struct {
	ID                 string
	FitnessScore       float64
	IsSidelined        bool
	SidelinedAt        time.Time
	TotalSidelinedTime time.Duration
	CreatedAt          time.Time

	// GA Fields
	Status           string
	Species          Species
	MarketFilter     string
	Genes            DNA
	ActionCount      int
	ReturnStream     []float64
	EquityCurve      []float64
	CapitalAllocated float64
}

// Arena manages the lifecycle and fitness calculation of organisms.
type Arena struct {
	organisms map[string]*Organism
	now       func() time.Time // useful for testing
}

// NewArena creates a new Arena instance.
func NewArena() *Arena {
	return &Arena{
		organisms: make(map[string]*Organism),
		now:       time.Now,
	}
}

// AddOrganism adds a new organism to the arena.
func (a *Arena) AddOrganism(id string, species string, marketFilter string, genes map[string]float64) {
	a.organisms[id] = &Organism{
		ID:           id,
		CreatedAt:    a.now(),
		Status:       StatusPaper,
		Species:      Species(species),
		MarketFilter: marketFilter,
		Genes:        genes,
		EquityCurve:  []float64{100.0},
	}
}

// GetOrganism retrieves an organism by its ID.
func (a *Arena) GetOrganism(id string) (*Organism, error) {
	org, ok := a.organisms[id]
	if !ok {
		return nil, ErrOrganismNotFound
	}
	return org, nil
}

// SidelineOrganism pauses an organism's fitness calculation time accumulation.
func (a *Arena) SidelineOrganism(id string) error {
	org, err := a.GetOrganism(id)
	if err != nil {
		return err
	}
	if org.IsSidelined {
		return nil // already sidelined
	}
	org.IsSidelined = true
	org.SidelinedAt = a.now()
	return nil
}

// ReactivateOrganism resumes an organism's fitness calculation.
func (a *Arena) ReactivateOrganism(id string) error {
	org, err := a.GetOrganism(id)
	if err != nil {
		return err
	}
	if !org.IsSidelined {
		return nil // already active
	}
	org.IsSidelined = false
	org.TotalSidelinedTime += a.now().Sub(org.SidelinedAt)
	return nil
}

// RecordAction records an executed action and its resulting PnL (simulated or real).
func (a *Arena) RecordAction(id string, pnl float64) error {
	org, err := a.GetOrganism(id)
	if err != nil {
		return err
	}
	org.ActionCount++
	org.ReturnStream = append(org.ReturnStream, pnl)

	lastEquity := 100.0 // Default fallback
	if len(org.EquityCurve) > 0 {
		lastEquity = org.EquityCurve[len(org.EquityCurve)-1]
	}

	org.EquityCurve = append(org.EquityCurve, lastEquity*(1.0+pnl))
	return nil
}

// CalculateFitness calculates the time-discounted fitness score, ignoring sidelined time.
// This calculates a rate: rawScore / effective_active_time_in_seconds
func (a *Arena) CalculateFitness(id string, rawScore float64) (float64, error) {
	org, err := a.GetOrganism(id)
	if err != nil {
		return 0, err
	}

	currentTime := a.now()
	totalTime := currentTime.Sub(org.CreatedAt)
	sidelinedTime := org.TotalSidelinedTime

	if org.IsSidelined {
		sidelinedTime += currentTime.Sub(org.SidelinedAt)
	}

	activeTime := totalTime - sidelinedTime
	if activeTime <= 0 {
		org.FitnessScore = 0
		return 0, nil
	}

	// Calculate a time-discounted fitness score (e.g., Score per second active)
	org.FitnessScore = rawScore / activeTime.Seconds()
	return org.FitnessScore, nil
}
