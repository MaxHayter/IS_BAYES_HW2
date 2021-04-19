package bayes

import (
	"github.com/MaxHayter/IS_BAYES_HW2/internal/graph"
)

func NewBayes(gr *graph.Graph) *Bayes {
	return &Bayes{Graph: gr}
}

func (b *Bayes) findAllDependents(from, now string, isChild bool, r map[string]map[string]struct{}) map[string]map[string]struct{} {
	if r[now] == nil {
		r[now] = make(map[string]struct{})
	}
	r[now][from] = struct{}{}

	isConst, _ := b.Graph.IsConstant(now)
	if isConst && !isChild {
		return b.findAllDependents(now, from, true, r)
	}
	for i := range b.Graph.Nodes[now].Children {
		ok := false
		if r[now] != nil {
			_, ok = r[now][b.Graph.Nodes[now].Children[i].Name]
		}

		if !ok && (!isConst) {
			if r[now] == nil {
				r[now] = make(map[string]struct{})
			}
			r[now][b.Graph.Nodes[now].Children[i].Name] = struct{}{}
			r = b.findAllDependents(now, b.Graph.Nodes[now].Children[i].Name, true, r)
		}
	}

	for i := range b.Graph.Nodes[now].Parents {
		ok := false
		if r[now] != nil {
			_, ok = r[now][b.Graph.Nodes[now].Parents[i].Name]
		}
		if !ok && (!isChild && !isConst || isChild && isConst) {
			if r[now] == nil {
				r[now] = make(map[string]struct{})
			}
			r[now][b.Graph.Nodes[now].Parents[i].Name] = struct{}{}
			r = b.findAllDependents(now, b.Graph.Nodes[now].Parents[i].Name, false, r)
		}
	}
	return r
}

func (b *Bayes) findInterestNodes(name string, all bool) map[string]struct{} {
	if all {
		ret := make(map[string]struct{}, len(b.Graph.Nodes))
		for node := range b.Graph.Nodes {
			ret[node] = struct{}{}
		}
		return ret
	}
	interestNodes := make(map[string]map[string]struct{})
	interestNodes[name] = make(map[string]struct{})

	for i := range b.Graph.Nodes[name].Children {
		if interestNodes[name] == nil {
			interestNodes[name] = make(map[string]struct{})
		}
		interestNodes[name][b.Graph.Nodes[name].Children[i].Name] = struct{}{}
		interestNodes = b.findAllDependents(name, b.Graph.Nodes[name].Children[i].Name, true, interestNodes)
	}

	for i := range b.Graph.Nodes[name].Parents {
		if interestNodes[name] == nil {
			interestNodes[name] = make(map[string]struct{})
		}
		interestNodes[name][b.Graph.Nodes[name].Parents[i].Name] = struct{}{}
		interestNodes = b.findAllDependents(name, b.Graph.Nodes[name].Parents[i].Name, false, interestNodes)
	}

	ret := make(map[string]struct{}, len(interestNodes))
	for node := range interestNodes {
		ret[node] = struct{}{}
		ret = b.completeParents(node, ret)
	}

	return ret
}

func (b *Bayes) completeParents(node string, nodes map[string]struct{}) map[string]struct{} {
	for parentIndex := range b.Graph.Nodes[node].Parents {
		nodes[b.Graph.Nodes[node].Parents[parentIndex].Name] = struct{}{}
		nodes = b.completeParents(b.Graph.Nodes[node].Parents[parentIndex].Name, nodes)
	}
	return nodes
}

func (b *Bayes) makeTable(interestNodes, obsNodes map[string]struct{}) []*Row {
	numVariant := 1
	for node := range interestNodes {
		if _, ok := obsNodes[node]; !ok {
			numVariant *= len(b.Graph.Nodes[node].States)
		}
	}

	table := make([]*Row, numVariant)

	for i := range table {
		table[i] = &Row{States: make(map[string]string, len(interestNodes))}
	}

	k := 1
	for node := range interestNodes {
		if _, ok := obsNodes[node]; ok {
			state, _ := b.Graph.GetConstantState(node)
			for i := 0; i < numVariant; i++ {
				table[i].States[node] = state
			}
		} else {
			lenStates := len(b.Graph.Nodes[node].States)
			for j := 0; j < k; j++ {
				for l := range b.Graph.Nodes[node].States {
					for m := (j * numVariant / k) + l*numVariant/(k*lenStates); m < (j*numVariant/k)+(l+1)*numVariant/(k*lenStates); m++ {
						table[m].States[node] = b.Graph.Nodes[node].States[l]
					}
				}
			}
			k *= lenStates
		}
	}

	for i := range table {
		table[i].Probability = 1
		for node := range table[i].States {
			c := b.getInternalCoefficient(node, table[i].States[node], table[i].States)
			table[i].Probability *= c
		}
	}
	return table
}

func (b *Bayes) getInternalCoefficient(name, state string, fullStates map[string]string) float64 {
	var find bool
	for i := range b.Graph.Nodes[name].InternalCoefficients[state] {
		find = true
		for node := range b.Graph.Nodes[name].InternalCoefficients[state][i].States {
			if fullStates[node] != b.Graph.Nodes[name].InternalCoefficients[state][i].States[node] {
				find = false
				break
			}
		}
		if find {
			return b.Graph.Nodes[name].InternalCoefficients[state][i].Value
		}
	}

	return 0
}

func (b *Bayes) SetConstant(name, state string) error {
	err := b.Graph.SetConstant(name, state)
	if err != nil {
		return err
	}

	return b.RecalculateWithObs(name)
}

func (b *Bayes) UnsetConstant(name string) error {
	err := b.Graph.UnsetConstant(name)
	if err != nil {
		return err
	}

	return b.RecalculateWithObs(name)
}

func (b *Bayes) recalculate(interestNodes, obsNodes map[string]struct{}, table []*Row) error {
	for node := range interestNodes {
		if _, ok := obsNodes[node]; ok {
			continue
		}
		var sum float64
		for stateIndex := range b.Graph.Nodes[node].States {
			var coefficient float64
			for i := range table {
				if table[i].States[node] == b.Graph.Nodes[node].States[stateIndex] {
					coefficient += table[i].Probability
				}
			}
			b.Graph.Nodes[node].Coefficients[stateIndex] = coefficient
			sum += coefficient
		}
		for stateIndex := range b.Graph.Nodes[node].States {
			b.Graph.Nodes[node].Coefficients[stateIndex] = b.Graph.Nodes[node].Coefficients[stateIndex] / sum
		}
	}

	return nil
}

func (b *Bayes) Recalculate() error {
	interestNodes := b.findInterestNodes("", true)

	table := b.makeTable(interestNodes, nil)

	return b.recalculate(interestNodes, nil, table)
}

func (b *Bayes) RecalculateWithObs(name string) error {
	interestNodes := b.findInterestNodes(name, false)

	obsNodes := make(map[string]struct{})
	for node := range interestNodes {
		if ok, _ := b.Graph.IsConstant(node); ok {
			obsNodes[node] = struct{}{}
		}
	}

	table := b.makeTable(interestNodes, obsNodes)

	return b.recalculate(interestNodes, obsNodes, table)
}
