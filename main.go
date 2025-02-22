package main

import (
	"context"
	"log"
	"maps"
	"obsidian-deps-view/analyzer"
	"obsidian-deps-view/obsidian"
	"os"
	"strings"
)

// todo parse not only std (do manual tests)
// todo explore behaviour when user works with not his own vault
// todo think about file perms
// todo move the std-related logic from obsidian.GraphCreator

// run then restart for package names will be added to obsidian spelling dict

func main() {
	obsidianVault, err := os.OpenRoot("/mnt/c/Users/Sasha/Documents/Obsidian/Std")
	if err != nil {
		log.Fatalln(err)
	}

	template, err := obsidian.LoadTemplate()
	if err != nil {
		log.Fatalln(err)
	}

	importsParser := analyzer.NewImportsParser()
	packages, err := importsParser.ParseImports(context.Background(), "std")
	if err != nil {
		log.Fatalln(err)
	}

	graphCreator := obsidian.NewGraphCreator(obsidianVault, template, "-")
	err = graphCreator.CreateGraph(packages)
	if err != nil {
		log.Fatalln(err)
	}

	spellingsLocation, err := obsidian.SpellingsLocation()
	if err != nil {
		log.Fatalln(err)
	}
	if spellingsLocation == "" {
		spellingsLocation = "/mnt/c/Users/Sasha/AppData/Roaming/obsidian/Custom Dictionary.txt"
	}

	spellings, err := os.OpenFile(spellingsLocation, os.O_RDWR|os.O_TRUNC, 0666) //todo perm
	if err != nil {
		log.Fatalln(err)
	}

	words := make([]string, 0)
	for pkg := range maps.Keys(packages) {
		for _, word := range strings.Split(pkg, "/") {
			if word != "" {
				words = append(words, word)
			}
		}
	}

	err = obsidian.AddSpellings(spellings, words)
	if err != nil {
		log.Fatalln(err)
	}
}
