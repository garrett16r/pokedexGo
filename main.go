package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
)

var apiPokemonUrl string = "https://pokeapi.co/api/v2/pokemon"
var apiTypeUrl string = "https://pokeapi.co/api/v2/type"
var pokemonNamesFile string
var cacheDir string
var typesDir string

func initialize() {
	programDir, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	cacheDir = fmt.Sprint(programDir, "/pokedexGo/cache")
	typesDir = fmt.Sprint(programDir, "/pokedexGo/types")
	pokemonNamesFile = fmt.Sprint(cacheDir, "/pokemonNames.txt")

	_, err = os.Stat(cacheDir)
	if os.IsNotExist(err) {
		log.Println("INFO: Cache folder not found. Creating...")

		os.MkdirAll(cacheDir, 0755)
		log.Println("INFO: Cache folder created at", cacheDir)

		log.Println("INFO: Downloading list of all Pokemon names...")

		os.Create(pokemonNamesFile)

		resp, err := http.Get(fmt.Sprint(apiPokemonUrl, "?limit=10000&offest=0"))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		v, _ := fastjson.ParseBytes(body)

		results := v.GetArray("results")
		var allPokemonNames string
		for _, r := range results {
			allPokemonNames += fmt.Sprint(string(r.GetStringBytes("name")), "\n")
		}

		err = os.WriteFile(pokemonNamesFile, []byte(allPokemonNames), 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("INFO: Download complete!")
	}

	_, err = os.Stat(typesDir)
	if os.IsNotExist(err) {
		log.Println("INFO: Types folder not found. Creating...")
		os.MkdirAll(typesDir, 0755)
		log.Println("INFO: Types folder created at", typesDir)

		log.Println("INFO: Downloading Pokemon type data...")

		for i := 1; i <= 19; i++ {
			resp, err := http.Get(fmt.Sprint(apiTypeUrl, "/", i))
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			typeFile := fmt.Sprint(typesDir, "/", fastjson.GetString(body, "name"), ".json")

			err = os.WriteFile(typeFile, body, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Println("INFO: Download complete!")
	}
}

func normalizeName(rawName string) (string, string) {
	prefixes := []string{"mega", "alolan", "galarian", "hisuian", "paldean"}
	suffixes := []string{"mega", "alola", "galar", "hisui", "paldea"}
	nameParts := strings.Split(rawName, " ")

	var apiPokemonName string
	var prettyPokemonName string

	for _, part := range nameParts {
		prettyPokemonName += strings.Title(part) + " "
	}
	prettyPokemonName = strings.Trim(prettyPokemonName, " ")

	// Create apiPokemonName
	if len(nameParts) > 1 && slices.Contains(prefixes, nameParts[0]) {
		apiPokemonName = fmt.Sprintf("%s-%s", nameParts[1], suffixes[slices.Index(prefixes, nameParts[0])])
	} else {
		apiPokemonName = strings.Trim(strings.ToLower(nameParts[0]), " ")

		if len(nameParts) > 1 {
			fmt.Println(prettyPokemonName)
			apiPokemonName = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(prettyPokemonName, ". ", "-"), " ", "-"))
		}

		apiPokemonName = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(apiPokemonName, ". ", "-"), " ", "-"))
	}

	return apiPokemonName, prettyPokemonName
}

func isValidPokemonName(name string) bool {
	allNames, err := os.ReadFile(pokemonNamesFile)
	if err != nil {
		log.Fatal(err)
	}

	allNamesArray := strings.Split(strings.TrimSpace(string(allNames)), "\n")
	pokemonMap := make(map[string]bool)
	for _, p := range allNamesArray {
		pokemonMap[p] = true
	}

	return pokemonMap[name]
}

func isNotCached(apiPokemonName string) bool {
	cacheFile := fmt.Sprint(cacheDir, "/", apiPokemonName, ".json")
	_, err := os.Stat(cacheFile)

	return os.IsNotExist(err)
}

