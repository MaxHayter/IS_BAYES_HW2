package graph

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const numSpaces = 4

func NewGraph() *Graph {
	return &Graph{Nodes: make(map[string]*Node)}
}

func (g *Graph) CreateNode(name string) error {
	if _, ok := g.Nodes[name]; ok {
		return errors.New("this name already exists")
	}

	g.Nodes[name] = &Node{Name: name}
	return nil
}

func (g *Graph) SetStates(name string, states []string) error {
	if _, ok := g.Nodes[name]; !ok {
		return errors.New("node with this name does not exist")
	}

	g.Nodes[name].States = states
	g.Nodes[name].Coefficients = make([]float64, len(states))
	g.Nodes[name].InternalCoefficients = make(map[string][]*Coefficient, len(states))

	return g.updateNode(name)
}

func (g *Graph) updateNode(name string) error {
	if _, ok := g.Nodes[name]; !ok {
		return errors.New("node with this name does not exist")
	}

	states := g.Nodes[name].States

	numVariant := 1
	for i := range g.Nodes[name].Parents {
		numVariant *= len(g.Nodes[name].Parents[i].States)
	}

	for i := range states {
		g.Nodes[name].InternalCoefficients[states[i]] = make([]*Coefficient, numVariant)
		for j := 0; j < numVariant; j++ {
			g.Nodes[name].InternalCoefficients[states[i]][j] = &Coefficient{States: make(map[string]string, len(g.Nodes[name].Parents))}
		}
	}

	if g.Nodes[name].Parents != nil {
		k := 1
		for i := range g.Nodes[name].Parents {
			lenStates := len(g.Nodes[name].Parents[i].States)
			for j := 0; j < k; j++ {
				for l := range g.Nodes[name].Parents[i].States {
					for m := (j * numVariant / k) + l*numVariant/(k*lenStates); m < (j*numVariant/k)+(l+1)*numVariant/(k*lenStates); m++ {
						for n := range states {
							g.Nodes[name].InternalCoefficients[states[n]][m].States[g.Nodes[name].Parents[i].Name] = g.Nodes[name].Parents[i].States[l]
							g.Nodes[name].InternalCoefficients[states[n]][m].Value = 1.0 / float64(len(states))
						}
					}
				}
			}
			k *= lenStates
		}
	} else {
		for i := range states {
			g.Nodes[name].InternalCoefficients[states[i]][0].Value = 1.0 / float64(len(states))
		}
	}

	return g.calculateCoefficients(name)
}

