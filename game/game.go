package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type Game struct {
	Token  string
	Client *http.Client
}

func (g Game) All() error {
	headquarters, err := g.getHeadquarters()
	if err != nil {
		return err
	}
	contractTradeSymbols, err := g.acceptContract()
	if err != nil {
		return err
	}
	ship, err := g.buyShip(headquarters)
	if err != nil {
		return err
	}
	if err := g.navigate(headquarters, ship); err != nil {
		return err
	}
	if err := g.extract(ship, contractTradeSymbols); err != nil {
		return err
	}
	return nil
}

func (g Game) getHeadquarters() (string, error) {
	agent, err := g.myAgent()
	if err != nil {
		return "", err
	}
	headquarters := agent.Headquarters
	waypoint, err := g.waypoint(headquarters)
	if err != nil {
		return "", err
	}
	fmt.Printf("%#v\n", waypoint)
	return headquarters, nil
}

func (g Game) acceptContract() (map[string]any, error) {
	contracts, err := g.myContracts()
	if err != nil {
		return nil, err
	}
	contract := contracts[0]
	if !contract.Accepted {
		accepted, err := g.accept(contract.ID)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%#v\n", accepted)
	}
	res := map[string]any{}
	for _, deliver := range contract.Terms.Deliver {
		res[deliver.TradeSymbol] = nil
	}
	return res, nil
}

func (g Game) buyShip(headquarters string) (string, error) {
	symbol, err := g.excavator()
	if err != nil {
		return "", err
	}
	if symbol != "" {
		return symbol, nil
	}
	return g.doBuyShip(headquarters)
}

func (g Game) excavator() (string, error) {
	ships, err := g.myShips()
	if err != nil {
		return "", err
	}
	for _, ship := range ships {
		if ship.Registration.Role != "EXCAVATOR" {
			continue
		}
		return ship.Symbol, nil
	}
	return "", nil
}

func (g Game) doBuyShip(headquarters string) (string, error) {
	shipyardWaypoints, err := g.waypointsWithFilter(headquarters, "traits=SHIPYARD")
	if err != nil {
		return "", err
	}
	for _, shipyardWaypoint := range shipyardWaypoints {
		shipyard, err := g.waypointAction(shipyardWaypoint.Symbol, "shipyard")
		if err != nil {
			return "", err
		}
		if _, ok := shipyard["ships"]; !ok {
			continue
		}
		ships, err := getListField(shipyard, "ships")
		if err != nil {
			return "", err
		}
		for i := range ships {
			ship, err := getMap(ships[i])
			if err != nil {
				return "", err
			}
			shipType, err := getStringField(ship, "type")
			if err != nil {
				return "", err
			}
			fmt.Printf("shipType %s\n", shipType)
			if shipType != "SHIP_MINING_DRONE" {
				continue
			}
			got, err := g.myShipsBuy(shipyardWaypoint.Symbol, "SHIP_MINING_DRONE")
			if err != nil {
				return "", err
			}
			symbol, err := getStringField(got, "symbol")
			if err != nil {
				return "", err
			}
			return symbol, nil
		}
	}
	return "", fmt.Errorf("failed to buy ship")
}

func (g Game) navigate(headquarters, ship string) error {
	orbit, err := g.myShipsAction(ship, "orbit", "POST")
	if err != nil {
		return err
	}
	nav, err := getMapField(orbit, "nav")
	if err != nil {
		return err
	}
	shipWaypoint, err := getStringField(nav, "waypointSymbol")
	if err != nil {
		return err
	}
	asteroidWaypoints, err := g.waypointsWithFilter(headquarters, "type=ENGINEERED_ASTEROID")
	if err != nil {
		return err
	}
	asteroid := asteroidWaypoints[0].Symbol
	if shipWaypoint != asteroid {
		navigate, err := g.myShipsNavigate(ship, asteroid)
		if err != nil {
			return err
		}
		fmt.Printf("%d\n", len(navigate))
		dock, err := g.myShipsAction(ship, "dock", "POST")
		if err != nil {
			return err
		}
		fmt.Printf("%d\n", len(dock))
		refuel, err := g.myShipsAction(ship, "refuel", "POST")
		if err != nil {
			return err
		}
		fmt.Printf("%d\n", len(refuel))
	}
	market, err := g.waypointAction(shipWaypoint, "market")
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", len(market))
	return nil
}

