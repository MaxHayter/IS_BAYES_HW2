package graph

type Node struct {
	Name                 string
	States               []string
	Coefficients         []float64
	InternalCoefficients map[string][]*Coefficient
	Children             []*Node
	Parents              []*Node
	constant             bool
}

type Coefficient struct {
	States map[string]string
	Value  float64
}

type Graph struct {
	Nodes map[string]*Node
}
