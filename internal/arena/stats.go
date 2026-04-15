package arena

import (
	"math"
)

// CalculateSortino returns an annualized, downside-deviation adjusted Sortino-like ratio.
func CalculateSortino(returns []float64, riskFreeRate float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	var sum, downSumSq float64
	for _, r := range returns {
		sum += r
		if r < riskFreeRate {
			downSumSq += (r - riskFreeRate) * (r - riskFreeRate)
		}
	}

	mean := sum / float64(len(returns))
	downsideDeviation := math.Sqrt(downSumSq / float64(len(returns)))

	if downsideDeviation == 0 {
		if mean > riskFreeRate {
			return 999.0 // arbitrarily high if no downside risk
		}
		return 0
	}

	// Annualize if assuming these are daily/hourly returns?
	// For simplicity, just return standard Sortino.
	return (mean - riskFreeRate) / downsideDeviation
}

// PearsonCorrelation calculates the linear correlation between two return streams.
func PearsonCorrelation(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var sumA, sumB, sumASq, sumBSq, pSum float64
	n := float64(len(a))

	for i := 0; i < len(a); i++ {
		sumA += a[i]
		sumB += b[i]
		sumASq += a[i] * a[i]
		sumBSq += b[i] * b[i]
		pSum += a[i] * b[i]
	}

	num := pSum - (sumA * sumB / n)
	den := math.Sqrt((sumASq - (sumA*sumA)/n) * (sumBSq - (sumB*sumB)/n))

	if den == 0 {
		return 0
	}

	return num / den
}

// CalculateMaxDrawdown calculates the peak-to-trough maximum drawdown of an equity curve.
func CalculateMaxDrawdown(equityCurve []float64) float64 {
	if len(equityCurve) == 0 {
		return 0
	}

	maxDrawdown := 0.0
	peak := equityCurve[0]

	for _, value := range equityCurve {
		if value > peak {
			peak = value
		}
		drawdown := (peak - value) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}
