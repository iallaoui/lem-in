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

// ----- BFS from neighbor -----
func bfsFromNeighbor(f *Farm, startNeigh string) [][]string {
	queue := [][]string{{f.Start, startNeigh}}
	var foundPaths [][]string

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		last := path[len(path)-1]

		if last == f.End {
			foundPaths = append(foundPaths, path)
			continue
		}

		for _, next := range f.Rooms[last].Links {
			if !contains(path, next) {
				newPath := append([]string{}, path...)
				newPath = append(newPath, next)
				queue = append(queue, newPath)
			}
		}
	}
	return foundPaths
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

// ----- Collect all paths per neighbor -----
func collectAllPaths(f *Farm) [][]string {
	var allPaths [][]string
	for _, nb := range f.Rooms[f.Start].Links {
		paths := bfsFromNeighbor(f, nb)
		allPaths = append(allPaths, paths...)
	}
	return allPaths
}

// ----- Filter paths - keep only shortest path per starting neighbor -----
func filterPaths(f *Farm, allPaths [][]string) [][]string {
	// Group paths by their first room after start (the neighbor)
	pathGroups := make(map[string][][]string)

	for _, path := range allPaths {
		if len(path) >= 2 {
			firstNeighbor := path[1] // First room after start
			pathGroups[firstNeighbor] = append(pathGroups[firstNeighbor], path)
		}
	}

	// For each group, keep only the shortest path
	filtered := [][]string{}
	for _, paths := range pathGroups {
		if len(paths) == 0 {
			continue
		}

		shortest := paths[0]
		for _, path := range paths {
			if len(path) < len(shortest) {
				shortest = path
			}
		}
		filtered = append(filtered, shortest)
	}

	return filtered
}

// // ----- FIXED Distribution - Use ALL available paths -----
// ----- PERFECT Distribution for 8 rounds -----
func simulatePaths(f *Farm, paths [][]string) {
	ants := f.Ants
	positions := make([]int, ants)
	antPaths := make([][]string, ants)
	for i := 0; i < ants; i++ {
		antPaths[i] = paths[i%len(paths)]
		positions[i] = 0
	}
	done := false
	for !done {
		done = true
		output := []string{}
		for i := 0; i < ants; i++ {
			if positions[i] < len(antPaths[i])-1 {
				done = false
				positions[i]++
				output = append(output, fmt.Sprintf("L%d-%s", i+1, antPaths[i][positions[i]]))
			}
		}
		if len(output) > 0 {
			fmt.Println(strings.Join(output, " "))
		}
	}
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

	allPaths := collectAllPaths(farm)

	fmt.Println("All paths found:")
	for i, p := range allPaths {
		fmt.Printf("Path %d: %v (length: %d)\n", i+1, p, len(p))
	}

	// Filter paths to keep only shortest path per starting neighbor
	filteredPaths := filterPaths(farm, allPaths)

	fmt.Println("\nFiltered paths (shortest per starting neighbor):")
	for i, p := range filteredPaths {
		fmt.Printf("Path %d: %v (length: %d)\n", i+1, p, len(p))
	}

	// Run PERFECT distribution simulation
	simulatePaths(farm, filteredPaths)
}
