package models

// SimulationParams contains all inputs for Monte Carlo simulation
type SimulationParams struct {
	// Tier 1 - Essential (always visible)
	TimeHorizonYears    int     `json:"timeHorizonYears"`
	MonthlyContribution float64 `json:"monthlyContribution"`
	RetirementAge       int     `json:"retirementAge"`
	CurrentAge          int     `json:"currentAge"`

	// Tier 2 - Important (collapsed section)
	ExpectedReturn       float64 `json:"expectedReturn"`       // default 0.07 (7%)
	InflationRate        float64 `json:"inflationRate"`        // default 0.03 (3%)
	ContributionGrowth   float64 `json:"contributionGrowth"`   // default 0.02 (2% annual raise)
	RetirementSpending   float64 `json:"retirementSpending"`   // monthly spending in retirement
	SocialSecurityAmount float64 `json:"socialSecurityAmount"` // monthly SS benefit
	SocialSecurityAge    int     `json:"socialSecurityAge"`    // age SS begins (default 67)

	// Tier 3 - Advanced (hidden by default)
	EmployerMatch         float64 `json:"employerMatch"`         // match percentage (e.g., 0.50 = 50%)
	EmployerMatchLimit    float64 `json:"employerMatchLimit"`    // annual cap on employer match
	Volatility            float64 `json:"volatility"`            // default 0.15 (15%)
	PensionIncome         float64 `json:"pensionIncome"`         // monthly pension
	OneTimeEvents         []Event `json:"oneTimeEvents"`
	WithdrawalStrategy    string  `json:"withdrawalStrategy"`    // "fixed", "dynamic", "guardrails"
	RetirementTaxRate     float64 `json:"retirementTaxRate"`     // effective tax rate in retirement
	RunHistoricalTest     bool    `json:"runHistoricalTest"`     // run against historical sequences
	ExcludeCreditCardDebt bool    `json:"excludeCreditCardDebt"` // exclude revolving credit from projections
	EnableGlidePath       bool    `json:"enableGlidePath"`       // auto-adjust risk by age (target-date style)

	// Tier 4 - Behavioral Risk (experimental)
	BehavioralRisk *BehavioralParams `json:"behavioralRisk,omitempty"` // Behavioral risk modeling parameters
}

// Event represents a one-time or recurring financial event
type Event struct {
	Year        int     `json:"year"`        // year relative to start (1, 2, 3...)
	Amount      float64 `json:"amount"`      // positive = income, negative = expense
	Description string  `json:"description"` // e.g., "Home purchase", "Inheritance"
	Recurring   bool    `json:"recurring"`   // if true, repeats every year after
}

// MonteCarloRequest is the API request for running a simulation
type MonteCarloRequest struct {
	Params     *SimulationParams `json:"params"`
	SaveResult bool              `json:"saveResult,omitempty"` // Whether to save the result to history
	Name       *string           `json:"name,omitempty"`       // Optional name for saved simulation
	Notes      *string           `json:"notes,omitempty"`      // Optional notes for saved simulation
}

// YearProjection contains projection data for a single year
type YearProjection struct {
	Year          int     `json:"year"`
	Age           int     `json:"age,omitempty"`
	P10           float64 `json:"p10"`           // 10th percentile (conservative)
	P25           float64 `json:"p25,omitempty"` // 25th percentile
	P50           float64 `json:"p50"`           // 50th percentile (median)
	P75           float64 `json:"p75,omitempty"` // 75th percentile
	P90           float64 `json:"p90"`           // 90th percentile (optimistic)
	Phase         string  `json:"phase"`         // "accumulation" or "distribution"
	Contributions float64 `json:"contributions"` // total contributed this year
	Withdrawals   float64 `json:"withdrawals"`   // total withdrawn this year
}

// Milestone represents a financial goal and probability of achieving it
type Milestone struct {
	Description    string  `json:"description"`    // e.g., "Reach $1M net worth"
	TargetAmount   float64 `json:"targetAmount"`   // target value
	MedianYear     int     `json:"medianYear"`     // P50 year achieved (0 if not reached)
	ProbabilityPct float64 `json:"probabilityPct"` // % of simulations that reach it
}

// Insight represents an actionable recommendation
type Insight struct {
	Type    string `json:"type"`    // "warning", "opportunity", "info", "success"
	Title   string `json:"title"`   // short title
	Message string `json:"message"` // detailed explanation
}

// MonteCarloResponse is the API response for a simulation
type MonteCarloResponse struct {
	Projections []YearProjection  `json:"projections"`
	Summary     ProjectionSummary `json:"summary"`
	Milestones  []Milestone       `json:"milestones,omitempty"`
	Insights    []Insight         `json:"insights,omitempty"`
}

