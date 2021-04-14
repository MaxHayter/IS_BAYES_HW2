package importing

import (
	"encoding/json"
	"io"
	"os"

	"github.com/MaxHayter/IS_BAYES_HW2/internal/graph"
)

func ImportGraph(path string) (*graph.Graph, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	imp := &Import{}

	err = json.Unmarshal(data, imp)
	if err != nil {
		return nil, err
	}

	gr := graph.NewGraph()

	for i := range imp.Nodes {
		err = gr.CreateNode(imp.Nodes[i].Name)
		if err != nil {
			return nil, err
		}

		err = gr.SetStates(imp.Nodes[i].Name, imp.Nodes[i].States)
		if err != nil {
			return nil, err
		}
	}

	for i := range imp.Edges {
		err = gr.AddEdges(imp.Edges[i].From, imp.Edges[i].To...)
		if err != nil {
			return nil, err
		}
	}

	for i := range imp.Coefficients {
		for j := range imp.Coefficients[i].Data {
			err = gr.SetInternalCoefficients(imp.Coefficients[i].Node, imp.Coefficients[i].Data[j].States,
				imp.Coefficients[i].Data[j].Values, imp.Coefficients[i].Data[j].Fields)
			if err != nil {
				return nil, err
			}
		}
	}

	return gr, nil
}
