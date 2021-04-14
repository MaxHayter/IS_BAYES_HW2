package bayes

import "github.com/MaxHayter/IS_BAYES_HW2/internal/graph"

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

func (b *Bayes) findInterestNodes(name string) map[string]struct{} {
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
	for node := range interestNodes { // TODO для корректной работы замените "interestNodes" на "b.Graph.Nodes"
		ret[node] = struct{}{}
	}

	return ret
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
	var sum float64
	for i := range table {
		table[i].Probability = 1
		for node := range table[i].States {
			c, _ := b.Graph.GetInternalCoefficient(node, table[i].States[node], table[i].States)
			table[i].Probability *= c
		}
		sum += table[i].Probability
	}

	return table
}

func normalize(table []*Row) []*Row {
	var sum float64 = 0
	for i := range table {
		sum += table[i].Probability
	}

	for i := range table {
		table[i].Probability /= sum
	}

	return table
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

func (b *Bayes) RecalculateWithObs(name string) error {
	interestNodes := b.findInterestNodes(name)

	obsNodes := make(map[string]struct{})
	for node := range interestNodes {
		if ok, _ := b.Graph.IsConstant(node); ok {
			obsNodes[node] = struct{}{}
		}
	}

	table := b.makeTable(interestNodes, obsNodes)

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
