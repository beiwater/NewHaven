package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Stats struct {
	Files      int
	Lines      int
	Blank      int
	Comment    int
	Code       int
}

func main() {
	root := "."
	dirStats := make(map[string]*Stats)
	mdStats := &Stats{}
	total := &Stats{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == "scripts" {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle Markdown separately
		if strings.HasSuffix(info.Name(), ".md") {
			stats, _ := analyzeFile(path)
			updateStats(mdStats, stats)
			return nil
		}

		// Only process .go files for the main table
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		stats, err := analyzeFile(path)
		if err != nil {
			fmt.Printf("Error analyzing %s: %v\n", path, err)
			return nil
		}

		dir := filepath.Dir(path)
		if _, ok := dirStats[dir]; !ok {
			dirStats[dir] = &Stats{}
		}

		updateStats(dirStats[dir], stats)
		updateStats(total, stats)

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking tree: %v\n", err)
		return
	}

	fmt.Println("### Go Source Code Statistics")
	printResults(dirStats, total)

	fmt.Println("\n### Markdown Documentation Statistics (Separate)")
	fmt.Printf("%-40s %8s %8s\n", "Type", "Files", "Lines")
	fmt.Println(strings.Repeat("-", 58))
	fmt.Printf("%-40s %8d %8d\n", "Markdown (.md)", mdStats.Files, mdStats.Lines)
}

func analyzeFile(path string) (*Stats, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := &Stats{Files: 1}
	scanner := bufio.NewScanner(file)
	inMultiComment := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		stats.Lines++

		if line == "" {
			stats.Blank++
			continue
		}

		if inMultiComment {
			stats.Comment++
			if strings.Contains(line, "*/") {
				inMultiComment = false
			}
			continue
		}

		if strings.HasPrefix(line, "/*") {
			stats.Comment++
			if !strings.Contains(line, "*/") {
				inMultiComment = true
			}
			continue
		}

		if strings.HasPrefix(line, "//") {
			stats.Comment++
			continue
		}

		stats.Code++
	}

	return stats, scanner.Err()
}

func updateStats(target, source *Stats) {
	target.Files += source.Files
	target.Lines += source.Lines
	target.Blank += source.Blank
	target.Comment += source.Comment
	target.Code += source.Code
}

func printResults(dirStats map[string]*Stats, total *Stats) {
	fmt.Printf("%-40s %8s %8s %8s %8s %8s\n", "Directory", "Files", "Lines", "Code", "Comment", "Blank")
	fmt.Println(strings.Repeat("-", 85))

	var dirs []string
	for dir := range dirStats {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	for _, dir := range dirs {
		s := dirStats[dir]
		fmt.Printf("%-40s %8d %8d %8d %8d %8d\n", dir, s.Files, s.Lines, s.Code, s.Comment, s.Blank)
	}

	fmt.Println(strings.Repeat("-", 85))
	fmt.Printf("%-40s %8d %8d %8d %8d %8d\n", "TOTAL", total.Files, total.Lines, total.Code, total.Comment, total.Blank)
}