// ProjectionSummary contains overall simulation results
type ProjectionSummary struct {
	StartingNetWorth     float64 `json:"startingNetWorth"`
	FinalP10             float64 `json:"finalP10"`
	FinalP25             float64 `json:"finalP25,omitempty"`
	FinalP50             float64 `json:"finalP50"`
	FinalP75             float64 `json:"finalP75,omitempty"`
	FinalP90             float64 `json:"finalP90"`
	Years                int     `json:"years"`
	Simulations          int     `json:"simulations"`
	SuccessRate          float64 `json:"successRate"`                    // % of simulations that don't run out of money
	RetirementYear       int     `json:"retirementYear"`                 // year retirement begins
	TotalContributions   float64 `json:"totalContributions"`             // sum of all contributions
	TotalWithdrawals     float64 `json:"totalWithdrawals"`               // sum of all withdrawals
	AccumulationWarnings int     `json:"accumulationWarnings,omitempty"` // simulations with pre-retirement negative net worth

	// Enhanced Success Metrics (Priority 3)
	EnhancedMetrics *EnhancedMetrics `json:"enhancedMetrics,omitempty"`
}

// EnhancedMetrics provides richer success analysis beyond simple success rate
type EnhancedMetrics struct {
	MedianWealthAtEnd   float64            `json:"medianWealthAtEnd"`   // P50 final wealth
	RuinProbabilities   []RuinProbability  `json:"ruinProbabilities"`   // Probability of ruin at various ages
	SafeFloor           SafeFloor          `json:"safeFloor"`           // Guaranteed minimum amount
	RecoveryMetrics     RecoveryAnalysis   `json:"recoveryMetrics"`     // Drawdown recovery stats
	PartialSuccessRate  float64            `json:"partialSuccessRate"`  // % that made it >50% of retirement years
	MedianYearsToRuin   float64            `json:"medianYearsToRuin"`   // Median years until failure (for failed sims)
	WealthAtRuin        float64            `json:"wealthAtRuin"`        // Median shortfall when failing
	SequenceAnalysis    *SequenceAnalysis  `json:"sequenceAnalysis,omitempty"`
}

// RuinProbability represents the probability of running out of money by a specific age
type RuinProbability struct {
	Age         int     `json:"age"`
	Probability float64 `json:"probability"` // 0-100 percentage
	YearsOut    int     `json:"yearsOut"`    // Years from simulation start
}

// SafeFloor represents the minimum guaranteed wealth in worst-case scenarios
type SafeFloor struct {
	GuaranteedMinimum float64 `json:"guaranteedMinimum"` // 5th percentile at worst point
	FloorYear         int     `json:"floorYear"`         // Year when floor occurs
	FloorAge          int     `json:"floorAge"`          // Age when floor occurs
	Description       string  `json:"description"`       // Human-readable explanation
}

// RecoveryAnalysis tracks how portfolios recover from drawdowns
type RecoveryAnalysis struct {
	AvgRecoveryYears   float64 `json:"avgRecoveryYears"`   // Average years to recover from 20%+ drawdown
	WorstDrawdown      float64 `json:"worstDrawdown"`      // Largest peak-to-trough decline (percentage)
	DrawdownCount      int     `json:"drawdownCount"`      // Average number of 20%+ drawdowns
	RecoverySuccessRate float64 `json:"recoverySuccessRate"` // % of 20%+ drawdowns that fully recovered
}

// SequenceAnalysis tracks how return sequences affect outcomes
type SequenceAnalysis struct {
	SequenceImpactScore  float64              `json:"sequenceImpactScore"`  // 0-100, how much sequence matters
	VulnerabilityPeriods []VulnerabilityPeriod `json:"vulnerabilityPeriods"` // Periods where bad returns hurt most
	WorstFirstDecade     *DecadeAnalysis      `json:"worstFirstDecade"`     // Analysis of worst first 10 years
	BestFirstDecade      *DecadeAnalysis      `json:"bestFirstDecade"`      // Analysis of best first 10 years
	EarlyReturnCorrelation float64            `json:"earlyReturnCorrelation"` // Correlation between early returns and success
}

// VulnerabilityPeriod identifies years where returns have outsized impact
type VulnerabilityPeriod struct {
	YearStart   int     `json:"yearStart"`
	YearEnd     int     `json:"yearEnd"`
	AgeStart    int     `json:"ageStart"`
	AgeEnd      int     `json:"ageEnd"`
	RiskFactor  float64 `json:"riskFactor"`  // Multiplier: 2.0 means 2x impact
	Description string  `json:"description"` // e.g., "First 5 years of retirement"
}

// DecadeAnalysis summarizes outcomes based on first decade returns
type DecadeAnalysis struct {
	AvgAnnualReturn float64 `json:"avgAnnualReturn"` // Average return in that decade
	SuccessRate     float64 `json:"successRate"`     // Success rate for sims with this decade
	AvgFinalWealth  float64 `json:"avgFinalWealth"`  // Average ending wealth
	SampleSize      int     `json:"sampleSize"`      // Number of simulations in this group
}

