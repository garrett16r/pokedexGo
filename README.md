[![Auto Build and Release](https://github.com/garrett16r/pokedexGo/actions/workflows/autoBuildRelease.yml/badge.svg)](https://github.com/garrett16r/pokedexGo/actions/workflows/autoBuildRelease.yml)

# pokedexGo
Look up Pokedex info through CLI. A rewrite of my [pokedexPy](https://github.com/garrett16r/pokedexPy) project in Go.

Developed as a way for me to learn and practice with Go, REST API, caching, and JSON parsing using information I'm already very familiar with.

Data is pulled from the amazing [PokeAPI](https://pokeapi.co) project.

# Setup
1. Download the latest [release](https://github.com/garrett16r/pokedexGo/releases) for your OS/architechure
2. Run the executable from the command line
3. The only argument the program takes is the name of a Pokemon (./pokedexGo [pokemon])

# Notes
- On first run, two folders will be created in the same directory as pokedexGo
- `cache/` stores previously pulled Pokemon data and a list of all valid Pokemon names
- `types/` stores one .json file per type, which will be used for calculating type weaknesses and resistances
- pokemonNames.txt and type JSON files will also be downloaded from PokeAPI at this time

- Regional forms and mega evolutions are supported. Using Alolan Ninetales as an example, "Alolan Ninetales" and "ninetales-alola" will both be accepted. The same formats are accepted for megas.
