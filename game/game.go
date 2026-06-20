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
		shipyard, err := g.waypointShipyard(shipyardWaypoint.Symbol)
		if err != nil {
			return "", err
		}
		for _, ship := range shipyard.Ships {
			if ship.Type != "SHIP_MINING_DRONE" {
				continue
			}
			got, err := g.myShipsBuy(shipyardWaypoint.Symbol, "SHIP_MINING_DRONE")
			if err != nil {
				return "", err
			}
			return got.Symbol, nil
		}
	}
	return "", fmt.Errorf("failed to buy ship")
}

func (g Game) navigate(headquarters, ship string) error {
	orbit, err := g.myShipsOrbit(ship)
	if err != nil {
		return err
	}
	asteroidWaypoints, err := g.waypointsWithFilter(headquarters, "type=ENGINEERED_ASTEROID")
	if err != nil {
		return err
	}
	asteroid := asteroidWaypoints[0].Symbol
	if orbit.Nav.WaypointSymbol != asteroid {
		navigate, err := g.myShipsNavigate(ship, asteroid)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", navigate)
		dock, err := g.myShipsDock(ship)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", dock)
		refuel, err := g.myShipsRefuel(ship)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", refuel)
	}
	market, err := g.waypointMarket(orbit.Nav.WaypointSymbol)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", market)
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
			orbit, err := g.myShipsOrbit(ship)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", orbit)
		}
		extract, err := g.myShipsExtract(ship)
		if err != nil {
			return err
		}
		fmt.Printf("sleep %d\n", extract.Error.Data.Cooldown.RemainingSeconds)
		time.Sleep(time.Duration(extract.Error.Data.Cooldown.RemainingSeconds) * time.Second)
	}
	return nil
}

func (g Game) sell(ship string, contractTradeSymbols map[string]any, isOrbit bool) (bool, error) {
	cargo, err := g.myShipsCargo(ship)
	if err != nil {
		return false, err
	}
	for _, item := range cargo.Inventory {
		if _, ok := contractTradeSymbols[item.Symbol]; ok {
			continue
		}
		if isOrbit {
			isOrbit = false
			dock, err := g.myShipsDock(ship)
			if err != nil {
				return false, err
			}
			fmt.Printf("%#v\n", dock)
		}
		sell, err := g.myShipsSell(ship, item.Symbol, item.Units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%#v\n", sell)
		jettison, err := g.myShipsJettison(ship, item.Symbol, item.Units)
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

type ShipyardShip struct {
	Type string `json:"type"`
}

type Shipyard struct {
	Ships []ShipyardShip `json:"ships"`
}

type WaypointShipyard struct {
	Data Shipyard `json:"data"`
}

func (g Game) waypointShipyard(waypoint string) (Shipyard, error) {
	var waypointShipyard WaypointShipyard
	if _, err := g.do("systems/{{.system}}/waypoints/{{.waypoint}}/shipyard", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}, nil, &waypointShipyard); err != nil {
		return Shipyard{}, err
	}
	return waypointShipyard.Data, nil
}

type WaypointMarket struct {
}

func (g Game) waypointMarket(waypoint string) (WaypointMarket, error) {
	var waypointMarket WaypointMarket
	if _, err := g.do("systems/{{.system}}/waypoints/{{.waypoint}}/market", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}, nil, &waypointMarket); err != nil {
		return WaypointMarket{}, err
	}
	return waypointMarket, nil
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

type MyShipsBuy struct {
	Data Ship `json:"data"`
}

func (g Game) myShipsBuy(waypoint, shipType string) (Ship, error) {
	var myShipsBuy MyShipsBuy
	if _, err := g.do("my/ships", "POST", nil, map[string]any{
		"shipType":       shipType,
		"waypointSymbol": waypoint,
	}, &myShipsBuy); err != nil {
		return Ship{}, nil
	}
	return myShipsBuy.Data, nil
}

type Nav struct {
	WaypointSymbol string `json:"waypointSymbol"`
}

type ShipOrbit struct {
	Nav Nav `json:"nav"`
}

type MyShipsOrbit struct {
	Data ShipOrbit `json:"data"`
}

func (g Game) myShipsOrbit(ship string) (ShipOrbit, error) {
	var myShipsOrbit MyShipsOrbit
	if _, err := g.do("my/ships/{{.ship}}/orbit", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsOrbit); err != nil {
		return ShipOrbit{}, err
	}
	return myShipsOrbit.Data, nil
}

type MyShipsDock struct {
}

func (g Game) myShipsDock(ship string) (MyShipsDock, error) {
	var myShipsDock MyShipsDock
	if _, err := g.do("my/ships/{{.ship}}/dock", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsDock); err != nil {
		return MyShipsDock{}, err
	}
	return myShipsDock, nil
}

type MyShipsRefuel struct {
}

func (g Game) myShipsRefuel(ship string) (MyShipsRefuel, error) {
	var myShipsRefuel MyShipsRefuel
	if _, err := g.do("my/ships/{{.ship}}/refuel", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsRefuel); err != nil {
		return MyShipsRefuel{}, err
	}
	return myShipsRefuel, nil
}

type MyShipsExtractErrorCooldown struct {
	RemainingSeconds int `json:"remainingSeconds"`
}

type MyShipsExtractErrorData struct {
	Cooldown MyShipsExtractErrorCooldown `json:"cooldown"`
}

type MyShipsExtractError struct {
	Data MyShipsExtractErrorData `json:"data"`
}

type MyShipsExtract struct {
	Error MyShipsExtractError `json:"error"`
}

func (g Game) myShipsExtract(ship string) (MyShipsExtract, error) {
	var myShipsExtract MyShipsExtract
	if _, err := g.do("my/ships/{{.ship}}/extract", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsExtract); err != nil {
		return MyShipsExtract{}, err
	}
	return myShipsExtract, nil
}

type Inventory struct {
	Symbol string `json:"symbol"`
	Units  int    `json:"units"`
}

type Cargo struct {
	Capacity  int         `json:"capacity"`
	Units     int         `json:"units"`
	Inventory []Inventory `json:"inventory"`
}

type MyShipsCargo struct {
	Data Cargo `json:"data"`
}

func (g Game) myShipsCargo(ship string) (Cargo, error) {
	var myShipsCargo MyShipsCargo
	if _, err := g.do("my/ships/{{.ship}}/cargo", "GET", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsCargo); err != nil {
		return Cargo{}, err
	}
	return myShipsCargo.Data, nil
}

func (g Game) myShipsAction(ship, action, method string) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/{{.action}}", method, map[string]string{
		"ship":   ship,
		"action": action,
	}, map[string]any{}, nil)
}

type MyShipsNavigate struct {
}

func (g Game) myShipsNavigate(ship, symbol string) (MyShipsNavigate, error) {
	var myShipsNavigate MyShipsNavigate
	if _, err := g.do("my/ships/{{.ship}}/navigate", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"waypointSymbol": symbol,
	}, &myShipsNavigate); err != nil {
		return MyShipsNavigate{}, err
	}
	return myShipsNavigate, nil
}

type MyShipsSell struct {
}

func (g Game) myShipsSell(ship, symbol string, units int) (MyShipsSell, error) {
	var myShipsSell MyShipsSell
	if _, err := g.do("my/ships/{{.ship}}/sell", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"symbol": symbol,
		"units":  units,
	}, &myShipsSell); err != nil {
		return MyShipsSell{}, err
	}
	return myShipsSell, nil
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
