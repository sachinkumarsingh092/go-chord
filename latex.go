package main

import (
	"fmt"
	"os"
	"text/template"
)

func GenerateChordVisualization(nodes []*Node) error {
	ringSize := uint32(1 << m)

	var nodeData []NodeData
	var arrows []ArrowData
	var tables []TableData
	var idLabels []IDLabel

	// Create ID labels for the entire circle
	for id := uint32(0); id < ringSize; id++ {
		angle := (float64(id) / float64(ringSize)) * 360.0
		idLabels = append(idLabels, IDLabel{
			ID:    id,
			Angle: angle,
		})
	}

	// Create node data for visualization
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		angle := (float64(node.id) / float64(ringSize)) * 360.0

		// Determine anchor based on angle
		anchor := "west"
		anchorID := "west"
		if angle > 45 && angle < 135 {
			anchor = "south"
			anchorID = "south"
		} else if angle >= 135 && angle < 225 {
			anchor = "east"
			anchorID = "east"
		} else if angle >= 225 && angle < 315 {
			anchor = "north"
			anchorID = "north"
		}

		nodeData = append(nodeData, NodeData{
			Label:    node.address,
			ID:       node.id,
			Angle:    angle,
			Anchor:   anchor,
			AnchorID: anchorID,
		})

		// Create finger table data
		var fingers []FingerData
		for j := 0; j < m; j++ {
			start := (node.id + (1 << j)) % ringSize
			fingers = append(fingers, FingerData{
				Index:    j,
				Start:    start,
				FingerID: node.finger[j].id,
			})

			// Create arrows for finger pointers
			fromAngle := angle
			toAngle := (float64(node.finger[j].id) / float64(ringSize)) * 360.0

			bend := "left=15"
			if toAngle < fromAngle {
				bend = "right=15"
			}

			arrows = append(arrows, ArrowData{
				From: fmt.Sprintf("%.1f", fromAngle),
				To:   fmt.Sprintf("%.1f", toAngle),
				Bend: bend,
			})
		}

		// Position tables around the circle
		tableX := 0.0
		tableY := 0.0
		tableAnchor := "north"

		if i == 0 {
			tableX = -6.5
			tableY = 2.0
			tableAnchor = "east"
		} else if i == 1 {
			tableX = 0
			tableY = 5.5
			tableAnchor = "south"
		} else if i == 2 {
			tableX = 6.5
			tableY = 2.0
			tableAnchor = "west"
		}

		tables = append(tables, TableData{
			NodeLabel: node.address,
			NodeID:    node.id,
			Fingers:   fingers,
			X:         tableX,
			Y:         tableY,
			Anchor:    tableAnchor,
		})
	}

	// Execute template
	f, err := os.Create("chord_visualization.tex")
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.New("chord").Parse(tikzTemplate)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"RingSize": ringSize,
		"Nodes":    nodeData,
		"Arrows":   arrows,
		"Tables":   tables,
		"IDLabels": idLabels,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	fmt.Println("LaTeX file generated: chord_visualization.tex")
	fmt.Println("\nNode Information:")
	for _, node := range nodes {
		fmt.Printf("\nNode: %s (ID: %d)\n", node.address, node.id)
		fmt.Println("Finger Table:")
		for i := 0; i < m; i++ {
			start := (node.id + (1 << i)) % ringSize
			fmt.Printf("  [%d] start: %d -> finger: %d\n", i, start, node.finger[i].id)
		}
	}

	return nil
}

type NodeData struct {
	Label    string
	ID       uint32
	Angle    float64
	Anchor   string
	AnchorID string
}

type ArrowData struct {
	From string
	To   string
	Bend string
}

type FingerData struct {
	Index    int
	Start    uint32
	FingerID uint32
}

type TableData struct {
	NodeLabel string
	NodeID    uint32
	Fingers   []FingerData
	X         float64
	Y         float64
	Anchor    string
}

type IDLabel struct {
	ID    uint32
	Angle float64
}

var tikzTemplate = `\documentclass[border=10pt]{standalone}
\usepackage{tikz}
\usepackage{booktabs}
\usepackage{array}
\usetikzlibrary{arrows.meta, positioning}
\definecolor{nodecolor}{RGB}{59, 130, 246}
\definecolor{fingercolor}{RGB}{168, 85, 247}

\begin{document}
\begin{tikzpicture}[>=Stealth, line cap=round, line join=round]
  
  % Draw the ring
  \def\R{3.5}
  \draw[thick, gray!50] (0,0) circle (\R);
  
  % Ring size annotation
  \node[gray, font=\footnotesize] at (0, -\R-0.5) {Ring Size: {{.RingSize}}};
  
  % Label circle with IDs
  {{range $id := .IDLabels -}}
  \node[gray, font=\tiny] at ({{$id.Angle}}:3.2) { {{- $id.ID -}} };
  {{end}}
  
  % Draw nodes on the circle
  {{range $idx, $node := .Nodes -}}
  \fill[nodecolor] ({{$node.Angle}}:3.5) circle (3pt);
  \node[nodecolor, font=\bfseries\small, anchor={{$node.Anchor}}] at ({{$node.Angle}}:3.8) { {{- $node.Label -}} };
  \node[gray, font=\tiny, anchor={{$node.AnchorID}}] at ({{$node.Angle}}:4.1) {({{$node.ID}})};
  {{end}}
  
  % Draw finger table arrows
  {{range $arrow := .Arrows -}}
  \draw[->, fingercolor!60, thick, shorten >=2pt] ({{$arrow.From}}:3.5) 
    to[bend {{$arrow.Bend}}] ({{$arrow.To}}:3.5);
  {{end}}
  
  % Finger tables positioned around the circle
  {{range $idx, $table := .Tables -}}
  \node[anchor={{$table.Anchor}}, font=\small] at ({{$table.X}}, {{$table.Y}}) {
    \begin{tabular}{|c|c|c|}
    \hline
    \multicolumn{3}{|c|}{\textbf{ {{- $table.NodeLabel -}} (ID: {{$table.NodeID}})}} \\
    \hline
    \textbf{i} & \textbf{start} & \textbf{finger} \\
    \hline
    {{range $finger := $table.Fingers -}}
    {{$finger.Index}} & {{$finger.Start}} & {{$finger.FingerID}} \\
    {{end -}}
    \hline
    \end{tabular}
  };
  {{end}}
  
\end{tikzpicture}
\end{document}
`
