package models

type MonteCarloRequest struct {
	Years int `json:"years"`
}

type YearProjection struct {
	Year int     `json:"year"`
	P10  float64 `json:"p10"`  // 10th percentile (bad case)
	P50  float64 `json:"p50"`  // 50th percentile (median)
	P90  float64 `json:"p90"`  // 90th percentile (good case)
}

type MonteCarloResponse struct {
	Projections []YearProjection `json:"projections"`
	Summary     ProjectionSummary `json:"summary"`
}

type ProjectionSummary struct {
	StartingNetWorth float64 `json:"startingNetWorth"`
	FinalP10         float64 `json:"finalP10"`
	FinalP50         float64 `json:"finalP50"`
	FinalP90         float64 `json:"finalP90"`
	Years            int     `json:"years"`
	Simulations      int     `json:"simulations"`
}
