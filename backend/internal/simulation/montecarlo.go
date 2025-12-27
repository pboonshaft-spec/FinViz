package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/finviz/backend/internal/models"
)

const NumSimulations = 5000

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RunMonteCarlo performs enhanced Monte Carlo simulation with two-phase modeling
func RunMonteCarlo(assets []models.Asset, debts []models.Debt, years int) models.MonteCarloResponse {
	// Use default params for backward compatibility
	params := models.DefaultSimulationParams()
	params.TimeHorizonYears = years
	return RunMonteCarloWithParams(assets, debts, &params)
}

// RunMonteCarloWithParams performs Monte Carlo simulation with full parameter support
func RunMonteCarloWithParams(assets []models.Asset, debts []models.Debt, params *models.SimulationParams) models.MonteCarloResponse {
	// Apply defaults for any missing values
	params.ApplyDefaults()

	// Calculate starting net worth
	var totalAssets, totalDebts float64
	for _, a := range assets {
		totalAssets += a.CurrentValue
	}
	for _, d := range debts {
		totalDebts += d.CurrentBalance
	}
	startingNetWorth := totalAssets - totalDebts

	years := params.TimeHorizonYears
	retirementYear := params.RetirementAge - params.CurrentAge
	if retirementYear < 0 {
		retirementYear = 0
	}
	if retirementYear > years {
		retirementYear = years
	}

	// Track results per year per simulation
	// results[sim][year] = net worth
	results := make([][]float64, NumSimulations)
	contributions := make([][]float64, NumSimulations)
	withdrawals := make([][]float64, NumSimulations)

	for sim := 0; sim < NumSimulations; sim++ {
		results[sim] = make([]float64, years)
		contributions[sim] = make([]float64, years)
		withdrawals[sim] = make([]float64, years)
	}

	// Track success (didn't run out of money)
	successCount := 0
	accumulationWarningCount := 0

	for sim := 0; sim < NumSimulations; sim++ {
		// Initialize portfolio value
		portfolioValue := startingNetWorth

		// Clone debt values for this simulation
		debtValues := make([]float64, len(debts))
		for i, d := range debts {
			debtValues[i] = d.CurrentBalance
		}

		// Track cumulative contributions/withdrawals
		var totalContrib, totalWithdraw float64

		// Current monthly contribution (will grow with inflation)
		monthlyContrib := params.MonthlyContribution

		// Current monthly spending (will grow with inflation)
		monthlySpending := params.RetirementSpending

		// Track Social Security benefit with COLA adjustments (state variable)
		ssBenefitAnnual := params.SocialSecurityAmount * 12

		success := true
		accumulationWarning := false

		// Track portfolio value at start of retirement for "fixed" withdrawal strategy
		retirementStartingValue := 0.0

		for year := 0; year < years; year++ {
			age := params.CurrentAge + year
			isRetired := year >= retirementYear

			var yearContribution, yearWithdrawal float64

			if !isRetired {
				// ACCUMULATION PHASE

				// Calculate annual contribution with employer match
				annualContrib := monthlyContrib * 12
				employerMatch := calculateEmployerMatch(annualContrib, params.EmployerMatch, params.EmployerMatchLimit)
				totalAnnualContrib := annualContrib + employerMatch

				portfolioValue += totalAnnualContrib
				yearContribution = totalAnnualContrib
				totalContrib += totalAnnualContrib

				// Grow contribution for next year (salary increase)
				monthlyContrib *= (1 + params.ContributionGrowth)
			} else {
				// DISTRIBUTION PHASE

				// Capture portfolio value at start of retirement (first year of distribution)
				if retirementStartingValue == 0 {
					retirementStartingValue = portfolioValue
				}

				// Calculate withdrawal based on strategy
				yearWithdrawal = calculateWithdrawal(portfolioValue, monthlySpending*12, params.WithdrawalStrategy, retirementStartingValue)

				// Add Social Security if eligible
				ssAge := params.SocialSecurityAge
				if age >= ssAge && params.SocialSecurityAmount > 0 {
					// Apply COLA for years after start (not first year receiving)
					if age > ssAge {
						ssBenefitAnnual *= 1.025 // 2.5% average COLA
					}
					yearWithdrawal -= ssBenefitAnnual // Reduces needed withdrawal
				}

				// Add pension if any
				if params.PensionIncome > 0 {
					yearWithdrawal -= params.PensionIncome * 12
				}

				// Ensure withdrawal need is non-negative
				if yearWithdrawal < 0 {
					yearWithdrawal = 0
				}

				// Calculate gross withdrawal needed (pre-tax)
				// To have X after taxes at rate T, you need X / (1 - T) gross
				grossWithdrawal := yearWithdrawal
				if params.RetirementTaxRate > 0 && params.RetirementTaxRate < 1 {
					grossWithdrawal = yearWithdrawal / (1 - params.RetirementTaxRate)
				}

				// Check if portfolio can cover the withdrawal (success detection)
				if grossWithdrawal > portfolioValue {
					// Cannot cover required spending - this is a failure
					success = false
					// Withdraw whatever is available
					grossWithdrawal = portfolioValue
				}

				portfolioValue -= grossWithdrawal
				totalWithdraw += grossWithdrawal

				// Grow spending for inflation (for next year's calculation)
				monthlySpending *= (1 + params.InflationRate)
			}

			// Apply one-time events
			for _, event := range params.OneTimeEvents {
				if event.Year == year+1 || (event.Recurring && event.Year <= year+1) {
					portfolioValue += event.Amount // positive = income, negative = expense
				}
			}

			// Pay down debts (simplified: minimum payments)
			for i, d := range debts {
				if debtValues[i] > 0 {
					if d.InterestRate != nil && *d.InterestRate > 0 {
						monthlyRate := *d.InterestRate / 100.0 / 12.0
						for m := 0; m < 12; m++ {
							debtValues[i] *= (1 + monthlyRate)
							if d.MinimumPayment != nil && *d.MinimumPayment > 0 {
								payment := math.Min(*d.MinimumPayment, debtValues[i])
								debtValues[i] -= payment
								if !isRetired {
									yearContribution += payment // Count debt payments as contributions
									totalContrib += payment
								}
							}
						}
					}
				}
				if debtValues[i] < 0 {
					debtValues[i] = 0
				}
			}

			// Generate investment return
			annualReturn := normalRandom(params.ExpectedReturn, params.Volatility)

			// Apply return to portfolio (not debts)
			if portfolioValue > 0 {
				portfolioValue *= (1 + annualReturn)
			}

			// Prevent negative portfolio
			if portfolioValue < 0 {
				portfolioValue = 0
			}

			// Calculate total debt remaining
			var remainingDebt float64
			for _, v := range debtValues {
				remainingDebt += v
			}

			// Calculate net worth
			netWorth := portfolioValue - remainingDebt

			// Track accumulation phase warnings (negative net worth before retirement)
			if !isRetired && netWorth < 0 {
				accumulationWarning = true
			}

			// Store results
			results[sim][year] = netWorth
			contributions[sim][year] = yearContribution
			withdrawals[sim][year] = yearWithdrawal
		}

		if success {
			successCount++
		}
		if accumulationWarning {
			accumulationWarningCount++
		}
	}

	// Calculate percentiles for each year
	projections := make([]models.YearProjection, years)
	for year := 0; year < years; year++ {
		yearValues := make([]float64, NumSimulations)
		var totalContrib, totalWithdraw float64
		for sim := 0; sim < NumSimulations; sim++ {
			yearValues[sim] = results[sim][year]
			totalContrib += contributions[sim][year]
			totalWithdraw += withdrawals[sim][year]
		}
		sort.Float64s(yearValues)

		phase := "accumulation"
		if year >= retirementYear {
			phase = "distribution"
		}

		projections[year] = models.YearProjection{
			Year:          year + 1,
			Age:           params.CurrentAge + year + 1,
			P10:           percentile(yearValues, 10),
			P25:           percentile(yearValues, 25),
			P50:           percentile(yearValues, 50),
			P75:           percentile(yearValues, 75),
			P90:           percentile(yearValues, 90),
			Phase:         phase,
			Contributions: totalContrib / float64(NumSimulations),
			Withdrawals:   totalWithdraw / float64(NumSimulations),
		}
	}

	// Calculate final year statistics
	finalValues := make([]float64, NumSimulations)
	var totalContribSum, totalWithdrawSum float64
	for sim := 0; sim < NumSimulations; sim++ {
		finalValues[sim] = results[sim][years-1]
		for year := 0; year < years; year++ {
			totalContribSum += contributions[sim][year]
			totalWithdrawSum += withdrawals[sim][year]
		}
	}
	sort.Float64s(finalValues)

	successRate := float64(successCount) / float64(NumSimulations) * 100

	response := models.MonteCarloResponse{
		Projections: projections,
		Summary: models.ProjectionSummary{
			StartingNetWorth:     startingNetWorth,
			FinalP10:             percentile(finalValues, 10),
			FinalP25:             percentile(finalValues, 25),
			FinalP50:             percentile(finalValues, 50),
			FinalP75:             percentile(finalValues, 75),
			FinalP90:             percentile(finalValues, 90),
			Years:                years,
			Simulations:          NumSimulations,
			SuccessRate:          successRate,
			RetirementYear:       retirementYear,
			TotalContributions:   totalContribSum / float64(NumSimulations),
			TotalWithdrawals:     totalWithdrawSum / float64(NumSimulations),
			AccumulationWarnings: accumulationWarningCount,
		},
		Milestones: calculateMilestones(results, startingNetWorth),
		Insights:   generateInsights(params, startingNetWorth, successRate, projections),
	}

	return response
}

