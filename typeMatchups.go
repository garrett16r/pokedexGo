package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
)

func getSingleTypeWeaknessesAndResistances(pkmnType string, typesDir string) ([]string, []string) {
	typeFile := fmt.Sprintf("%s/%s.json", typesDir, strings.ToLower(pkmnType))
	typeDataJson, err := os.ReadFile(typeFile)
	if err != nil {
		log.Fatal(err)
	}

	var weaknesses []string
	var resistances []string

	for i := range 19 {
		weaknesses = append(weaknesses, strings.Title(fastjson.GetString(typeDataJson, "damage_relations", "double_damage_from", strconv.Itoa(i), "name")), "2")
		resistances = append(resistances, strings.Title(fastjson.GetString(typeDataJson, "damage_relations", "half_damage_from", strconv.Itoa(i), "name")), "0.5")
		resistances = append(resistances, strings.Title(fastjson.GetString(typeDataJson, "damage_relations", "no_damage_from", strconv.Itoa(i), "name")), "0")
	}

	var weaknessesTrimmed []string
	var resistancesTrimmed []string

	for i := 0; i < len(weaknesses); i += 2 {
		if weaknesses[i] != "" {
			weaknessesTrimmed = append(weaknessesTrimmed, weaknesses[i], "2")
		}

		if resistances[i] != "" {
			if resistances[i+1] == "0" {
				resistancesTrimmed = append(resistancesTrimmed, resistances[i], "0")
			} else {
				resistancesTrimmed = append(resistancesTrimmed, resistances[i], "0.5")
			}
		}
	}

	return weaknessesTrimmed, resistancesTrimmed
}

func getMultiTypeWeaknessesAndResistances(pkmnType1 string, pkmnType2 string, typesDir string) ([]string, []string) {
	var weaknesses []string
	var resistances []string

	// Load weaknesses and resistances for each of the pokemon's two types
	type1Weaknesses, type1Resistances := getSingleTypeWeaknessesAndResistances(pkmnType1, typesDir)
	type2Weaknesses, type2Resistances := getSingleTypeWeaknessesAndResistances(pkmnType2, typesDir)

	// Perform a check for negated or doubled weaknesses and resistances
	for i := 0; i < len(type1Weaknesses); i += 2 {
		currentWeak := type1Weaknesses[i]

		if slices.Contains(type2Weaknesses, currentWeak) {
			weaknesses = append(weaknesses, currentWeak, "4")
		} else if slices.Contains(type2Resistances, currentWeak) { // Weakness cancelled out by resistance, so ignore it
			continue
		} else {
			weaknesses = append(weaknesses, currentWeak, "2")
		}
	}

	for i := 0; i < len(type2Weaknesses); i += 2 {
		currentWeak := type2Weaknesses[i]

		if slices.Contains(weaknesses, currentWeak) {
			continue
		}

		if slices.Contains(type1Resistances, currentWeak) { // Weakness cancelled out by resistance, so ignore it
			continue
		} else {
			weaknesses = append(weaknesses, currentWeak, "2")
		}
	}

	for i := 0; i < len(type1Resistances); i += 2 {
		currentRes := type1Resistances[i]

		if type1Resistances[i+1] == "0" {
			resistances = append(resistances, currentRes, "0")
			continue
		}

		if slices.Contains(type2Resistances, currentRes) {
			resistances = append(resistances, currentRes, "0.25")
		} else if slices.Contains(type2Weaknesses, currentRes) { // Resistance cancelled out by weakness, so ignore it
			continue
		} else {
			resistances = append(resistances, currentRes, "0.5")
		}
	}

	for i := 0; i < len(type2Resistances); i += 2 {
		currentRes := type2Resistances[i]

		// Prevent duplicates from being added to final list
		if slices.Contains(resistances, currentRes) {
			continue
		}

		if type2Resistances[i+1] == "0" {
			resistances = append(resistances, currentRes, "0")
			continue
		}

		if slices.Contains(type1Weaknesses, currentRes) { // Resistance cancelled out by weakness, so ignore it
			continue
		} else {
			resistances = append(resistances, currentRes, "0.5")
		}
	}

	return weaknesses, resistances
}