func (g Game) extract(ship string, contractTradeSymbols map[string]any) error {
	isOrbit := false
	var err error
	for {
		isOrbit, err = g.sell(ship, contractTradeSymbols, isOrbit)
		if err != nil {
			return err
		}
		isDone, err := g.isDone(ship)
		if err != nil {
			return err
		}
		if isDone {
			return nil
		}
		if !isOrbit {
			isOrbit = true
			orbit, err := g.myShipsAction(ship, "orbit", "POST")
			if err != nil {
				return err
			}
			fmt.Printf("%d\n", len(orbit))
		}
		extract, err := g.myShipsAction(ship, "extract", "POST")
		if err != nil {
			return err
		}
		if _, ok := extract["error"]; ok {
			errorMap, err := getMapField(extract, "error")
			if err != nil {
				return err
			}
			data, err := getMapField(errorMap, "data")
			if err != nil {
				return err
			}
			cooldown, err := getMapField(data, "cooldown")
			if err != nil {
				return err
			}
			remainingSeconds, err := getIntField(cooldown, "remainingSeconds")
			if err != nil {
				return err
			}
			fmt.Printf("sleep %d\n", remainingSeconds)
			time.Sleep(time.Duration(remainingSeconds) * time.Second)
		}
	}
	return nil
}

func (g Game) sell(ship string, contractTradeSymbols map[string]any, isOrbit bool) (bool, error) {
	cargo, err := g.myShipsAction(ship, "cargo", "GET")
	if err != nil {
		return false, err
	}
	inventory, err := getListField(cargo, "inventory")
	if err != nil {
		return false, err
	}
	for i := range inventory {
		item, err := getMap(inventory[i])
		if err != nil {
			return false, err
		}
		symbol, err := getStringField(item, "symbol")
		if err != nil {
			return false, err
		}
		if _, ok := contractTradeSymbols[symbol]; ok {
			continue
		}
		units, err := getIntField(item, "units")
		if err != nil {
			return false, err
		}
		if isOrbit {
			isOrbit = false
			dock, err := g.myShipsAction(ship, "dock", "POST")
			if err != nil {
				return false, err
			}
			fmt.Printf("%d\n", len(dock))
		}
		sell, err := g.myShipsSell(ship, symbol, units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%d\n", len(sell))
		jettison, err := g.myShipsJettison(ship, symbol, units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%d\n", len(jettison))
	}
	return isOrbit, nil
}

func (g Game) isDone(ship string) (bool, error) {
	cargo, err := g.myShipsAction(ship, "cargo", "GET")
	if err != nil {
		return false, err
	}
	inventory, err := getListField(cargo, "inventory")
	if err != nil {
		return false, err
	}
	capacity, err := getIntField(cargo, "capacity")
	if err != nil {
		return false, err
	}
	units, err := getIntField(cargo, "units")
	if err != nil {
		return false, err
	}
	if units == capacity && len(inventory) == 1 {
		return true, nil
	}
	return false, nil
}

type Agent struct {
	Headquarters string `json:"headquarters"`
}

type MyAgent struct {
	Data Agent `json:"data"`
}

func (g Game) myAgent() (Agent, error) {
	var myAgent MyAgent
	if _, err := g.do("my/agent", "GET", nil, nil, &myAgent); err != nil {
		return Agent{}, err
	}
	return myAgent.Data, nil
}

type Waypoint struct {
	Symbol string `json:"symbol"`
}

func (g Game) waypoint(waypoint string) (Waypoint, error) {
	var res Waypoint
	if _, err := g.do("systems/{{.system}}/waypoints/{{.waypoint}}", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}, nil, &res); err != nil {
		return Waypoint{}, err
	}
	return res, nil
}

type Waypoints struct {
	Data []Waypoint `json:"data"`
}

func (g Game) waypointsWithFilter(waypoint, filter string) ([]Waypoint, error) {
	var waypoints Waypoints
	if _, err := g.do("systems/{{.system}}/waypoints?{{.filter}}", "GET", map[string]string{
		"system": strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"filter": filter,
	}, nil, &waypoints); err != nil {
		return nil, err
	}
	return waypoints.Data, nil
}

func (g Game) waypointAction(waypoint, action string) (map[string]any, error) {
	return g.do("systems/{{.system}}/waypoints/{{.waypoint}}/{{.action}}", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
		"action":   action,
	}, nil, nil)
}

type Deliver struct {
	TradeSymbol string `json:"tradeSymbol"`
}

