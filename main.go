package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Room structure
type Room struct {
	Name  string
	X, Y  int
	Links []string
}

// Farm structure
type Farm struct {
	Ants  int
	Rooms map[string]*Room
	Start string
	End   string
}

// ----- Parse input -----
func parseInput(filename string) (*Farm, error) {
	farm := &Farm{Rooms: make(map[string]*Room)}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	startSet := false
	endSet := false
	scanner := bufio.NewScanner(file)
	var lastCmd string
	lineCount := 0
	coords := make(map[string]bool) // check duplicate coordinates

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || (strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "##")) {
			continue
		}
		if lineCount == 0 {
			ants, err := strconv.Atoi(line)
			if err != nil || ants <= 0 {
				return nil, fmt.Errorf("invalid number of ants: %q", line)
			}
			farm.Ants = ants
			lineCount++
			continue
		}
		if strings.HasPrefix(line, "##") {
			lastCmd = line
			continue
		}
		if strings.Contains(line, " ") {
			parts := strings.Fields(line)
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid room definition: %q", line)
			}
			name := parts[0]
			if _, exists := farm.Rooms[name]; exists {
				return nil, fmt.Errorf("duplicate room name: %q", name)
			}
			x, err1 := strconv.Atoi(parts[1])
			y, err2 := strconv.Atoi(parts[2])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("invalid coordinates for room %q", name)
			}
			coordKey := fmt.Sprintf("%d-%d", x, y)
			if coords[coordKey] {
				return nil, fmt.Errorf("duplicate coordinates (%d,%d)", x, y)
			}
			coords[coordKey] = true
			farm.Rooms[name] = &Room{Name: name, X: x, Y: y}

			if lastCmd == "##start" {
				if startSet {
					return nil, fmt.Errorf("more than one start room defined")
				}
				farm.Start = name
				startSet = true
			}
			if lastCmd == "##end" {
				if endSet {
					return nil, fmt.Errorf("more than one end room defined")
				}
				farm.End = name
				endSet = true
			}
			lastCmd = ""
			continue
		}
		if strings.Contains(line, "-") {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid tunnel line: %q", line)
			}
			a, b := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if farm.Rooms[a] == nil || farm.Rooms[b] == nil {
				return nil, fmt.Errorf("tunnel references unknown room(s): %q", line)
			}
			farm.Rooms[a].Links = append(farm.Rooms[a].Links, b)
			farm.Rooms[b].Links = append(farm.Rooms[b].Links, a)
		} else {
			return nil, fmt.Errorf("invalid line format: %q", line)
		}
	}

	if farm.Start == "" || farm.End == "" {
		return nil, fmt.Errorf("missing start or end room")
	}
	return farm, nil
}

// ----- Optimized BFS to find shortest path avoiding blocked rooms -----
func bfsShortestPath(f *Farm, startNeighbor string, blockedRooms map[string]bool) []string {
	queue := [][]string{{f.Start, startNeighbor}}
	visited := make(map[string]bool)
	visited[f.Start] = true
	visited[startNeighbor] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		last := path[len(path)-1]

		if last == f.End {
			return path
		}

		for _, next := range f.Rooms[last].Links {
			if !visited[next] && !blockedRooms[next] {
				visited[next] = true
				newPath := make([]string, len(path))
				copy(newPath, path)
				newPath = append(newPath, next)
				queue = append(queue, newPath)
			}
		}
	}
	return nil
}

// ----- Find non-overlapping paths for each neighbor -----
func findNonOverlappingPaths(f *Farm) [][]string {
	var selectedPaths [][]string
	blockedRooms := make(map[string]bool)

	// Sort neighbors by number of links (try neighbors with fewer connections first)
	neighbors := make([]string, len(f.Rooms[f.Start].Links))
	copy(neighbors, f.Rooms[f.Start].Links)

	// Simple sorting by number of links
	for i := 0; i < len(neighbors)-1; i++ {
		for j := i + 1; j < len(neighbors); j++ {
			if len(f.Rooms[neighbors[j]].Links) < len(f.Rooms[neighbors[i]].Links) {
				neighbors[i], neighbors[j] = neighbors[j], neighbors[i]
			}
		}
	}

	// Find path for each neighbor
	for _, neighbor := range neighbors {
		path := bfsShortestPath(f, neighbor, blockedRooms)
		if path != nil {
			selectedPaths = append(selectedPaths, path)

			// Block intermediate rooms from this path (except start and end)
			for i := 1; i < len(path)-1; i++ {
				blockedRooms[path[i]] = true
			}
		}
	}

	return selectedPaths
}

// ----- Find all shortest paths for each neighbor (without blocking) -----
func findAllShortestPaths(f *Farm) [][]string {
	var allPaths [][]string

	for _, neighbor := range f.Rooms[f.Start].Links {
		// Use BFS to find shortest path for this neighbor
		queue := [][]string{{f.Start, neighbor}}
		visited := make(map[string]bool)
		visited[f.Start] = true
		visited[neighbor] = true

		var shortestPath []string

		for len(queue) > 0 && shortestPath == nil {
			path := queue[0]
			queue = queue[1:]
			current := path[len(path)-1]

			if current == f.End {
				shortestPath = path
				break
			}

			for _, next := range f.Rooms[current].Links {
				if !visited[next] {
					visited[next] = true
					newPath := make([]string, len(path))
					copy(newPath, path)
					newPath = append(newPath, next)
					queue = append(queue, newPath)
				}
			}
		}

		if shortestPath != nil {
			allPaths = append(allPaths, shortestPath)
		}
	}

	return allPaths
}

