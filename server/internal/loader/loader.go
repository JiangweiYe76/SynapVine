package loader

import (
	"encoding/json"
	"os"

	"ai-graph-server/internal/model"
)

// LoadGraphData reads and parses the graph data from a JSON file
// Returns a GraphData struct containing nodes and edges
func LoadGraphData(path string) (*model.GraphData, error) {
	// Read the entire file into memory
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse JSON data into GraphData struct
	var graph model.GraphData
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, err
	}

	return &graph, nil
}
