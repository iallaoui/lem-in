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

	scanner := bufio.NewScanner(file)
	var lastCmd string
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
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
				return nil, fmt.Errorf("invalid room line: %q", line)
			}
			name := parts[0]
			x, _ := strconv.Atoi(parts[1])
			y, _ := strconv.Atoi(parts[2])
			farm.Rooms[name] = &Room{Name: name, X: x, Y: y}
			if lastCmd == "##start" {
				farm.Start = name
			}
			if lastCmd == "##end" {
				farm.End = name
			}
			lastCmd = ""
			continue
		}
		if strings.Contains(line, "-") {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid tunnel line: %q", line)
			}
			a, b := parts[0], parts[1]
			if farm.Rooms[a] != nil {
				farm.Rooms[a].Links = append(farm.Rooms[a].Links, b)
			}
			if farm.Rooms[b] != nil {
				farm.Rooms[b].Links = append(farm.Rooms[b].Links, a)
			}
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
func contains(path []string, room string) bool {
	for _, r := range path {
		if r == room {
			return true
		}
	}
	return false
}

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

// ----- Optimized simulation -----
// ----- Optimized simulation -----
// ----- Optimized simulation with balancing -----
func simulateAnts(f *Farm, paths [][]string) {
	if len(paths) == 0 {
		fmt.Println("No valid paths found!")
		return
	}

	ants := f.Ants
	positions := make([]int, ants)
	antPaths := make([]int, ants)

	// --- 1) حساب أطوال المسارات
	type pathInfo struct {
		index int
		len   int
	}
	infos := make([]pathInfo, len(paths))
	for i, path := range paths {
		infos[i] = pathInfo{i, len(path) - 1}
	}

	// --- 2) توزيع محسّن للنمل
	antsAssigned := make([]int, len(paths))
	for a := 0; a < ants; a++ {
		best := 0
		bestScore := infos[0].len + antsAssigned[infos[0].index]
		for _, pi := range infos {
			score := pi.len + antsAssigned[pi.index]
			if score < bestScore {
				best = pi.index
				bestScore = score
			}
		}
		antPaths[a] = best
		antsAssigned[best]++
		positions[a] = 0
	}

	// --- 3) Simulation
	round := 1
	antsFinished := 0
	totalMoves := 0

	fmt.Printf("\n=== Starting Optimized Simulation ===\n")
	fmt.Printf("Ants: %d, Paths: %d\n", ants, len(paths))
	for i, p := range paths {
		fmt.Printf("Path %d (len %d): %d ants\n", i+1, len(p)-1, antsAssigned[i])
	}

	for antsFinished < ants {
		moves := []string{}
		occupied := make(map[string]bool)

		for ant := 0; ant < ants; ant++ {
			currentPath := paths[antPaths[ant]]
			if positions[ant] >= len(currentPath)-1 {
				continue
			}
			nextPos := positions[ant] + 1
			nextRoom := currentPath[nextPos]

			if nextRoom == f.End || !occupied[nextRoom] {
				positions[ant] = nextPos
				moves = append(moves, fmt.Sprintf("L%d-%s", ant+1, nextRoom))
				if nextRoom != f.End {
					occupied[nextRoom] = true
				}
				if nextRoom == f.End {
					antsFinished++
				}
				totalMoves++
			}
		}

		if len(moves) > 0 {
			fmt.Printf("Turn %3d: %s\n", round, strings.Join(moves, " "))
		}
		round++
	}

	// --- 4) Statistics
	fmt.Printf("\n=== Simulation Statistics ===\n")
	fmt.Printf("Total turns: %d\n", round-1)
	theoretical := (ants + sum(pathLengths(paths)) - 1) / len(paths)
	fmt.Printf("Theoretical minimum: %d\n", theoretical)
	fmt.Printf("Efficiency: %.2f%%\n", float64(theoretical)/float64(round-1)*100)
}
// helper: يحسب أطوال جميع المسارات
func pathLengths(paths [][]string) []int {
	lengths := make([]int, len(paths))
	for i, path := range paths {
		lengths[i] = len(path) - 1
	}
	return lengths
}

// helper
func sum(arr []int) int {
	total := 0
	for _, v := range arr {
		total += v
	}
	return total
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
	simulateAnts(farm, finalPaths)
}