// ----- Helper -----
// ----- Check if two paths share intermediate rooms -----
func pathsShareRooms(path1, path2 []string) bool {
	// Create set of intermediate rooms for path1
	rooms1 := make(map[string]bool)
	for i := 1; i < len(path1)-1; i++ {
		rooms1[path1[i]] = true
	}

	// Check if path2 uses any of these rooms
	for i := 1; i < len(path2)-1; i++ {
		if rooms1[path2[i]] {
			return true
		}
	}
	return false
}

// ----- Select best non-conflicting paths -----
func selectBestPaths(f *Farm, allPaths [][]string) [][]string {
	if len(allPaths) == 0 {
		return nil
	}

	// Sort paths by length (shortest first)
	for i := 0; i < len(allPaths)-1; i++ {
		for j := i + 1; j < len(allPaths); j++ {
			if len(allPaths[j]) < len(allPaths[i]) {
				allPaths[i], allPaths[j] = allPaths[j], allPaths[i]
			}
		}
	}

	var selected [][]string
	usedRooms := make(map[string]bool)

	for _, path := range allPaths {
		conflict := false

		// Check if this path conflicts with any selected path
		for _, selectedPath := range selected {
			if pathsShareRooms(path, selectedPath) {
				conflict = true
				break
			}
		}

		if !conflict {
			selected = append(selected, path)
			// Mark intermediate rooms as used
			for i := 1; i < len(path)-1; i++ {
				usedRooms[path[i]] = true
			}
		}
	}

	return selected
}


func distributeAnts(ants int, paths [][]string) [][]int {
	lengths := make([]int, len(paths))
	for i, p := range paths {
		lengths[i] = len(p) - 1
	}

	distribution := make([][]int, len(paths))
	assigned := make([]int, len(paths))

	for a := 1; a <= ants; a++ {
		best := 0
		bestScore := lengths[0] + assigned[0]
		for i := 1; i < len(paths); i++ {
			score := lengths[i] + assigned[i]
			if score < bestScore {
				best = i
				bestScore = score
			}
		}
		distribution[best] = append(distribution[best], a)
		assigned[best]++
	}
	return distribution
}

func simulateAnts(paths [][]string, antDistribution [][]int) string {
	var finalResult string
	type AntPosition struct {
		ant  int
		path int
		step int
	}

	var antPositions []AntPosition
	for pathIndex, ants := range antDistribution {
		for _, ant := range ants {
			antPositions = append(antPositions, AntPosition{ant, pathIndex, 0})
		}
	}
	for len(antPositions) > 0 {
		var moves []string
		var newPositions []AntPosition
		usedLinks := make(map[string]bool)

		for _, pos := range antPositions {
			if pos.step < len(paths[pos.path])-1 {
				currentRoom := paths[pos.path][pos.step]
				nextRoom := paths[pos.path][pos.step+1]
				link := currentRoom + "-" + nextRoom
				if !usedLinks[link] {
					moves = append(moves, fmt.Sprintf("L%d-%s", pos.ant, nextRoom))
					newPositions = append(newPositions, AntPosition{pos.ant, pos.path, pos.step + 1})
					usedLinks[link] = true
				} else {
					newPositions = append(newPositions, pos)
				}
			}
		}
		if len(moves) > 0 {
			finalResult += strings.Join(moves, " ")
			finalResult += "\n"
		}
		antPositions = newPositions
	}
	return finalResult
}


// ----- MAIN -----
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . input.txt")
		return
	}
	filename := os.Args[1]
	farm, err := parseInput(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Farm: %d ants, start=%s, end=%s\n", farm.Ants, farm.Start, farm.End)
	fmt.Printf("Start room has %d neighbors: %v\n", len(farm.Rooms[farm.Start].Links), farm.Rooms[farm.Start].Links)

	// Method 1: Find all shortest paths first
	fmt.Println("\n=== Finding all shortest paths ===")
	allPaths := findAllShortestPaths(farm)
	fmt.Printf("Found %d shortest paths:\n", len(allPaths))
	for i, p := range allPaths {
		fmt.Printf("Path %d: %v (length: %d)\n", i+1, p, len(p))
	}

	// Method 2: Select non-conflicting paths
	fmt.Println("\n=== Selecting non-conflicting paths ===")
	bestPaths := selectBestPaths(farm, allPaths)
	fmt.Printf("Selected %d non-conflicting paths:\n", len(bestPaths))
	for i, p := range bestPaths {
		fmt.Printf("Path %d: %v (length: %d)\n", i+1, p, len(p))
	}

	// Method 3: Find non-overlapping paths directly
	fmt.Println("\n=== Finding non-overlapping paths directly ===")
	nonOverlapPaths := findNonOverlappingPaths(farm)
	fmt.Printf("Found %d non-overlapping paths:\n", len(nonOverlapPaths))
	for i, p := range nonOverlapPaths {
		fmt.Printf("Path %d: %v (length: %d)\n", i+1, p, len(p))
	}

	// Use the best set of paths
	var finalPaths [][]string
	if len(nonOverlapPaths) > len(bestPaths) {
		finalPaths = nonOverlapPaths
	} else {
		finalPaths = bestPaths
	}

	if len(finalPaths) == 0 {
		fmt.Println("No valid paths found!")
		return
	}

	// Run simulation
	fmt.Println("\n=== Simulation ===")
	antDistribution := distributeAnts(farm.Ants, finalPaths)

	result := simulateAnts(finalPaths, antDistribution)
	fmt.Print(result)
}
