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
	Params *SimulationParams `json:"params"`
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
	SuccessRate          float64 `json:"successRate"`                      // % of simulations that don't run out of money
	RetirementYear       int     `json:"retirementYear"`                   // year retirement begins
	TotalContributions   float64 `json:"totalContributions"`               // sum of all contributions
	TotalWithdrawals     float64 `json:"totalWithdrawals"`                 // sum of all withdrawals
	AccumulationWarnings int     `json:"accumulationWarnings,omitempty"`   // simulations with pre-retirement negative net worth
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
