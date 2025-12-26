package simulation

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/finviz/backend/internal/models"
)

const NumSimulations = 1000

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RunMonteCarlo performs Monte Carlo simulation on assets and debts
func RunMonteCarlo(assets []models.Asset, debts []models.Debt, years int) models.MonteCarloResponse {
	// Calculate starting net worth
	var totalAssets, totalDebts float64
	for _, a := range assets {
		totalAssets += a.CurrentValue
	}
	for _, d := range debts {
		totalDebts += d.CurrentBalance
	}
	startingNetWorth := totalAssets - totalDebts

	// Run simulations
	// results[year][simulation] = net worth value
	results := make([][]float64, years)
	for y := 0; y < years; y++ {
		results[y] = make([]float64, NumSimulations)
	}

	for sim := 0; sim < NumSimulations; sim++ {
		// Clone asset values for this simulation
		assetValues := make([]float64, len(assets))
		for i, a := range assets {
			assetValues[i] = a.CurrentValue
		}

		// Clone debt values
		debtValues := make([]float64, len(debts))
		for i, d := range debts {
			debtValues[i] = d.CurrentBalance
		}

		for year := 0; year < years; year++ {
			// Simulate asset growth
			for i, a := range assets {
				expectedReturn := a.GetReturn() / 100.0
				volatility := a.GetVolatility() / 100.0

				// Generate random return using normal distribution
				annualReturn := normalRandom(expectedReturn, volatility)
				assetValues[i] *= (1 + annualReturn)

				// Prevent negative asset values
				if assetValues[i] < 0 {
					assetValues[i] = 0
				}
			}

			// Simulate debt reduction (simple model: minimum payments)
			for i, d := range debts {
				if debtValues[i] > 0 {
					// Apply interest if rate > 0
					if d.InterestRate != nil && *d.InterestRate > 0 {
						monthlyRate := *d.InterestRate / 100.0 / 12.0
						for m := 0; m < 12; m++ {
							debtValues[i] *= (1 + monthlyRate)
							// Apply minimum payment if > 0
							if d.MinimumPayment != nil && *d.MinimumPayment > 0 {
								debtValues[i] -= *d.MinimumPayment
								if debtValues[i] < 0 {
									debtValues[i] = 0
								}
							}
						}
					}
				}
			}

			// Calculate net worth for this year
			var yearAssets, yearDebts float64
			for _, v := range assetValues {
				yearAssets += v
			}
			for _, v := range debtValues {
				yearDebts += v
			}
			results[year][sim] = yearAssets - yearDebts
		}
	}

	// Calculate percentiles for each year
	projections := make([]models.YearProjection, years)
	for year := 0; year < years; year++ {
		values := results[year]
		sort.Float64s(values)

		projections[year] = models.YearProjection{
			Year: year + 1,
			P10:  percentile(values, 10),
			P50:  percentile(values, 50),
			P90:  percentile(values, 90),
		}
	}

	// Final year percentiles
	finalValues := results[years-1]
	sort.Float64s(finalValues)

	return models.MonteCarloResponse{
		Projections: projections,
		Summary: models.ProjectionSummary{
			StartingNetWorth: startingNetWorth,
			FinalP10:         percentile(finalValues, 10),
			FinalP50:         percentile(finalValues, 50),
			FinalP90:         percentile(finalValues, 90),
			Years:            years,
			Simulations:      NumSimulations,
		},
	}
}

// normalRandom generates a random number from normal distribution
// using Box-Muller transform
func normalRandom(mean, stddev float64) float64 {
	u1 := rand.Float64()
	u2 := rand.Float64()

	// Box-Muller transform
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)

	return mean + stddev*z
}

// percentile calculates the p-th percentile of a sorted slice
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}
