package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// POSIX make spec:
// https://pubs.opengroup.org/onlinepubs/9699919799/utilities/make.html

type Target struct {
	Prerequisites []string
	Commands      []string
}

func NewTarget(prerequisites []string) *Target {
	t := Target{
		Prerequisites: prerequisites,
	}
	return &t
}

func (t *Target) AddCommand(command string) {
	t.Commands = append(t.Commands, command)
}

type Graph map[string]*Target

// TODO: loop check
// TODO: concurrency
func execute(graph Graph, name string) error {
	// lookup current target
	target, ok := graph[name]
	if !ok {
		return fmt.Errorf("target does not exist: %s", name)
	}

	// execute any prerequisites (recursive call)
	for _, preprequisite := range target.Prerequisites {
		execute(graph, preprequisite)
	}

	// execute current target (base case)
	for _, command := range target.Commands {
		fmt.Println(command)

		fields := strings.Fields(command)
		cmd, args := fields[0], fields[1:]
		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			return err
		}

		fmt.Print(string(out))
	}

	return nil
}

func run() error {
	// TODO: flags
	fileName := "Makefile"

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	graph := make(map[string]*Target)

	var current string
	for _, line := range lines {
		// ignore empty lines
		if len(line) == 0 {
			continue
		}
		// ignore comments
		if line[0] == '#' {
			continue
		}
		// ignore dot directives
		if line[0] == '.' {
			continue
		}

		// add commands to the current target
		if line[0] == '\t' {
			target, ok := graph[current]
			if !ok {
				return fmt.Errorf("target does not exist: %s", target)
			}

			command := strings.TrimSpace(line)
			target.AddCommand(command)

			continue
		}

		// create target and add prerequisites
		fields := strings.Fields(line)
		name, prerequisites := fields[0], fields[1:]
		name, _ = strings.CutSuffix(name, ":")
		graph[name] = NewTarget(prerequisites)

		// update current target
		current = name
	}

	name := "default"
	if len(os.Args) >= 2 {
		name = os.Args[1]
	}

	return execute(graph, name)
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
