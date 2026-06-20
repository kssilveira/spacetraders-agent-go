package agent

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

type Agent struct {
	Token  string
	Client *http.Client
}

func (a Agent) All() error {
	headquarters, err := a.getHeadquarters()
	if err != nil {
		return err
	}
	contractTradeSymbols, err := a.acceptContract()
	if err != nil {
		return err
	}
	ship, err := a.buyShip(headquarters)
	if err != nil {
		return err
	}
	if err := a.navigate(headquarters, ship); err != nil {
		return err
	}
	if err := a.extract(ship, contractTradeSymbols); err != nil {
		return err
	}
	return nil
}

func (a Agent) getHeadquarters() (string, error) {
	agent, err := a.myAgent()
	if err != nil {
		return "", err
	}
	headquarters := agent.Headquarters
	waypoint, err := a.waypoint(headquarters)
	if err != nil {
		return "", err
	}
	fmt.Printf("%#v\n", waypoint)
	return headquarters, nil
}

func (a Agent) acceptContract() (map[string]any, error) {
	contracts, err := a.myContracts()
	if err != nil {
		return nil, err
	}
	contract := contracts[0]
	if !contract.Accepted {
		accepted, err := a.accept(contract.ID)
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

func (a Agent) buyShip(headquarters string) (string, error) {
	symbol, err := a.excavator()
	if err != nil {
		return "", err
	}
	if symbol != "" {
		return symbol, nil
	}
	return a.doBuyShip(headquarters)
}

func (a Agent) excavator() (string, error) {
	ships, err := a.myShips()
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

func (a Agent) doBuyShip(headquarters string) (string, error) {
	shipyardWaypoints, err := a.waypointsWithFilter(headquarters, "traits=SHIPYARD")
	if err != nil {
		return "", err
	}
	for _, shipyardWaypoint := range shipyardWaypoints {
		shipyard, err := a.waypointShipyard(shipyardWaypoint.Symbol)
		if err != nil {
			return "", err
		}
		for _, ship := range shipyard.Ships {
			if ship.Type != "SHIP_MINING_DRONE" {
				continue
			}
			got, err := a.myShipsBuy(shipyardWaypoint.Symbol, "SHIP_MINING_DRONE")
			if err != nil {
				return "", err
			}
			return got.Symbol, nil
		}
	}
	return "", fmt.Errorf("failed to buy ship")
}

func (a Agent) navigate(headquarters, ship string) error {
	orbit, err := a.myShipsOrbit(ship)
	if err != nil {
		return err
	}
	asteroidWaypoints, err := a.waypointsWithFilter(headquarters, "type=ENGINEERED_ASTEROID")
	if err != nil {
		return err
	}
	asteroid := asteroidWaypoints[0].Symbol
	if orbit.Nav.WaypointSymbol != asteroid {
		navigate, err := a.myShipsNavigate(ship, asteroid)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", navigate)
		dock, err := a.myShipsDock(ship)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", dock)
		refuel, err := a.myShipsRefuel(ship)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", refuel)
	}
	market, err := a.waypointMarket(orbit.Nav.WaypointSymbol)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", market)
	return nil
}

func (a Agent) extract(ship string, contractTradeSymbols map[string]any) error {
	isOrbit := false
	var err error
	for {
		isOrbit, err = a.sell(ship, contractTradeSymbols, isOrbit)
		if err != nil {
			return err
		}
		isDone, err := a.isDone(ship)
		if err != nil {
			return err
		}
		if isDone {
			return nil
		}
		if !isOrbit {
			isOrbit = true
			orbit, err := a.myShipsOrbit(ship)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", orbit)
		}
		extract, err := a.myShipsExtract(ship)
		if err != nil {
			return err
		}
		fmt.Printf("sleep %d\n", extract.Error.Data.Cooldown.RemainingSeconds)
		time.Sleep(time.Duration(extract.Error.Data.Cooldown.RemainingSeconds) * time.Second)
	}
	return nil
}

func (a Agent) sell(ship string, contractTradeSymbols map[string]any, isOrbit bool) (bool, error) {
	cargo, err := a.myShipsCargo(ship)
	if err != nil {
		return false, err
	}
	for _, item := range cargo.Inventory {
		if _, ok := contractTradeSymbols[item.Symbol]; ok {
			continue
		}
		if isOrbit {
			isOrbit = false
			dock, err := a.myShipsDock(ship)
			if err != nil {
				return false, err
			}
			fmt.Printf("%#v\n", dock)
		}
		sell, err := a.myShipsSell(ship, item.Symbol, item.Units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%#v\n", sell)
		jettison, err := a.myShipsJettison(ship, item.Symbol, item.Units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%#v\n", jettison)
	}
	return isOrbit, nil
}

func (a Agent) isDone(ship string) (bool, error) {
	cargo, err := a.myShipsCargo(ship)
	if err != nil {
		return false, err
	}
	if cargo.Units == cargo.Capacity && len(cargo.Inventory) == 1 {
		return true, nil
	}
	return false, nil
}

type MyAgentData struct {
	Headquarters string `json:"headquarters"`
}

type MyAgent struct {
	Data MyAgentData `json:"data"`
}

func (a Agent) myAgent() (MyAgentData, error) {
	var myAgent MyAgent
	if err := a.do("my/agent", "GET", nil, nil, &myAgent); err != nil {
		return MyAgentData{}, err
	}
	return myAgent.Data, nil
}

type Waypoint struct {
	Symbol string `json:"symbol"`
}

func (a Agent) waypoint(waypoint string) (Waypoint, error) {
	var res Waypoint
	if err := a.do("systems/{{.system}}/waypoints/{{.waypoint}}", "GET", map[string]string{
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

func (a Agent) waypointsWithFilter(waypoint, filter string) ([]Waypoint, error) {
	var waypoints Waypoints
	if err := a.do("systems/{{.system}}/waypoints?{{.filter}}", "GET", map[string]string{
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

func (a Agent) waypointShipyard(waypoint string) (Shipyard, error) {
	var waypointShipyard WaypointShipyard
	if err := a.do("systems/{{.system}}/waypoints/{{.waypoint}}/shipyard", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}, nil, &waypointShipyard); err != nil {
		return Shipyard{}, err
	}
	return waypointShipyard.Data, nil
}

type WaypointMarket struct {
}

func (a Agent) waypointMarket(waypoint string) (WaypointMarket, error) {
	var waypointMarket WaypointMarket
	if err := a.do("systems/{{.system}}/waypoints/{{.waypoint}}/market", "GET", map[string]string{
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

func (a Agent) myContracts() ([]Contract, error) {
	var myContracts MyContracts
	fmt.Printf("%#v\n", myContracts)
	if err := a.do("my/contracts", "GET", nil, nil, &myContracts); err != nil {
		return nil, err
	}
	return myContracts.Data, nil
}

type Accept struct {
}

func (a Agent) accept(id string) (Accept, error) {
	var accept Accept
	if err := a.do("my/contracts/{{.id}}/accept", "POST", map[string]string{
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

func (a Agent) myShips() ([]Ship, error) {
	var myShips MyShips
	if err := a.do("my/ships", "GET", nil, nil, &myShips); err != nil {
		return nil, err
	}
	return myShips.Data, nil
}

type MyShipsBuy struct {
	Data Ship `json:"data"`
}

func (a Agent) myShipsBuy(waypoint, shipType string) (Ship, error) {
	var myShipsBuy MyShipsBuy
	if err := a.do("my/ships", "POST", nil, map[string]any{
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

func (a Agent) myShipsOrbit(ship string) (ShipOrbit, error) {
	var myShipsOrbit MyShipsOrbit
	if err := a.do("my/ships/{{.ship}}/orbit", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsOrbit); err != nil {
		return ShipOrbit{}, err
	}
	return myShipsOrbit.Data, nil
}

type MyShipsDock struct {
}

func (a Agent) myShipsDock(ship string) (MyShipsDock, error) {
	var myShipsDock MyShipsDock
	if err := a.do("my/ships/{{.ship}}/dock", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsDock); err != nil {
		return MyShipsDock{}, err
	}
	return myShipsDock, nil
}

type MyShipsRefuel struct {
}

func (a Agent) myShipsRefuel(ship string) (MyShipsRefuel, error) {
	var myShipsRefuel MyShipsRefuel
	if err := a.do("my/ships/{{.ship}}/refuel", "POST", map[string]string{
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

func (a Agent) myShipsExtract(ship string) (MyShipsExtract, error) {
	var myShipsExtract MyShipsExtract
	if err := a.do("my/ships/{{.ship}}/extract", "POST", map[string]string{
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

func (a Agent) myShipsCargo(ship string) (Cargo, error) {
	var myShipsCargo MyShipsCargo
	if err := a.do("my/ships/{{.ship}}/cargo", "GET", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsCargo); err != nil {
		return Cargo{}, err
	}
	return myShipsCargo.Data, nil
}

type MyShipsNavigate struct {
}

func (a Agent) myShipsNavigate(ship, symbol string) (MyShipsNavigate, error) {
	var myShipsNavigate MyShipsNavigate
	if err := a.do("my/ships/{{.ship}}/navigate", "POST", map[string]string{
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

func (a Agent) myShipsSell(ship, symbol string, units int) (MyShipsSell, error) {
	var myShipsSell MyShipsSell
	if err := a.do("my/ships/{{.ship}}/sell", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"symbol": symbol,
		"units":  units,
	}, &myShipsSell); err != nil {
		return MyShipsSell{}, err
	}
	return myShipsSell, nil
}

type MyShipsJettison struct {
}

func (a Agent) myShipsJettison(ship, symbol string, units int) (MyShipsJettison, error) {
	var myShipsJettison MyShipsJettison
	if err := a.do("my/ships/{{.ship}}/jettison", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"symbol": symbol,
		"units":  units,
	}, &myShipsJettison); err != nil {
		return MyShipsJettison{}, err
	}
	return myShipsJettison, nil
}

const (
	baseUrl = "https://api.spacetraders.io/v2"
)

func (a Agent) do(pathTemplate, method string, templateData map[string]string, payloadData map[string]any, v any) error {
	parsedTemplate, err := template.New("pathTemplate").Parse(pathTemplate)
	if err != nil {
		return err
	}
	var builder strings.Builder
	if err = parsedTemplate.Execute(&builder, templateData); err != nil {
		return err
	}
	path := builder.String()
	url := fmt.Sprintf("%s/%s", baseUrl, path)
	payload, err := json.Marshal(payloadData)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+a.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	time.Sleep(time.Second / 2)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf(" => %d %s\n", resp.StatusCode, path)
	fmt.Println(prettyJSON.String())
	fmt.Printf(" => %d %s\n", resp.StatusCode, path)
	var res map[string]any
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return err
		}
	}
	return nil
}