func pullPokedexInfo(apiPokemonName string) {
	cacheFile := fmt.Sprint(cacheDir, "/", apiPokemonName, ".json")
	pokemonUrl := fmt.Sprint(apiPokemonUrl, "/", apiPokemonName)

	resp, err := http.Get(pokemonUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	err = os.WriteFile(cacheFile, body, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func showPokedexInfo(apiPokemonName string, prettyPokemonName string) {

	// Display:
	// Name (dex#)
	// Types
	// Weaknesses
	// Resistances
	// Abilities (indicate HA)
	// Sprite links
	// Cry link

	var dexNum string
	var types []string
	var weaknesses []string
	var resistances []string
	var abilities []string
	var spriteLinks []string
	var cryLink string

	cacheFile := fmt.Sprintf("%s/%s.json", cacheDir, apiPokemonName)
	pokemonDataJson, err := os.ReadFile(cacheFile)
	if err != nil {
		log.Fatal(err)
	}

	dexNum = strings.Split(fastjson.GetString(pokemonDataJson, "species", "url"), "/")[6]

	types = append(types, strings.Title(fastjson.GetString(pokemonDataJson, "types", "0", "type", "name")))
	types = append(types, strings.Title(fastjson.GetString(pokemonDataJson, "types", "1", "type", "name")))
	typesString := types[0]
	if types[1] != "" {
		typesString += fmt.Sprintf("/%s", types[1])
		weaknesses, resistances = getMultiTypeWeaknessesAndResistances(types[0], types[1], typesDir)
	} else {
		weaknesses, resistances = getSingleTypeWeaknessesAndResistances(types[0], typesDir)
	}

	var weaknessesString string
	for i := 0; i < len(weaknesses); i += 2 {
		weaknessesString += fmt.Sprintf("  - %s (x%s)\n", weaknesses[i], weaknesses[i+1])
	}

	var resistancesString string
	for i := 0; i < len(resistances); i += 2 {
		resistancesString += fmt.Sprintf("  - %s (x%s)\n", resistances[i], resistances[i+1])
	}

	// Compile formatted list of abilities, indicating which ones, if any, are hidden
	var abilitiesString string
	for i := range 4 {
		nextAbility := strings.ReplaceAll(strings.Title(fastjson.GetString(pokemonDataJson, "abilities", strconv.Itoa(i), "ability", "name")), "-", " ")
		if nextAbility == "" {
			break
		}

		var hiddenSuffix string
		if fastjson.GetBool(pokemonDataJson, "abilities", strconv.Itoa(i), "is_hidden") {
			hiddenSuffix = "(Hidden)"
		}
		abilities = append(abilities, fmt.Sprintf("%s %s", nextAbility, hiddenSuffix))
		abilitiesString += fmt.Sprintf("  - %s\n", abilities[i])
	}

	spriteLinks = append(spriteLinks, fastjson.GetString(pokemonDataJson, "sprites", "other", "official-artwork", "front_default"))
	spriteLinks = append(spriteLinks, fastjson.GetString(pokemonDataJson, "sprites", "other", "official-artwork", "front_shiny"))

	cryLink = fastjson.GetString(pokemonDataJson, "cries", "latest")

	fmt.Printf("%s (#%s) - %s type\n", prettyPokemonName, dexNum, typesString)
	fmt.Println("--------------------------")
	fmt.Printf("Weaknesses:\n%s", weaknessesString)
	fmt.Println("--------------------------")
	fmt.Printf("Resistances:\n%s", resistancesString)
	fmt.Println("--------------------------")
	fmt.Printf("Abilities:\n%s", abilitiesString)
	fmt.Print("--------------------------\n")
	fmt.Printf("Normal sprite: %s\n", spriteLinks[0])
	fmt.Printf("Shiny sprite: %s\n", spriteLinks[1])
	fmt.Println("--------------------------")
	fmt.Printf("Cry audio: %s\n", cryLink)
}

func main() {

	initialize()

	args := os.Args[1:] // just the user-provided args
	if len(args) < 1 {
		fmt.Println("ERROR: Incomplete command.")
		fmt.Println("Usage: pokedex [pokemon_name]")
		fmt.Println("For megas, use the format 'pokemon-mega'")
		fmt.Println("For regional forms, use the format 'pokemon-alola' or 'pokemon-galar'")
		fmt.Println("Replace any spaces or other punctuation with '-' (e.g. 'mr-mime', 'nidoran-m')")
		fmt.Println("Find a full list of pokemon names in .pokedexGo/cache/pokemonNames.txt")
		fmt.Println("Exiting.")
		os.Exit(0)
	}

	rawName := strings.Join(args, " ")
	apiPokemonName, prettyPokemonName := normalizeName(rawName)

	if !isValidPokemonName(apiPokemonName) {
		log.Fatalf("Invalid Pokemon name [%s]! Find a full list of pokemon names in .pokedexGo/cache/pokemonNames.txt", apiPokemonName)
	}

	if isNotCached(apiPokemonName) {
		log.Println("Pulling pokedex info for", apiPokemonName)
		pullPokedexInfo(apiPokemonName)
	}

	showPokedexInfo(apiPokemonName, prettyPokemonName)

}