// calculateEmployerMatch calculates the employer 401k match
func calculateEmployerMatch(annualContrib, matchRate, matchLimit float64) float64 {
	if matchRate <= 0 {
		return 0
	}
	match := annualContrib * matchRate
	if matchLimit > 0 && match > matchLimit {
		match = matchLimit
	}
	return match
}

// calculateWithdrawal determines withdrawal amount based on strategy
func calculateWithdrawal(portfolioValue, desiredSpending float64, strategy string, initialValue float64) float64 {
	switch strategy {
	case "fixed":
		// Classic 4% rule - 4% of initial portfolio
		return initialValue * 0.04
	case "dynamic":
		// 4% of current portfolio value
		return portfolioValue * 0.04
	case "guardrails":
		// 4% with 3-5% guardrails based on performance
		if portfolioValue <= 0 {
			return 0 // Can't withdraw from empty portfolio
		}
		baseWithdrawal := initialValue * 0.04
		currentRate := baseWithdrawal / portfolioValue
		if currentRate < 0.03 {
			// Portfolio doing well, can withdraw more
			return portfolioValue * 0.03
		} else if currentRate > 0.05 {
			// Portfolio struggling, reduce withdrawal
			return portfolioValue * 0.05
		}
		return baseWithdrawal
	default:
		// Default to desired spending
		return desiredSpending
	}
}