func (g *Graph) AddEdges(from string, to ...string) error {
	if len(to) < 1 {
		return errors.New("incorrect input data")
	}

	for i := range to {
		err := g.addEdge(from, to[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) addEdge(from, to string) error {
	if _, ok := g.Nodes[from]; !ok {
		return errors.New("node (from) with this name does not exist")
	}

	if _, ok := g.Nodes[to]; !ok {
		return errors.New("node (to) with this name does not exist")
	}

	if isContainNodes(g.Nodes[from].Children, to) || isContainNodes(g.Nodes[to].Parents, from) {
		return errors.New("edge between nodes already exists")
	}

	g.Nodes[from].Children = append(g.Nodes[from].Children, g.Nodes[to])
	g.Nodes[to].Parents = append(g.Nodes[to].Parents, g.Nodes[from])

	return g.updateNode(to)
}

func isContainNodes(nodes []*Node, name string) bool {
	for i := range nodes {
		if nodes[i].Name == name {
			return true
		}
	}
	return false
}

func isContainString(list []string, str string) bool {
	for i := range list {
		if list[i] == str {
			return true
		}
	}
	return false
}

func compareMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if _, ok := b[i]; !ok || a[i] != b[i] {
			return false
		}
	}

	return true
}

func (g *Graph) SetInternalCoefficients(name string, states []string, values []float64, InternalCoefficientStates map[string]string) error {
	if _, ok := g.Nodes[name]; !ok {
		return errors.New("node with this name does not exist")
	}

	if len(states) != len(values) {
		return errors.New("incorrect input data")
	}

	var sum float64
	for i := range values {
		sum += values[i]
	}
	if math.Round(sum) != 1.0 {
		return errors.New("incorrect input values")
	}

	for i := range states {
		if !isContainString(g.Nodes[name].States, states[i]) {
			return errors.New("node with this name does not exist")
		}
	}

	index := -1
	for i := range g.Nodes[name].InternalCoefficients[states[0]] {
		if compareMaps(InternalCoefficientStates, g.Nodes[name].InternalCoefficients[states[0]][i].States) {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("bad states")
	}

	for i := range states {
		g.Nodes[name].InternalCoefficients[states[i]][index].Value = values[i]
	}

	return g.calculateCoefficients(name)
}

func (g *Graph) getCoefficient(name, state string) (float64, error) {
	if _, ok := g.Nodes[name]; !ok {
		return 0, errors.New("node with this name does not exist")
	}

	for i := range g.Nodes[name].States {
		if g.Nodes[name].States[i] == state {
			return g.Nodes[name].Coefficients[i], nil
		}
	}

	return 0, errors.New("incorrect state")
}

func (g *Graph) calculateCoefficients(name string) error {
	if _, ok := g.Nodes[name]; !ok {
		return errors.New("node with this name does not exist")
	}

	if g.Nodes[name].Parents == nil {
		for i := range g.Nodes[name].States {
			g.Nodes[name].Coefficients[i] = g.Nodes[name].InternalCoefficients[g.Nodes[name].States[i]][0].Value
		}
	} else {
		for i := range g.Nodes[name].States {
			var c float64
			for j := range g.Nodes[name].InternalCoefficients[g.Nodes[name].States[i]] {
				cState := g.Nodes[name].InternalCoefficients[g.Nodes[name].States[i]][j].Value
				for pNodeName, state := range g.Nodes[name].InternalCoefficients[g.Nodes[name].States[i]][j].States {
					if coef, err := g.getCoefficient(pNodeName, state); err == nil {
						cState *= coef

					}
				}
				c += cState
			}
			g.Nodes[name].Coefficients[i] = c
		}
	}

	return nil
}

func (g *Graph) SetConstant(name, state string) error {
	if _, ok := g.Nodes[name]; !ok {
		return errors.New("node with this name does not exist")
	}

	if !isContainString(g.Nodes[name].States, state) {
		return errors.New("state does not exist")
	}

	for i := range g.Nodes[name].States {
		if g.Nodes[name].States[i] == state {
			g.Nodes[name].Coefficients[i] = 1
		} else {
			g.Nodes[name].Coefficients[i] = 0
		}
	}

	g.Nodes[name].constant = true

	return nil
}

func (g *Graph) UnsetConstant(name string) error {
	if _, ok := g.Nodes[name]; !ok {
		return errors.New("node with this name does not exist")
	}

	if !g.Nodes[name].constant {
		return errors.New("this node is not constant")
	}

	err := g.calculateCoefficients(name)
	if err != nil {
		return err
	}

	g.Nodes[name].constant = false

	return nil
}

func (g *Graph) GetConstantState(name string) (string, error) {
	if _, ok := g.Nodes[name]; !ok {
		return "", errors.New("node with this name does not exist")
	}

	if !g.Nodes[name].constant {
		return "", errors.New("this node is not constant")
	}

	for i := range g.Nodes[name].States {
		if g.Nodes[name].Coefficients[i] == 1 {
			return g.Nodes[name].States[i], nil
		}
	}

	return "", nil
}

func (g *Graph) IsConstant(name string) (bool, error) {
	if _, ok := g.Nodes[name]; !ok {
		return false, errors.New("node with this name does not exist")
	}

	return g.Nodes[name].constant, nil
}

func (g *Graph) GetInternalCoefficient(name, state string, fullStates map[string]string) (float64, error) {
	if _, ok := g.Nodes[name]; !ok {
		return 0, errors.New("node with this name does not exist")
	}

	var misses []map[string]string
	var indexes []int

	for i := range g.Nodes[name].InternalCoefficients[state] {
		find := true
		miss := make(map[string]string)
		for node := range g.Nodes[name].InternalCoefficients[state][i].States {
			if _, ok := fullStates[node]; !ok {
				miss[node] = g.Nodes[name].InternalCoefficients[state][i].States[node]
				continue
			}
			if fullStates[node] != g.Nodes[name].InternalCoefficients[state][i].States[node] {
				find = false
				break
			}
		}
		if find {
			indexes = append(indexes, i)
			misses = append(misses, miss)
		}
	}

	if len(indexes) == 1 {
		return g.Nodes[name].InternalCoefficients[state][indexes[0]].Value, nil
	} else if len(indexes) > 1 && misses != nil {
		var coef float64
		for i := range indexes {
			summand := g.Nodes[name].InternalCoefficients[state][indexes[i]].Value
			for node := range misses[i] {
				isConst, _ := g.IsConstant(node)
				constState, _ := g.GetConstantState(node)
				if !isConst || isConst && misses[i][node] == constState {
					c, _ := g.GetInternalCoefficient(node, misses[i][node], fullStates)
					summand *= c
				}
			}
			coef += summand
		}
		return coef, nil
	}

	return 0, nil
}

func getMaxLen(nodes ...*Node) int {
	max := len(nodes[0].States)
	for i := range nodes {
		if len(nodes[i].States) > max {
			max = len(nodes[i].States)
		}
	}
	return max
}

func getMaxStrLen(nodes []*Node) int {
	max := 0
	for i := range nodes {
		for j := range nodes[i].States {
			val := fmt.Sprintf("%s - %.2f%%", nodes[i].States[j], nodes[i].Coefficients[j]*100)
			if len(val) > max {
				max = len(val)
			}
			if len(nodes[i].States[j]) > max {
				max = len(nodes[i].States[i])
			}
		}
	}
	return max
}

func (g *Graph) PrettyPrint(numCol int) {
	nodes := make([]*Node, 0, len(g.Nodes))
	for name := range g.Nodes {
		nodes = append(nodes, g.Nodes[name])
	}

	maxLen := getMaxStrLen(nodes)

	for i := 0; i < len(nodes); i += numCol {
		var end int
		for end = i; end < i+numCol && end < len(nodes); end++ {
		}
		max := getMaxLen(nodes[i:end]...)
		for j := i; j < end; j++ {
			numSpace := (maxLen - len(nodes[j].Name)) + numSpaces
			fmt.Printf("%s%s%s|", strings.Repeat(" ", numSpace/2), nodes[j].Name, strings.Repeat(" ",
				maxLen-len(nodes[j].Name)+numSpaces-numSpace/2))
		}
		fmt.Printf("\n%s\n", strings.Repeat("-", (maxLen+numSpaces+1)*numCol))
		for j := 0; j < max; j++ {
			for k := i; k < end; k++ {
				if j < len(nodes[k].States) {
					val := fmt.Sprintf("%s - %.2f%%", nodes[k].States[j], nodes[k].Coefficients[j]*100)
					numSpace := (maxLen - len(val)) + numSpaces
					fmt.Printf("%s%s%s|", strings.Repeat(" ", numSpace/2), val, strings.Repeat(" ",
						maxLen-len(val)+numSpaces-numSpace/2))
				} else {
					fmt.Printf("%s|", strings.Repeat(" ", maxLen+numSpaces))
				}
			}
			fmt.Println()
		}
		fmt.Printf("\n%s\n", strings.Repeat("-", (maxLen+numSpaces+1)*numCol))
	}
}