// SimulationRun tracks a single simulation's details for sequence analysis
type SimulationRun struct {
	Returns      []float64 `json:"returns,omitempty"` // Annual returns for this run
	FinalWealth  float64   `json:"finalWealth"`
	FailureYear  int       `json:"failureYear"`  // -1 if successful
	EarlyReturns float64   `json:"earlyReturns"` // Avg return years 1-10
	Success      bool      `json:"success"`
}

// BehavioralImpact tracks the cost of investor behavior
type BehavioralImpact struct {
	BaselineSuccessRate     float64 `json:"baselineSuccessRate"`
	WithBehaviorSuccessRate float64 `json:"withBehaviorSuccessRate"`
	SuccessRateDelta        float64 `json:"successRateDelta"`
	MissedGainsDollar       float64 `json:"missedGainsDollar"` // Dollar cost of behavior
	PanicEvents             float64 `json:"panicEvents"`       // Avg panic sells per simulation
}

// ScenarioComparisonRequest is the API request for comparing multiple scenarios
type ScenarioComparisonRequest struct {
	Scenarios []Scenario `json:"scenarios"`
}

// Scenario represents a named simulation scenario
type Scenario struct {
	Name   string            `json:"name"`
	Params *SimulationParams `json:"params"`
}

// ScenarioComparisonResponse is the API response for scenario comparison
type ScenarioComparisonResponse struct {
	Scenarios   []ScenarioResult `json:"scenarios"`
	Comparisons []ScenarioDiff   `json:"comparisons"`
	BestScenario string          `json:"bestScenario"` // Name of scenario with highest success rate
}

// ScenarioResult contains the results for a single scenario
type ScenarioResult struct {
	Name        string            `json:"name"`
	Summary     ProjectionSummary `json:"summary"`
	Projections []YearProjection  `json:"projections"`
}

// ScenarioDiff compares two scenarios
type ScenarioDiff struct {
	ScenarioA          string  `json:"scenarioA"`
	ScenarioB          string  `json:"scenarioB"`
	SuccessRateDiff    float64 `json:"successRateDiff"`    // A - B
	FinalP50Diff       float64 `json:"finalP50Diff"`       // A - B
	ContributionsDiff  float64 `json:"contributionsDiff"`  // A - B
	Recommendation     string  `json:"recommendation"`     // Human-readable recommendation
}

// BehavioralParams configures behavioral risk modeling
type BehavioralParams struct {
	Enabled           bool    `json:"enabled"`           // Whether to apply behavioral modeling
	Model             string  `json:"model"`             // "none", "moderate", "severe"
	PanicSellThreshold float64 `json:"panicSellThreshold"` // Drawdown % triggering panic (e.g., -0.20)
	PanicSellPct      float64 `json:"panicSellPct"`      // % of portfolio sold in panic (e.g., 0.50)
	RecoveryDelay     int     `json:"recoveryDelay"`     // Months before re-entering market
}

// DefaultSimulationParams returns params with sensible defaults
func DefaultSimulationParams() SimulationParams {
	return SimulationParams{
		TimeHorizonYears:     30,
		MonthlyContribution:  0,
		RetirementAge:        65,
		CurrentAge:           35,
		ExpectedReturn:       0.07,
		InflationRate:        0.03,
		ContributionGrowth:   0.02,
		RetirementSpending:   0,
		SocialSecurityAmount: 0,
		SocialSecurityAge:    67,
		EmployerMatch:        0,
		EmployerMatchLimit:   0,
		Volatility:           0.15,
		PensionIncome:        0,
		OneTimeEvents:        []Event{},
		WithdrawalStrategy:   "fixed",
		RetirementTaxRate:    0.22,
		RunHistoricalTest:    false,
		EnableGlidePath:      false,
	}
}

// ApplyDefaults fills in zero values with defaults
func (p *SimulationParams) ApplyDefaults() {
	defaults := DefaultSimulationParams()

	if p.TimeHorizonYears == 0 {
		p.TimeHorizonYears = defaults.TimeHorizonYears
	}
	if p.RetirementAge == 0 {
		p.RetirementAge = defaults.RetirementAge
	}
	if p.CurrentAge == 0 {
		p.CurrentAge = defaults.CurrentAge
	}
	if p.ExpectedReturn == 0 {
		p.ExpectedReturn = defaults.ExpectedReturn
	}
	if p.InflationRate == 0 {
		p.InflationRate = defaults.InflationRate
	}
	if p.ContributionGrowth == 0 {
		p.ContributionGrowth = defaults.ContributionGrowth
	}
	if p.SocialSecurityAge == 0 {
		p.SocialSecurityAge = defaults.SocialSecurityAge
	}
	if p.Volatility == 0 {
		p.Volatility = defaults.Volatility
	}
	if p.WithdrawalStrategy == "" {
		p.WithdrawalStrategy = defaults.WithdrawalStrategy
	}
	if p.RetirementTaxRate == 0 {
		p.RetirementTaxRate = defaults.RetirementTaxRate
	}
	if p.OneTimeEvents == nil {
		p.OneTimeEvents = []Event{}
	}
}