// calculateMilestones identifies key financial milestones
func calculateMilestones(results [][]float64, startingNetWorth float64) []models.Milestone {
	milestones := []models.Milestone{}
	targets := []float64{100000, 250000, 500000, 1000000, 2000000, 5000000}

	for _, target := range targets {
		if target <= startingNetWorth {
			continue
		}

		reachedCount := 0
		yearSum := 0
		yearCount := 0

		for sim := 0; sim < NumSimulations; sim++ {
			for year := 0; year < len(results[sim]); year++ {
				if results[sim][year] >= target {
					reachedCount++
					yearSum += year + 1
					yearCount++
					break
				}
			}
		}

		if reachedCount > 0 {
			milestones = append(milestones, models.Milestone{
				Description:    formatCurrency(target) + " net worth",
				TargetAmount:   target,
				MedianYear:     yearSum / yearCount,
				ProbabilityPct: float64(reachedCount) / float64(NumSimulations) * 100,
			})
		}
	}

	return milestones
}

// generateInsights creates actionable recommendations
func generateInsights(params *models.SimulationParams, startingNetWorth, successRate float64, projections []models.YearProjection) []models.Insight {
	insights := []models.Insight{}

	// Success rate insights
	if successRate >= 90 {
		insights = append(insights, models.Insight{
			Type:    "success",
			Title:   "On Track",
			Message: "Your plan has a high probability of success. You're well-positioned for retirement.",
		})
	} else if successRate >= 75 {
		insights = append(insights, models.Insight{
			Type:    "info",
			Title:   "Good Progress",
			Message: "Your plan has a reasonable success rate. Consider small adjustments to improve certainty.",
		})
	} else if successRate >= 50 {
		insights = append(insights, models.Insight{
			Type:    "warning",
			Title:   "Needs Attention",
			Message: "Your success rate is below ideal. Consider increasing contributions or adjusting retirement age.",
		})
	} else {
		insights = append(insights, models.Insight{
			Type:    "warning",
			Title:   "High Risk",
			Message: "Your current plan has significant risk of running out of money. Consider major adjustments.",
		})
	}

	// Contribution insights
	if params.MonthlyContribution > 0 && params.EmployerMatch == 0 {
		insights = append(insights, models.Insight{
			Type:    "opportunity",
			Title:   "Employer Match",
			Message: "If your employer offers 401(k) matching, you may be leaving free money on the table.",
		})
	}

	// Social Security insights
	if params.SocialSecurityAmount == 0 && params.CurrentAge < 60 {
		insights = append(insights, models.Insight{
			Type:    "info",
			Title:   "Social Security",
			Message: "Consider adding estimated Social Security benefits for more accurate projections.",
		})
	}

	// Retirement age insights
	if params.RetirementAge < 62 && successRate < 80 {
		insights = append(insights, models.Insight{
			Type:    "opportunity",
			Title:   "Delay Retirement",
			Message: "Working 2-3 more years could significantly improve your success rate.",
		})
	}

	return insights
}

// formatCurrency formats a number as currency string
func formatCurrency(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("$%.0fM", amount/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("$%.0fK", amount/1000)
	}
	return fmt.Sprintf("$%.0f", amount)
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
