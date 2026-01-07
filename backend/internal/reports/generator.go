package reports

import (
	"fmt"
	"time"

	"github.com/finviz/backend/internal/models"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// ReportData contains all information needed for a financial plan report
type ReportData struct {
	ClientName   string
	AdvisorName  string
	GeneratedAt  time.Time
	Assets       []models.Asset
	Debts        []models.Debt
	Simulation   *models.MonteCarloResponse
	Params       *models.SimulationParams
	TotalAssets  float64
	TotalDebts   float64
	NetWorth     float64
}

// GenerateFinancialPlanReport creates a PDF report for a financial plan
func GenerateFinancialPlanReport(data ReportData) ([]byte, error) {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(15).
		WithTopMargin(15).
		WithRightMargin(15).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	// Header
	addHeader(m, data)

	// Executive Summary
	addExecutiveSummary(m, data)

	// Net Worth Summary
	addNetWorthSection(m, data)

	// Projection Summary
	if data.Simulation != nil {
		addProjectionSection(m, data)
	}

	// Asset Details
	if len(data.Assets) > 0 {
		addAssetTable(m, data.Assets)
	}

	// Debt Details
	if len(data.Debts) > 0 {
		addDebtTable(m, data.Debts)
	}

	// Milestones
	if data.Simulation != nil && len(data.Simulation.Milestones) > 0 {
		addMilestonesSection(m, data.Simulation.Milestones)
	}

	// Insights/Recommendations
	if data.Simulation != nil && len(data.Simulation.Insights) > 0 {
		addInsightsSection(m, data.Simulation.Insights)
	}

	// Disclaimer
	addDisclaimer(m)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return doc.GetBytes(), nil
}

func addHeader(m core.Maroto, data ReportData) {
	m.AddRow(20,
		col.New(12).Add(
			text.New("Financial Plan Report", props.Text{
				Size:  24,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	m.AddRow(8,
		col.New(6).Add(
			text.New(fmt.Sprintf("Prepared for: %s", data.ClientName), props.Text{
				Size:  12,
				Style: fontstyle.Bold,
			}),
		),
		col.New(6).Add(
			text.New(fmt.Sprintf("Date: %s", data.GeneratedAt.Format("January 2, 2006")), props.Text{
				Size:  12,
				Align: align.Right,
			}),
		),
	)

	if data.AdvisorName != "" {
		m.AddRow(6,
			col.New(12).Add(
				text.New(fmt.Sprintf("Prepared by: %s", data.AdvisorName), props.Text{
					Size: 10,
				}),
			),
		)
	}

	m.AddRow(5, line.NewCol(12))
}

func addExecutiveSummary(m core.Maroto, data ReportData) {
	m.AddRow(12,
		col.New(12).Add(
			text.New("Executive Summary", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	summary := fmt.Sprintf(
		"This report summarizes the financial position and retirement projection for %s. "+
			"Current net worth is %s, composed of %s in assets and %s in debts.",
		data.ClientName,
		formatCurrency(data.NetWorth),
		formatCurrency(data.TotalAssets),
		formatCurrency(data.TotalDebts),
	)

	if data.Simulation != nil {
		summary += fmt.Sprintf(
			" Based on a %d-year Monte Carlo simulation with %d scenarios, "+
				"there is a %.1f%% probability of achieving financial goals.",
			data.Simulation.Summary.Years,
			data.Simulation.Summary.Simulations,
			data.Simulation.Summary.SuccessRate,
		)
	}

	m.AddRow(20,
		col.New(12).Add(
			text.New(summary, props.Text{
				Size: 10,
			}),
		),
	)

	m.AddRow(3)
}

func addNetWorthSection(m core.Maroto, data ReportData) {
	m.AddRow(12,
		col.New(12).Add(
			text.New("Net Worth Summary", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	// Net worth metrics in a row
	m.AddRow(15,
		col.New(4).Add(
			text.New("Total Assets", props.Text{
				Size:  10,
				Align: align.Center,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
		col.New(4).Add(
			text.New("Total Debts", props.Text{
				Size:  10,
				Align: align.Center,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
		col.New(4).Add(
			text.New("Net Worth", props.Text{
				Size:  10,
				Align: align.Center,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
	)

	assetColor := &props.Color{Red: 0, Green: 150, Blue: 100}
	debtColor := &props.Color{Red: 200, Green: 50, Blue: 50}
	netColor := assetColor
	if data.NetWorth < 0 {
		netColor = debtColor
	}

	m.AddRow(12,
		col.New(4).Add(
			text.New(formatCurrency(data.TotalAssets), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: assetColor,
			}),
		),
		col.New(4).Add(
			text.New(formatCurrency(data.TotalDebts), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: debtColor,
			}),
		),
		col.New(4).Add(
			text.New(formatCurrency(data.NetWorth), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: netColor,
			}),
		),
	)

	m.AddRow(5)
}

func addProjectionSection(m core.Maroto, data ReportData) {
	m.AddRow(12,
		col.New(12).Add(
			text.New("Retirement Projection", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	summary := data.Simulation.Summary

	// Success rate highlight
	successColor := &props.Color{Red: 0, Green: 150, Blue: 100}
	if summary.SuccessRate < 80 {
		successColor = &props.Color{Red: 200, Green: 150, Blue: 0}
	}
	if summary.SuccessRate < 50 {
		successColor = &props.Color{Red: 200, Green: 50, Blue: 50}
	}

	m.AddRow(8,
		col.New(12).Add(
			text.New(fmt.Sprintf("Success Rate: %.1f%%", summary.SuccessRate), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: successColor,
			}),
		),
	)

	m.AddRow(6,
		col.New(12).Add(
			text.New(fmt.Sprintf("Based on %d Monte Carlo simulations over %d years",
				summary.Simulations, summary.Years), props.Text{
				Size:  9,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
	)

	// Projection outcomes table
	m.AddRow(10,
		col.New(4).Add(
			text.New("Scenario", props.Text{Size: 10, Style: fontstyle.Bold}),
		),
		col.New(4).Add(
			text.New("Final Net Worth", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right}),
		),
		col.New(4).Add(
			text.New("Description", props.Text{Size: 10, Style: fontstyle.Bold}),
		),
	)

	scenarios := []struct {
		name  string
		value float64
		desc  string
	}{
		{"Conservative (P10)", summary.FinalP10, "Worst 10% of outcomes"},
		{"Median (P50)", summary.FinalP50, "Typical outcome"},
		{"Optimistic (P90)", summary.FinalP90, "Best 10% of outcomes"},
	}

	for _, s := range scenarios {
		m.AddRow(8,
			col.New(4).Add(
				text.New(s.name, props.Text{Size: 9}),
			),
			col.New(4).Add(
				text.New(formatCurrency(s.value), props.Text{Size: 9, Align: align.Right}),
			),
			col.New(4).Add(
				text.New(s.desc, props.Text{Size: 9, Color: &props.Color{Red: 100, Green: 100, Blue: 100}}),
			),
		)
	}

	// Parameters used
	if data.Params != nil {
		m.AddRow(10,
			col.New(12).Add(
				text.New("Simulation Parameters", props.Text{
					Size:  12,
					Style: fontstyle.Bold,
				}),
			),
		)

		params := []struct {
			label string
			value string
		}{
			{"Current Age", fmt.Sprintf("%d", data.Params.CurrentAge)},
			{"Retirement Age", fmt.Sprintf("%d", data.Params.RetirementAge)},
			{"Time Horizon", fmt.Sprintf("%d years", data.Params.TimeHorizonYears)},
			{"Monthly Contribution", formatCurrency(data.Params.MonthlyContribution)},
			{"Expected Return", fmt.Sprintf("%.1f%%", data.Params.ExpectedReturn*100)},
			{"Volatility", fmt.Sprintf("%.1f%%", data.Params.Volatility*100)},
			{"Inflation Rate", fmt.Sprintf("%.1f%%", data.Params.InflationRate*100)},
		}

		if data.Params.RetirementSpending > 0 {
			params = append(params, struct{ label, value string }{
				"Monthly Retirement Spending", formatCurrency(data.Params.RetirementSpending),
			})
		}

		if data.Params.SocialSecurityAmount > 0 {
			params = append(params, struct{ label, value string }{
				"Social Security", fmt.Sprintf("%s/mo starting at age %d",
					formatCurrency(data.Params.SocialSecurityAmount), data.Params.SocialSecurityAge),
			})
		}

		for i := 0; i < len(params); i += 2 {
			if i+1 < len(params) {
				m.AddRow(6,
					col.New(3).Add(text.New(params[i].label+":", props.Text{Size: 9})),
					col.New(3).Add(text.New(params[i].value, props.Text{Size: 9, Style: fontstyle.Bold})),
					col.New(3).Add(text.New(params[i+1].label+":", props.Text{Size: 9})),
					col.New(3).Add(text.New(params[i+1].value, props.Text{Size: 9, Style: fontstyle.Bold})),
				)
			} else {
				m.AddRow(6,
					col.New(3).Add(text.New(params[i].label+":", props.Text{Size: 9})),
					col.New(9).Add(text.New(params[i].value, props.Text{Size: 9, Style: fontstyle.Bold})),
				)
			}
		}
	}

	m.AddRow(5)
}

func addAssetTable(m core.Maroto, assets []models.Asset) {
	m.AddRow(12,
		col.New(12).Add(
			text.New("Asset Details", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	// Header row
	m.AddRow(8,
		col.New(5).Add(text.New("Asset Name", props.Text{Size: 10, Style: fontstyle.Bold})),
		col.New(3).Add(text.New("Type", props.Text{Size: 10, Style: fontstyle.Bold})),
		col.New(4).Add(text.New("Value", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
	)

	for _, asset := range assets {
		typeName := "Unknown"
		if asset.AssetType != nil {
			typeName = asset.AssetType.Name
		}
		m.AddRow(6,
			col.New(5).Add(text.New(asset.Name, props.Text{Size: 9})),
			col.New(3).Add(text.New(typeName, props.Text{Size: 9})),
			col.New(4).Add(text.New(formatCurrency(asset.CurrentValue), props.Text{Size: 9, Align: align.Right})),
		)
	}

	m.AddRow(5)
}

func addDebtTable(m core.Maroto, debts []models.Debt) {
	m.AddRow(12,
		col.New(12).Add(
			text.New("Debt Details", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	// Header row
	m.AddRow(8,
		col.New(4).Add(text.New("Debt Name", props.Text{Size: 10, Style: fontstyle.Bold})),
		col.New(3).Add(text.New("Balance", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
		col.New(2).Add(text.New("Rate", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
		col.New(3).Add(text.New("Min Payment", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
	)

	for _, debt := range debts {
		rate := "N/A"
		if debt.InterestRate != nil {
			rate = fmt.Sprintf("%.2f%%", *debt.InterestRate)
		}
		payment := "N/A"
		if debt.MinimumPayment != nil {
			payment = formatCurrency(*debt.MinimumPayment)
		}

		m.AddRow(6,
			col.New(4).Add(text.New(debt.Name, props.Text{Size: 9})),
			col.New(3).Add(text.New(formatCurrency(debt.CurrentBalance), props.Text{Size: 9, Align: align.Right})),
			col.New(2).Add(text.New(rate, props.Text{Size: 9, Align: align.Right})),
			col.New(3).Add(text.New(payment, props.Text{Size: 9, Align: align.Right})),
		)
	}

	m.AddRow(5)
}

func addMilestonesSection(m core.Maroto, milestones []models.Milestone) {
	m.AddRow(12,
		col.New(12).Add(
			text.New("Financial Milestones", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	m.AddRow(8,
		col.New(5).Add(text.New("Milestone", props.Text{Size: 10, Style: fontstyle.Bold})),
		col.New(3).Add(text.New("Median Year", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Center})),
		col.New(4).Add(text.New("Probability", props.Text{Size: 10, Style: fontstyle.Bold, Align: align.Right})),
	)

	for _, ms := range milestones {
		m.AddRow(6,
			col.New(5).Add(text.New(ms.Description, props.Text{Size: 9})),
			col.New(3).Add(text.New(fmt.Sprintf("Year %d", ms.MedianYear), props.Text{Size: 9, Align: align.Center})),
			col.New(4).Add(text.New(fmt.Sprintf("%.1f%%", ms.ProbabilityPct), props.Text{Size: 9, Align: align.Right})),
		)
	}

	m.AddRow(5)
}

func addInsightsSection(m core.Maroto, insights []models.Insight) {
	m.AddRow(12,
		col.New(12).Add(
			text.New("Recommendations", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 0, Green: 82, Blue: 147},
			}),
		),
	)

	for _, insight := range insights {
		iconColor := &props.Color{Red: 0, Green: 100, Blue: 200}
		switch insight.Type {
		case "warning":
			iconColor = &props.Color{Red: 200, Green: 150, Blue: 0}
		case "success":
			iconColor = &props.Color{Red: 0, Green: 150, Blue: 100}
		case "opportunity":
			iconColor = &props.Color{Red: 100, Green: 100, Blue: 200}
		}

		m.AddRow(8,
			col.New(12).Add(
				text.New(fmt.Sprintf("• %s", insight.Title), props.Text{
					Size:  10,
					Style: fontstyle.Bold,
					Color: iconColor,
				}),
			),
		)
		m.AddRow(8,
			col.New(12).Add(
				text.New(fmt.Sprintf("  %s", insight.Message), props.Text{
					Size: 9,
				}),
			),
		)
	}

	m.AddRow(5)
}

func addDisclaimer(m core.Maroto) {
	m.AddRow(3, line.NewCol(12))

	m.AddRow(20,
		col.New(12).Add(
			text.New("IMPORTANT DISCLOSURE: This report is for informational purposes only and does not "+
				"constitute financial, investment, tax, or legal advice. Past performance does not guarantee "+
				"future results. Monte Carlo simulations are based on historical data and stated assumptions; "+
				"actual outcomes will vary. Projections assume reinvestment of returns and do not account for "+
				"taxes unless explicitly stated. Please consult with a qualified financial advisor, tax "+
				"professional, or attorney before making any significant financial decisions.", props.Text{
				Size:  8,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
	)
}

func formatCurrency(amount float64) string {
	if amount >= 1000000 {
		return fmt.Sprintf("$%.2fM", amount/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("$%.0fK", amount/1000)
	}
	if amount < 0 {
		return fmt.Sprintf("-$%.2f", -amount)
	}
	return fmt.Sprintf("$%.2f", amount)
}
