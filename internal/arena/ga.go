package arena

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type DNA map[string]float64

type Species string

const (
	StatusPaper       = "paper"
	StatusWalkForward = "walk_forward"
	StatusLive        = "live"
	StatusHibernating = "hibernating"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateFingerprint generates the cryptographic ID of an organism
func GenerateFingerprint(species Species, marketFilter string, genes DNA) string {
	var keys []string
	for k := range genes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var geneStr string
	for _, k := range keys {
		geneStr += fmt.Sprintf("%s:%f,", k, genes[k])
	}

	raw := fmt.Sprintf("%s|%s|%s", species, marketFilter, geneStr)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// Evolve handles the main epoch: culling correlated/underperforming strategies, and crossing over new ones.
func (a *Arena) Evolve(topK int, newPopulationSize int) {
	if len(a.organisms) < 2 {
		return // Not enough to evolve
	}

	// 1. Calculate fitness (Adjusted Sortino)
	var activeOrgs []*Organism
	for _, org := range a.organisms {
		if org.Status == StatusHibernating || org.ActionCount < 50 {
			// Skip immature organisms or hibernating ones
			continue
		}

		// Calculate Sortino Ratio
		org.FitnessScore = CalculateSortino(org.ReturnStream, 0.0) // 0% risk free rate for simplicity

		// Apply Max Drawdown Penalty
		mdd := CalculateMaxDrawdown(org.EquityCurve)
		org.FitnessScore -= (mdd * 2.0) // Arbitrary heavy penalty for drawdown

		activeOrgs = append(activeOrgs, org)
	}

	// 2. Sort by initial fitness (Descending)
	sort.SliceStable(activeOrgs, func(i, j int) bool {
		return activeOrgs[i].FitnessScore > activeOrgs[j].FitnessScore
	})

	// 3. Apply Correlation Penalty (Ensemble Diversity)
	for i := 0; i < len(activeOrgs); i++ {
		for j := i + 1; j < len(activeOrgs); j++ {
			corr := PearsonCorrelation(activeOrgs[i].ReturnStream, activeOrgs[j].ReturnStream)
			if corr > 0.85 {
				// Highly correlated, penalize the lower-ranked one heavily
				activeOrgs[j].FitnessScore *= 0.5 // Cut score in half
			}
		}
	}

	// 4. Re-sort post-correlation penalty
	sort.SliceStable(activeOrgs, func(i, j int) bool {
		return activeOrgs[i].FitnessScore > activeOrgs[j].FitnessScore
	})

	// 5. Selection (Top K survive and get capital)
	survivors := activeOrgs
	if len(activeOrgs) > topK {
		survivors = activeOrgs[:topK]
	}

	// Adjust capital allocation
	totalFitness := 0.0
	for _, s := range survivors {
		if s.FitnessScore > 0 {
			totalFitness += s.FitnessScore
		}
	}

	for _, org := range a.organisms {
		org.CapitalAllocated = 0.0 // reset all
		if org.Status == StatusLive && org.FitnessScore <= 0 {
			org.Status = StatusHibernating
		}
	}

	if totalFitness > 0 {
		for _, s := range survivors {
			if s.FitnessScore > 0 {
				s.CapitalAllocated = 100.0 * (s.FitnessScore / totalFitness) // Spread $100 bankroll
				if s.Status == StatusPaper || s.Status == StatusWalkForward {
					s.Status = StatusLive
				}
			}
		}
	}

	// 6. Crossover & Mutation (Generate Offspring)
	if len(survivors) >= 2 {
		for i := 0; i < (newPopulationSize - len(survivors)); i++ {
			p1 := survivors[rand.Intn(len(survivors))]
			p2 := survivors[rand.Intn(len(survivors))]

			childGenes := make(DNA)
			for k, v1 := range p1.Genes {
				v2, ok := p2.Genes[k]
				if !ok {
					v2 = v1
				}
				// 50/50 Crossover
				if rand.Float64() > 0.5 {
					childGenes[k] = v1
				} else {
					childGenes[k] = v2
				}

				// Mutation (10% chance)
				if rand.Float64() < 0.10 {
					childGenes[k] *= (1.0 + (rand.Float64()*0.20 - 0.10)) // +/- 10% shift
				}
			}

			// Generate fingerprint
			// Assuming child belongs to p1's species for simplicity if mixed
			childID := GenerateFingerprint(p1.Species, p1.MarketFilter, childGenes)
			if _, exists := a.organisms[childID]; !exists {
				a.organisms[childID] = &Organism{
					ID:           childID,
					Species:      p1.Species,
					MarketFilter: p1.MarketFilter,
					Genes:        childGenes,
					Status:       StatusPaper,
					CreatedAt:    a.now(),
					EquityCurve:  []float64{100.0}, // Start at $100 baseline
				}
			}
		}
	}
}
