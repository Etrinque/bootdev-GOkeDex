package menu

import (
	"encoding/json"
	"errors"
	"fmt"
	pokecache "goProjects/gokedex/pokeCache"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// CLI ContentsMenu
type CliCmd struct {
	KeyMap   string
	Name     string
	Desc     string
	CallBack func(arg ...string) error
}

var CmdMap = make(map[string]CliCmd)
var PokeCache *pokecache.Cache

var (
	CurrentUrl = "http://pokeapi.co/api/v2/location-area"
	NextPage   string
	PrevPage   string
)

func init() {
	CmdMap["menu"] = CliCmd{"m", "Menu", "Display menu.", DisplayMenu}
	CmdMap["help"] = CliCmd{"h", "Help", "Display help", HelpFunc}
	CmdMap["mapLoc"] = CliCmd{"l", "Locations", "Map Results = 20", MapLocWrap}
	CmdMap["quit"] = CliCmd{"q", "Quit", "Exit Program", ExitFunc}
	CmdMap["fwd"] = CliCmd{"n", "Fwd", "Next", MapNext}
	CmdMap["back"] = CliCmd{"p", "Back", "Prev", MapPrev}
	CmdMap["explore"] = CliCmd{"e", "Explore", "Exploring Location", MapExplore}
	CmdMap["catch"] = CliCmd{"c", "Catch", "Catch come Pokemon", GoCatch}
	CmdMap["return"] = CliCmd{"r", "Return", "Return", ReturnFunc}
	CmdMap["details"] = CliCmd{"d", "Details", "View Pokémon details", ViewPokeDex}
}

func DisplayMenu(arg ...string) error {
	menu := []string{
		".=======Options======.",
		"|M ->: Menu..........|",
		"|H ->: Help..........|",
		"|L ->: View Locations|",
		"|Q ->: Exit Program..|",
		".====================.",
	}
	for i := range menu {
		fmt.Println(menu[i])
	}
	return nil
}

// Help Menu
func HelpFunc(arg ...string) error {
	return nil
}

// Return to Main Menu
func ReturnFunc(arg ...string) error {
	DisplayMenu()
	return nil
}

// Exit the Program
func ExitFunc(arg ...string) error {
	os.Exit(0)
	return nil
}

// struct for unmarshalling Json to Go
type JtoG struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"results"`
}

// Call to Cache init
func Setup(cache *pokecache.Cache) {
	PokeCache = cache
}

// Display and traversal of api-Locations
func GetData(CurrentUrl string) ([]byte, error) {
	if data, exists := PokeCache.Get(CurrentUrl); exists {
		resp := data
		return resp, nil
	}

	resp, err := http.Get(CurrentUrl)

	if err != nil {
		return nil, fmt.Errorf("error retrieving from source: %v", err)

	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error retrieving from source: %v", err)

	}
	PokeCache.Set(CurrentUrl, body)
	return body, nil
}

// retrieve location id
func GetID(url string) (string, error) {
	split := strings.Split(url, "/")

	if len(split) >= 2 {
		id := split[len(split)-2]
		return id, nil
	}
	return "", fmt.Errorf("invalid URL")
}

// Wrapper for continuity
func MapLocWrap(arg ...string) error {
	body, err := GetData(CurrentUrl)
	if err != nil {
		return fmt.Errorf("zom:100: %v", err)
	}
	return GoParse(body)
}

// Parse Api calls
func GoParse(body []byte) error {
	var result JtoG
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error Marshalling Data")
	}

	for _, location := range result.Results {
		id, err := GetID(location.Url)
		if err != nil {
			continue
		}
		fmt.Printf("ID: %s, %s\n", id, location.Name)
	}

	NextPage = result.Next
	PrevPage = result.Previous
	//commands for more results
	fmt.Println(
		".=======================.\n",
		"|N -> NEXT 20 results   |\n",
		"|P -> PREV 20 results   |\n",
		"|E -> Explore a location|\n",
		"|R -> Return to Menu    |\n",
		"|Q -> Quit Program      |\n",
		".=======================.",
	)
	return nil
}

// Next Api location results, -->20items/call
func MapNext(arg ...string) error {
	if NextPage != "" {
		CurrentUrl = NextPage
		body, err := GetData(CurrentUrl)
		if err != nil {
			err = errors.New("end of results")
			fmt.Println(err)
		}
		return GoParse(body)
	}
	return nil
}

// previous Api location results, <--20items/call
func MapPrev(arg ...string) error {
	if PrevPage != "" {
		CurrentUrl = PrevPage
		body, err := GetData(CurrentUrl)
		if err != nil {
			err = errors.New("previous Results not Available")
			fmt.Println(err)
		}
		return GoParse(body)
	}
	return nil
}

type AreaInfo struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

// Explore via Location-Area endpoint
func MapExplore(arg ...string) error {

	if len(arg) < 1 {
		return fmt.Errorf("must specify Location ID")
	}
	if len(arg) > 3 {
		return fmt.Errorf("id must be 3 or less characters")
	}

	id := arg[0]

	url := "http://pokeapi.co/api/v2/location-area/" + id

	//Call
	data, err := GetData(url)
	if err != nil {
		return err //fmt.Errorf("response not collected")
	}

	fmt.Println(".======================.")
	fmt.Println("|Looking for pokemon...|")
	fmt.Println(".======================.")

	//Parse
	var ai AreaInfo
	if err := json.Unmarshal(data, &ai); err != nil {
		return fmt.Errorf("error parsing data: %v", err)
	}

	for _, item := range ai.PokemonEncounters {
		name := item.Pokemon.Name
		url := item.Pokemon.URL
		pokemonID, err := GetID(url)
		if err != nil {
			fmt.Println("Cannot locate Pokemon")
		}
		fmt.Printf("ID: %s Name: %s\n", pokemonID, name)
	}

	fmt.Println(".==========================.")
	fmt.Println("|G -> Get Pokemon log by ID|")
	fmt.Println("|C -> Catch some Pokemon   |")
	fmt.Println(".==========================.")
	return nil

}

// Data for PokeDex
type PokeLog struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Abilities      []struct {
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
		Ability  struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
	} `json:"abilities"`
	Forms []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	Moves []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt int `json:"level_learned_at"`
			VersionGroup   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Species struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
}

// Get PokeLog data from Api
func GetPokeData(arg ...string) ([]byte, error) {
	var log PokeLog
	pokemonName := arg[1]
	pokeurl := "http://pokeapi.co/api/v2/pokemon/" + pokemonName
	body, _ := GetData(pokeurl)

	if err := json.Unmarshal(body, &log); err != nil {
		fmt.Println("Please enter name Exactly...")
		return nil, fmt.Errorf("failed to retrieve pokemon: %v", err)
	}

	fmt.Print(
		"=========LOADING=========\n",
		"Name:  %s\n",
		"ID:  %d\n",
		"BaseXP:  %d\n",
		"Species:  %s\n",
		"=========END LOG=========",
		log.Name, log.ID, log.BaseExperience, log.Species,
	)
	return body, nil
}

func ViewPokeDex(arg ...string) error {

	if len(arg) < 1 {
		fmt.Println("Specify the Pokemon Name")
		fmt.Println("PokeDex:")
		for name := range Pokedex {
			fmt.Printf("Name: %s", name)
		}
		return nil
	}

	name := arg[0]

	if pokemon, exists := Pokedex[name]; exists {
		fmt.Print(
			"=================POKEDEX==================\n",
			"ID: %s\n",
			"Name: %s\n",
			"Height: %s\n",
			"Weight: %s\n",
			"Species: %s\n",
			"XP: %s\n",
			"=================END--LOG=================\n",
			pokemon.ID, pokemon.Name, pokemon.Height, pokemon.Weight, pokemon.Species.Name, pokemon.BaseExperience,
		)
		return nil
	}

	fmt.Println("Not in PokeDex")
	return nil
}

var Pokedex = make(map[string]PokeLog)

// Catch Pokemon
func GoCatch(arg ...string) error {

	if len(arg) < 1 {
		fmt.Println(
			".===========================.",
			"|Must specify Pokemon ID No.|",
			".===========================.",
		)
	}

	pokemonName := arg[0]
	pokeurl := "http://pokeapi.co/api/v2/pokemon/" + pokemonName
	body, err := GetData(pokeurl)
	if err != nil {
		return err
	}

	var log PokeLog
	if err := json.Unmarshal(body, &log); err != nil {
		return fmt.Errorf("error retrieving log: %v", err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(255) > log.BaseExperience {
		fmt.Printf("%s was caught!\n", log.Name)
		Pokedex[log.Name] = log
	} else {
		fmt.Printf("%s escaped!\n", log.Name)
	}
	return nil
}
