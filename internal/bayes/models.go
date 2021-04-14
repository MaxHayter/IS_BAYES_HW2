package bayes

import "github.com/MaxHayter/IS_BAYES_HW2/internal/graph"

type Bayes struct {
	Graph *graph.Graph
}

type Row struct {
	States      map[string]string
	Probability float64
}
