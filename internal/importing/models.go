package importing

type Node struct {
	Name   string
	States []string
}

type Edge struct {
	From string
	To   []string
}

type Data struct {
	States []string
	Fields map[string]string
	Values []float64
}

type Coefficient struct {
	Node string
	Data []*Data
}

type Import struct {
	Nodes        []*Node
	Edges        []*Edge
	Coefficients []*Coefficient
}