type Terms struct {
	Deliver []Deliver `json:"deliver"`
}

type Contract struct {
	ID       string `json:"id"`
	Accepted bool   `json:"accepted"`
	Terms    Terms  `json:"terms"`
}

type MyContracts struct {
	Data []Contract `json:"data"`
}

func (g Game) myContracts() ([]Contract, error) {
	var myContracts MyContracts
	fmt.Printf("%#v\n", myContracts)
	if _, err := g.do("my/contracts", "GET", nil, nil, &myContracts); err != nil {
		return nil, err
	}
	return myContracts.Data, nil
}

type Accept struct {
}

func (g Game) accept(id string) (Accept, error) {
	var accept Accept
	if _, err := g.do("my/contracts/{{.id}}/accept", "POST", map[string]string{
		"id": id,
	}, nil, &accept); err != nil {
		return Accept{}, err
	}
	return accept, nil
}

type Registration struct {
	Role string `json:"role"`
}

type Ship struct {
	Symbol       string       `json:"symbol"`
	Registration Registration `json:"registration"`
}

type MyShips struct {
	Data []Ship `json:"data"`
}

func (g Game) myShips() ([]Ship, error) {
	var myShips MyShips
	if _, err := g.do("my/ships", "GET", nil, nil, &myShips); err != nil {
		return nil, err
	}
	return myShips.Data, nil
}

func (g Game) myShipsBuy(waypoint, shipType string) (map[string]any, error) {
	return g.do("my/ships", "POST", nil, map[string]any{
		"shipType":       shipType,
		"waypointSymbol": waypoint,
	}, nil)
}

func (g Game) myShipsAction(ship, action, method string) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/{{.action}}", method, map[string]string{
		"ship":   ship,
		"action": action,
	}, map[string]any{}, nil)
}

func (g Game) myShipsNavigate(ship, symbol string) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/navigate", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"waypointSymbol": symbol,
	}, nil)
}

func (g Game) myShipsSell(ship, symbol string, units int) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/sell", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"symbol": symbol,
		"units":  units,
	}, nil)
}

func (g Game) myShipsJettison(ship, symbol string, units int) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/jettison", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"symbol": symbol,
		"units":  units,
	}, nil)
}

const (
	baseUrl = "https://api.spacetraders.io/v2"
)

func (g Game) do(pathTemplate, method string, templateData map[string]string, payloadData map[string]any, v any) (map[string]any, error) {
	parsedTemplate, err := template.New("pathTemplate").Parse(pathTemplate)
	if err != nil {
		return nil, err
	}
	var builder strings.Builder
	if err = parsedTemplate.Execute(&builder, templateData); err != nil {
		return nil, err
	}
	path := builder.String()
	url := fmt.Sprintf("%s/%s", baseUrl, path)
	payload, err := json.Marshal(payloadData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+g.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second / 2)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		return nil, err
	}
	fmt.Printf(" => %d %s\n", resp.StatusCode, path)
	fmt.Println(prettyJSON.String())
	fmt.Printf(" => %d %s\n", resp.StatusCode, path)
	var res map[string]any
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return nil, err
		}
	}
	data, err := getMapField(res, "data")
	if err != nil {
		return res, nil
	}
	return data, nil
}

func getStringField(data map[string]any, key string) (string, error) {
	res, ok := data[key].(string)
	if !ok {
		return "", fmt.Errorf("%#v missing string field %s", data, key)
	}
	return res, nil
}

func getIntField(data map[string]any, key string) (int, error) {
	res, ok := data[key].(float64)
	if !ok {
		return 0, fmt.Errorf("%#v missing int field %s", data, key)
	}
	return int(res), nil
}

func getBoolField(data map[string]any, key string) (bool, error) {
	res, ok := data[key].(bool)
	if !ok {
		return false, fmt.Errorf("%#v missing bool field %s", data, key)
	}
	return res, nil
}

func getMapField(data map[string]any, key string) (map[string]any, error) {
	res, ok := data[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%#v missing map field %s", data, key)
	}
	return res, nil
}

func getListField(data map[string]any, key string) ([]any, error) {
	res, ok := data[key].([]any)
	if !ok {
		return nil, fmt.Errorf("%#v missing list field %s", data, key)
	}
	return res, nil
}

func getMap(data any) (map[string]any, error) {
	res, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%#v not a map", data)
	}
	return res, nil
}
