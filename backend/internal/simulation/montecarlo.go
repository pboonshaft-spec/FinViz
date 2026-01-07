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

// SimulationTracker holds per-simulation tracking data for enhanced metrics
type SimulationTracker struct {
	NetWorth    []float64 // Net worth at each year
	Returns     []float64 // Annual returns for each year
	FailureYear int       // Year of failure (-1 if successful)
	Success     bool      // Did the simulation succeed?
	PeakValue   float64   // Highest portfolio value reached
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

	// Enhanced tracking for advanced metrics
	simTrackers := make([]SimulationTracker, NumSimulations)

	for sim := 0; sim < NumSimulations; sim++ {
		results[sim] = make([]float64, years)
		contributions[sim] = make([]float64, years)
		withdrawals[sim] = make([]float64, years)
		// Initialize enhanced tracker
		simTrackers[sim] = SimulationTracker{
			NetWorth:    make([]float64, years),
			Returns:     make([]float64, years),
			FailureYear: -1, // -1 means no failure
			Success:     true,
			PeakValue:   startingNetWorth,
		}
	}

	// Track success (didn't run out of money)
	successCount := 0
	accumulationWarningCount := 0

	// Determine if this is an accumulation-only simulation
	isAccumulationOnly := retirementYear >= years

	for sim := 0; sim < NumSimulations; sim++ {
		// Initialize portfolio value
		portfolioValue := startingNetWorth
		peakValue := startingNetWorth

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

		// Track final net worth for accumulation-only success calculation
		var finalNetWorth float64

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
			var annualReturn float64
			if params.EnableGlidePath {
				// Use age-adjusted return and volatility (target-date style)
				glideReturn, glideVolatility := calculateGlidePathParams(age, params.RetirementAge)
				annualReturn = normalRandom(glideReturn, glideVolatility)
			} else {
				// Use static return and volatility
				annualReturn = normalRandom(params.ExpectedReturn, params.Volatility)
			}

			// Track the return for sequence analysis
			simTrackers[sim].Returns[year] = annualReturn

			// Apply return to portfolio (not debts)
			if portfolioValue > 0 {
				portfolioValue *= (1 + annualReturn)
			}

			// Prevent negative portfolio
			if portfolioValue < 0 {
				portfolioValue = 0
			}

			// Track peak value for drawdown analysis
			if portfolioValue > peakValue {
				peakValue = portfolioValue
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

			// Store in enhanced tracker
			simTrackers[sim].NetWorth[year] = netWorth

			// Track failure year (first year we couldn't meet obligations)
			if !success && simTrackers[sim].FailureYear == -1 {
				simTrackers[sim].FailureYear = year
			}

			// Track final net worth
			finalNetWorth = netWorth
		}

		// For accumulation-only simulations, success means ending with positive net worth
		if isAccumulationOnly {
			if finalNetWorth <= 0 {
				success = false
			}
		}

		// Store final tracker state
		simTrackers[sim].Success = success
		simTrackers[sim].PeakValue = peakValue

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

	// Calculate enhanced metrics
	enhancedMetrics := calculateEnhancedMetrics(simTrackers, params, retirementYear, years)

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
			EnhancedMetrics:      enhancedMetrics,
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

// calculateGlidePathParams returns age-adjusted return and volatility
// Simulates target-date fund behavior: high stocks when young, shifting to bonds near retirement
func calculateGlidePathParams(age, retirementAge int) (expectedReturn, volatility float64) {
	// Asset class assumptions
	stockReturn := 0.07     // 7% expected return for stocks
	bondReturn := 0.03      // 3% expected return for bonds
	stockVolatility := 0.18 // 18% volatility for stocks
	bondVolatility := 0.05  // 5% volatility for bonds

	// Calculate years to retirement
	yearsToRetirement := retirementAge - age
	if yearsToRetirement < 0 {
		yearsToRetirement = 0
	}

	// Glide path: 90% stocks at 40+ years out, linearly decreasing to 40% at retirement
	maxYears := 40.0
	var stockAllocation float64
	if float64(yearsToRetirement) >= maxYears {
		stockAllocation = 0.90 // Max 90% stocks
	} else {
		// Linear interpolation from 90% to 40%
		progress := float64(yearsToRetirement) / maxYears
		stockAllocation = 0.40 + (0.50 * progress) // 40% base + up to 50% more
	}

	// Blend return and volatility based on allocation
	expectedReturn = (stockAllocation * stockReturn) + ((1 - stockAllocation) * bondReturn)
	volatility = (stockAllocation * stockVolatility) + ((1 - stockAllocation) * bondVolatility)

	return expectedReturn, volatility
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

// calculateEnhancedMetrics computes all enhanced success metrics
func calculateEnhancedMetrics(trackers []SimulationTracker, params *models.SimulationParams, retirementYear, years int) *models.EnhancedMetrics {
	if len(trackers) == 0 || years == 0 {
		return nil
	}

	// Calculate median wealth at end
	finalWealth := make([]float64, len(trackers))
	for i, t := range trackers {
		if len(t.NetWorth) > 0 {
			finalWealth[i] = t.NetWorth[len(t.NetWorth)-1]
		}
	}
	sort.Float64s(finalWealth)
	medianWealthAtEnd := percentile(finalWealth, 50)

	// Calculate ruin probabilities at various ages
	ruinProbabilities := calculateRuinProbabilities(trackers, params, years)

	// Calculate safe floor (minimum guaranteed wealth at worst point)
	safeFloor := calculateSafeFloor(trackers, params, years)

	// Calculate recovery metrics
	recoveryMetrics := calculateRecoveryMetrics(trackers)

	// Calculate partial success rate (made it >50% of retirement years)
	partialSuccessRate := calculatePartialSuccessRate(trackers, retirementYear, years)

	// Calculate median years to ruin for failed simulations
	medianYearsToRuin, wealthAtRuin := calculateRuinStats(trackers, retirementYear)

	// Calculate sequence of returns analysis
	sequenceAnalysis := calculateSequenceAnalysis(trackers, params, years)

	return &models.EnhancedMetrics{
		MedianWealthAtEnd:  medianWealthAtEnd,
		RuinProbabilities:  ruinProbabilities,
		SafeFloor:          safeFloor,
		RecoveryMetrics:    recoveryMetrics,
		PartialSuccessRate: partialSuccessRate,
		MedianYearsToRuin:  medianYearsToRuin,
		WealthAtRuin:       wealthAtRuin,
		SequenceAnalysis:   sequenceAnalysis,
	}
}

// calculateRuinProbabilities computes probability of ruin at key ages
func calculateRuinProbabilities(trackers []SimulationTracker, params *models.SimulationParams, years int) []models.RuinProbability {
	// Key ages to check: 70, 75, 80, 85, 90, 95
	checkAges := []int{70, 75, 80, 85, 90, 95}
	results := []models.RuinProbability{}

	for _, age := range checkAges {
		yearIndex := age - params.CurrentAge - 1 // Convert age to year index (0-based)
		if yearIndex < 0 || yearIndex >= years {
			continue
		}

		failedCount := 0
		for _, t := range trackers {
			// Check if simulation had failed by this year
			if t.FailureYear != -1 && t.FailureYear <= yearIndex {
				failedCount++
			}
		}

		probability := float64(failedCount) / float64(len(trackers)) * 100

		results = append(results, models.RuinProbability{
			Age:         age,
			Probability: probability,
			YearsOut:    yearIndex + 1,
		})
	}

	return results
}

// calculateSafeFloor finds the 5th percentile minimum wealth across all years
func calculateSafeFloor(trackers []SimulationTracker, params *models.SimulationParams, years int) models.SafeFloor {
	worstFloor := math.MaxFloat64
	floorYear := 0
	floorAge := params.CurrentAge

	// For each year, find the 5th percentile and track the worst
	for year := 0; year < years; year++ {
		yearValues := make([]float64, len(trackers))
		for i, t := range trackers {
			if year < len(t.NetWorth) {
				yearValues[i] = t.NetWorth[year]
			}
		}
		sort.Float64s(yearValues)

		p5 := percentile(yearValues, 5)
		if p5 < worstFloor {
			worstFloor = p5
			floorYear = year + 1
			floorAge = params.CurrentAge + year + 1
		}
	}

	if worstFloor == math.MaxFloat64 {
		worstFloor = 0
	}

	description := fmt.Sprintf("At age %d (year %d), there's a 95%% chance your portfolio will be at least %s",
		floorAge, floorYear, formatCurrency(worstFloor))

	return models.SafeFloor{
		GuaranteedMinimum: worstFloor,
		FloorYear:         floorYear,
		FloorAge:          floorAge,
		Description:       description,
	}
}

// calculateRecoveryMetrics analyzes drawdown recovery patterns
func calculateRecoveryMetrics(trackers []SimulationTracker) models.RecoveryAnalysis {
	totalDrawdowns := 0
	totalRecoveryYears := 0.0
	recoveredCount := 0
	worstDrawdown := 0.0

	for _, t := range trackers {
		if len(t.NetWorth) < 2 {
			continue
		}

		// Track peak and current value for drawdown detection
		peak := t.NetWorth[0]
		inDrawdown := false
		drawdownStart := 0
		currentDrawdownPct := 0.0

		for year := 1; year < len(t.NetWorth); year++ {
			value := t.NetWorth[year]

			// Update peak
			if value > peak && !inDrawdown {
				peak = value
			}

			// Calculate current drawdown percentage
			if peak > 0 {
				drawdownPct := (peak - value) / peak
				if drawdownPct > currentDrawdownPct {
					currentDrawdownPct = drawdownPct
				}

				// Detect start of significant drawdown (20%+)
				if drawdownPct >= 0.20 && !inDrawdown {
					inDrawdown = true
					drawdownStart = year
					totalDrawdowns++

					if drawdownPct > worstDrawdown {
						worstDrawdown = drawdownPct
					}
				}

				// Detect recovery (back to previous peak)
				if inDrawdown && value >= peak {
					inDrawdown = false
					recoveryYears := float64(year - drawdownStart)
					totalRecoveryYears += recoveryYears
					recoveredCount++
					peak = value // Reset peak
					currentDrawdownPct = 0
				}
			}
		}

		// Track worst drawdown even if not recovered
		if currentDrawdownPct > worstDrawdown {
			worstDrawdown = currentDrawdownPct
		}
	}

	avgRecoveryYears := 0.0
	if recoveredCount > 0 {
		avgRecoveryYears = totalRecoveryYears / float64(recoveredCount)
	}

	avgDrawdownCount := 0
	recoverySuccessRate := 0.0
	if len(trackers) > 0 {
		avgDrawdownCount = totalDrawdowns / len(trackers)
		if totalDrawdowns > 0 {
			recoverySuccessRate = float64(recoveredCount) / float64(totalDrawdowns) * 100
		}
	}

	return models.RecoveryAnalysis{
		AvgRecoveryYears:    avgRecoveryYears,
		WorstDrawdown:       worstDrawdown * 100, // Convert to percentage
		DrawdownCount:       avgDrawdownCount,
		RecoverySuccessRate: recoverySuccessRate,
	}
}

// calculatePartialSuccessRate computes % of simulations that survived >50% of retirement
func calculatePartialSuccessRate(trackers []SimulationTracker, retirementYear, years int) float64 {
	if retirementYear >= years {
		// Accumulation-only simulation
		return 0
	}

	retirementYears := years - retirementYear
	halfwayYear := retirementYear + (retirementYears / 2)
	partialSuccessCount := 0

	for _, t := range trackers {
		// Partial success if failure happened after halfway through retirement (or no failure)
		if t.FailureYear == -1 || t.FailureYear > halfwayYear {
			partialSuccessCount++
		}
	}

	return float64(partialSuccessCount) / float64(len(trackers)) * 100
}

// calculateRuinStats computes median years to ruin and wealth at ruin for failed simulations
func calculateRuinStats(trackers []SimulationTracker, retirementYear int) (medianYearsToRuin float64, wealthAtRuin float64) {
	failedYears := []float64{}
	wealthShortfalls := []float64{}

	for _, t := range trackers {
		if !t.Success && t.FailureYear != -1 {
			// Years from retirement start to ruin
			yearsToRuin := t.FailureYear - retirementYear
			if yearsToRuin < 0 {
				yearsToRuin = 0
			}
			failedYears = append(failedYears, float64(yearsToRuin))

			// Wealth at failure (typically 0 or near 0)
			if t.FailureYear < len(t.NetWorth) {
				wealthShortfalls = append(wealthShortfalls, t.NetWorth[t.FailureYear])
			}
		}
	}

	if len(failedYears) == 0 {
		return 0, 0
	}

	sort.Float64s(failedYears)
	sort.Float64s(wealthShortfalls)

	medianYearsToRuin = percentile(failedYears, 50)
	wealthAtRuin = percentile(wealthShortfalls, 50)

	return medianYearsToRuin, wealthAtRuin
}

// simResult holds per-simulation data for sequence analysis
type simResult struct {
	earlyReturn float64
	success     bool
	finalWealth float64
}

// calculateSequenceAnalysis examines the impact of return sequences on outcomes
func calculateSequenceAnalysis(trackers []SimulationTracker, params *models.SimulationParams, years int) *models.SequenceAnalysis {
	if len(trackers) == 0 || years < 10 {
		return nil
	}

	// Calculate early returns (first 10 years) for each simulation
	results := make([]simResult, len(trackers))
	for i, t := range trackers {
		// Calculate average return for first 10 years (or available years)
		earlyYears := 10
		if earlyYears > len(t.Returns) {
			earlyYears = len(t.Returns)
		}

		var sumReturns float64
		for y := 0; y < earlyYears; y++ {
			sumReturns += t.Returns[y]
		}
		avgEarlyReturn := sumReturns / float64(earlyYears)

		finalWealth := 0.0
		if len(t.NetWorth) > 0 {
			finalWealth = t.NetWorth[len(t.NetWorth)-1]
		}

		results[i] = simResult{
			earlyReturn:  avgEarlyReturn,
			success:      t.Success,
			finalWealth:  finalWealth,
		}
	}

	// Sort by early returns to find worst and best first decades
	sort.Slice(results, func(i, j int) bool {
		return results[i].earlyReturn < results[j].earlyReturn
	})

	// Worst first decade (bottom 10% of early returns)
	worstDecileSize := len(results) / 10
	if worstDecileSize < 1 {
		worstDecileSize = 1
	}
	worstDecade := analyzeDecade(results[:worstDecileSize])

	// Best first decade (top 10% of early returns)
	bestDecade := analyzeDecade(results[len(results)-worstDecileSize:])

	// Calculate correlation between early returns and success
	// Simple: compare success rate in bottom vs top quartile
	quartileSize := len(results) / 4
	if quartileSize < 1 {
		quartileSize = 1
	}

	bottomQuartileSuccess := 0
	for _, r := range results[:quartileSize] {
		if r.success {
			bottomQuartileSuccess++
		}
	}
	topQuartileSuccess := 0
	for _, r := range results[len(results)-quartileSize:] {
		if r.success {
			topQuartileSuccess++
		}
	}

	// Correlation approximation: how much does early return matter?
	// Higher difference = higher sequence impact
	bottomRate := float64(bottomQuartileSuccess) / float64(quartileSize)
	topRate := float64(topQuartileSuccess) / float64(quartileSize)
	earlyReturnCorrelation := (topRate - bottomRate) * 100 // Scale to 0-100

	// Sequence Impact Score (0-100): how much does sequence matter?
	// Based on spread between worst and best decade outcomes
	sequenceImpactScore := 0.0
	if worstDecade != nil && bestDecade != nil && bestDecade.SuccessRate > 0 {
		impactDiff := bestDecade.SuccessRate - worstDecade.SuccessRate
		sequenceImpactScore = impactDiff // Already 0-100
	}

	// Identify vulnerability periods (years where bad returns hurt most)
	vulnerabilityPeriods := identifyVulnerabilityPeriods(params, years)

	return &models.SequenceAnalysis{
		SequenceImpactScore:    sequenceImpactScore,
		VulnerabilityPeriods:   vulnerabilityPeriods,
		WorstFirstDecade:       worstDecade,
		BestFirstDecade:        bestDecade,
		EarlyReturnCorrelation: earlyReturnCorrelation,
	}
}

// analyzeDecade summarizes outcomes for a group of simulations
func analyzeDecade(results []simResult) *models.DecadeAnalysis {
	if len(results) == 0 {
		return nil
	}

	var sumReturns, sumWealth float64
	successCount := 0
	for _, r := range results {
		sumReturns += r.earlyReturn
		sumWealth += r.finalWealth
		if r.success {
			successCount++
		}
	}

	return &models.DecadeAnalysis{
		AvgAnnualReturn: sumReturns / float64(len(results)) * 100, // Convert to percentage
		SuccessRate:     float64(successCount) / float64(len(results)) * 100,
		AvgFinalWealth:  sumWealth / float64(len(results)),
		SampleSize:      len(results),
	}
}

// identifyVulnerabilityPeriods identifies years with outsized impact on success
func identifyVulnerabilityPeriods(params *models.SimulationParams, years int) []models.VulnerabilityPeriod {
	periods := []models.VulnerabilityPeriod{}

	retirementYear := params.RetirementAge - params.CurrentAge
	if retirementYear < 0 {
		retirementYear = 0
	}

	// First 5 years of retirement are most vulnerable (sequence of returns risk)
	if retirementYear < years {
		endYear := retirementYear + 5
		if endYear > years {
			endYear = years
		}
		periods = append(periods, models.VulnerabilityPeriod{
			YearStart:   retirementYear + 1,
			YearEnd:     endYear,
			AgeStart:    params.RetirementAge,
			AgeEnd:      params.CurrentAge + endYear,
			RiskFactor:  2.0, // Bad returns in this period have 2x impact
			Description: "First 5 years of retirement - highest sequence risk",
		})
	}

	// Years 5-10 of retirement are also elevated risk
	if retirementYear+5 < years {
		endYear := retirementYear + 10
		if endYear > years {
			endYear = years
		}
		periods = append(periods, models.VulnerabilityPeriod{
			YearStart:   retirementYear + 6,
			YearEnd:     endYear,
			AgeStart:    params.CurrentAge + retirementYear + 5,
			AgeEnd:      params.CurrentAge + endYear,
			RiskFactor:  1.5,
			Description: "Years 5-10 of retirement - elevated sequence risk",
		})
	}

	return periods
}

// BehavioralState tracks panic sell state across years
type BehavioralState struct {
	InPanic         bool
	MonthsOutOfMarket int
	PanicSellEvents int
	MissedGains     float64
	CashPosition    float64 // Amount held in cash after panic sell
}

// RunMonteCarloWithBehavior runs simulation with behavioral risk applied
func RunMonteCarloWithBehavior(assets []models.Asset, debts []models.Debt, params *models.SimulationParams) (models.MonteCarloResponse, *models.BehavioralImpact) {
	// Run baseline (disciplined investor)
	baselineResponse := RunMonteCarloWithParams(assets, debts, params)

	// If no behavioral params or not enabled, just return baseline
	if params.BehavioralRisk == nil || !params.BehavioralRisk.Enabled {
		return baselineResponse, nil
	}

	// Get behavioral settings
	panicThreshold := params.BehavioralRisk.PanicSellThreshold
	if panicThreshold == 0 {
		panicThreshold = -0.20 // Default: -20% drawdown triggers panic
	}
	panicSellPct := params.BehavioralRisk.PanicSellPct
	if panicSellPct == 0 {
		panicSellPct = 0.50 // Default: sell 50% in panic
	}
	recoveryDelay := params.BehavioralRisk.RecoveryDelay
	if recoveryDelay == 0 {
		recoveryDelay = 6 // Default: 6 months before re-entering
	}

	// Run behavioral simulation
	behavioralResult := runBehavioralSimulation(assets, debts, params, panicThreshold, panicSellPct, recoveryDelay)

	// Calculate impact
	impact := &models.BehavioralImpact{
		BaselineSuccessRate:     baselineResponse.Summary.SuccessRate,
		WithBehaviorSuccessRate: behavioralResult.Summary.SuccessRate,
		SuccessRateDelta:        baselineResponse.Summary.SuccessRate - behavioralResult.Summary.SuccessRate,
		MissedGainsDollar:       baselineResponse.Summary.FinalP50 - behavioralResult.Summary.FinalP50,
		PanicEvents:             calculateAvgPanicEvents(params),
	}

	return behavioralResult, impact
}

// runBehavioralSimulation runs Monte Carlo with behavioral effects
func runBehavioralSimulation(assets []models.Asset, debts []models.Debt, params *models.SimulationParams, panicThreshold, panicSellPct float64, recoveryDelay int) models.MonteCarloResponse {
	params.ApplyDefaults()

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

	results := make([][]float64, NumSimulations)
	contributions := make([][]float64, NumSimulations)
	withdrawals := make([][]float64, NumSimulations)

	for sim := 0; sim < NumSimulations; sim++ {
		results[sim] = make([]float64, years)
		contributions[sim] = make([]float64, years)
		withdrawals[sim] = make([]float64, years)
	}

	successCount := 0
	isAccumulationOnly := retirementYear >= years

	for sim := 0; sim < NumSimulations; sim++ {
		portfolioValue := startingNetWorth
		peakValue := startingNetWorth

		debtValues := make([]float64, len(debts))
		for i, d := range debts {
			debtValues[i] = d.CurrentBalance
		}

		var totalContrib, totalWithdraw float64
		monthlyContrib := params.MonthlyContribution
		monthlySpending := params.RetirementSpending
		ssBenefitAnnual := params.SocialSecurityAmount * 12

		success := true
		retirementStartingValue := 0.0

		// Behavioral state
		behavState := BehavioralState{}

		for year := 0; year < years; year++ {
			age := params.CurrentAge + year
			isRetired := year >= retirementYear

			var yearContribution, yearWithdrawal float64

			if !isRetired {
				annualContrib := monthlyContrib * 12
				employerMatch := calculateEmployerMatch(annualContrib, params.EmployerMatch, params.EmployerMatchLimit)
				totalAnnualContrib := annualContrib + employerMatch
				portfolioValue += totalAnnualContrib
				yearContribution = totalAnnualContrib
				totalContrib += totalAnnualContrib
				monthlyContrib *= (1 + params.ContributionGrowth)
			} else {
				if retirementStartingValue == 0 {
					retirementStartingValue = portfolioValue
				}

				yearWithdrawal = calculateWithdrawal(portfolioValue, monthlySpending*12, params.WithdrawalStrategy, retirementStartingValue)

				ssAge := params.SocialSecurityAge
				if age >= ssAge && params.SocialSecurityAmount > 0 {
					if age > ssAge {
						ssBenefitAnnual *= 1.025
					}
					yearWithdrawal -= ssBenefitAnnual
				}

				if params.PensionIncome > 0 {
					yearWithdrawal -= params.PensionIncome * 12
				}

				if yearWithdrawal < 0 {
					yearWithdrawal = 0
				}

				grossWithdrawal := yearWithdrawal
				if params.RetirementTaxRate > 0 && params.RetirementTaxRate < 1 {
					grossWithdrawal = yearWithdrawal / (1 - params.RetirementTaxRate)
				}

				if grossWithdrawal > portfolioValue {
					success = false
					grossWithdrawal = portfolioValue
				}

				portfolioValue -= grossWithdrawal
				totalWithdraw += grossWithdrawal
				monthlySpending *= (1 + params.InflationRate)
			}

			// Apply one-time events
			for _, event := range params.OneTimeEvents {
				if event.Year == year+1 || (event.Recurring && event.Year <= year+1) {
					portfolioValue += event.Amount
				}
			}

			// Generate return
			var annualReturn float64
			if params.EnableGlidePath {
				glideReturn, glideVolatility := calculateGlidePathParams(age, params.RetirementAge)
				annualReturn = normalRandom(glideReturn, glideVolatility)
			} else {
				annualReturn = normalRandom(params.ExpectedReturn, params.Volatility)
			}

			// Apply behavioral effects
			portfolioValue, behavState = applyBehavioralEffects(
				portfolioValue, peakValue, annualReturn, behavState,
				panicThreshold, panicSellPct, recoveryDelay,
			)

			if portfolioValue > peakValue {
				peakValue = portfolioValue
			}

			var remainingDebt float64
			for _, v := range debtValues {
				remainingDebt += v
			}

			netWorth := portfolioValue - remainingDebt
			results[sim][year] = netWorth
			contributions[sim][year] = yearContribution
			withdrawals[sim][year] = yearWithdrawal
		}

		if isAccumulationOnly && results[sim][years-1] <= 0 {
			success = false
		}

		if success {
			successCount++
		}
	}

	// Calculate percentiles
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

	return models.MonteCarloResponse{
		Projections: projections,
		Summary: models.ProjectionSummary{
			StartingNetWorth:   startingNetWorth,
			FinalP10:           percentile(finalValues, 10),
			FinalP25:           percentile(finalValues, 25),
			FinalP50:           percentile(finalValues, 50),
			FinalP75:           percentile(finalValues, 75),
			FinalP90:           percentile(finalValues, 90),
			Years:              years,
			Simulations:        NumSimulations,
			SuccessRate:        float64(successCount) / float64(NumSimulations) * 100,
			RetirementYear:     retirementYear,
			TotalContributions: totalContribSum / float64(NumSimulations),
			TotalWithdrawals:   totalWithdrawSum / float64(NumSimulations),
		},
	}
}

// applyBehavioralEffects simulates panic selling and recovery delay
func applyBehavioralEffects(portfolioValue, peakValue, annualReturn float64, state BehavioralState, panicThreshold, panicSellPct float64, recoveryDelay int) (float64, BehavioralState) {
	// Calculate current drawdown
	drawdown := 0.0
	if peakValue > 0 {
		drawdown = (portfolioValue - peakValue) / peakValue
	}

	// Check if we should panic sell
	if !state.InPanic && drawdown <= panicThreshold {
		// Panic sell!
		state.InPanic = true
		state.MonthsOutOfMarket = 0
		state.PanicSellEvents++

		// Sell portion of portfolio to cash
		sellAmount := portfolioValue * panicSellPct
		state.CashPosition = sellAmount
		portfolioValue -= sellAmount

		// Apply the negative return to remaining portfolio
		if portfolioValue > 0 {
			portfolioValue *= (1 + annualReturn)
		}
	} else if state.InPanic {
		// In panic mode - money is in cash, missing returns
		state.MonthsOutOfMarket += 12 // Assuming annual steps

		// Track missed gains (what cash would have earned)
		if annualReturn > 0 {
			state.MissedGains += state.CashPosition * annualReturn
		}

		// Apply return to remaining invested portion
		if portfolioValue > 0 {
			portfolioValue *= (1 + annualReturn)
		}

		// Check if recovery period is over
		if state.MonthsOutOfMarket >= recoveryDelay {
			// Re-enter market - add cash back to portfolio
			portfolioValue += state.CashPosition
			state.CashPosition = 0
			state.InPanic = false
		}
	} else {
		// Normal behavior - full market participation
		if portfolioValue > 0 {
			portfolioValue *= (1 + annualReturn)
		}
	}

	if portfolioValue < 0 {
		portfolioValue = 0
	}

	return portfolioValue, state
}

// calculateAvgPanicEvents estimates average panic events based on volatility
func calculateAvgPanicEvents(params *models.SimulationParams) float64 {
	// Higher volatility = more drawdowns = more panic events
	// This is an approximation based on historical data
	yearsOfInvesting := float64(params.TimeHorizonYears)
	volatility := params.Volatility

	// Roughly, with 15% volatility, expect ~0.3 panic events per year on average
	// (20%+ drawdowns happen about once every 3-4 years)
	panicRate := (volatility / 0.15) * 0.3
	return yearsOfInvesting * panicRate
}
