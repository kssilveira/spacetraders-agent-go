package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Token  string
	Client *http.Client
}

type MyAgentData struct {
	Headquarters string `json:"headquarters"`
}

type MyAgent struct {
	Data MyAgentData `json:"data"`
}

func (c Client) MyAgent() (MyAgentData, error) {
	var myAgent MyAgent
	if err := c.do("my/agent", &myAgent, Do{}); err != nil {
		return MyAgentData{}, err
	}
	return myAgent.Data, nil
}

type Waypoint struct {
	Symbol string `json:"symbol"`
}

func (c Client) Waypoint(waypoint string) (Waypoint, error) {
	var res Waypoint
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}", &res, Do{Template: map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}}); err != nil {
		return Waypoint{}, err
	}
	return res, nil
}

type Waypoints struct {
	Data []Waypoint `json:"data"`
}

func (c Client) WaypointsWithFilter(waypoint, filter string) ([]Waypoint, error) {
	var waypoints Waypoints
	if err := c.do("systems/{{.system}}/waypoints?{{.filter}}", &waypoints, Do{Template: map[string]string{
		"system": strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"filter": filter,
	}}); err != nil {
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

func (c Client) WaypointShipyard(waypoint string) (Shipyard, error) {
	var waypointShipyard WaypointShipyard
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}/shipyard", &waypointShipyard, Do{Template: map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}}); err != nil {
		return Shipyard{}, err
	}
	return waypointShipyard.Data, nil
}

type WaypointMarket struct {
}

func (c Client) WaypointMarket(waypoint string) (WaypointMarket, error) {
	var waypointMarket WaypointMarket
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}/market", &waypointMarket, Do{Template: map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}}); err != nil {
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

func (c Client) MyContracts() ([]Contract, error) {
	var myContracts MyContracts
	fmt.Printf("%#v\n", myContracts)
	if err := c.do("my/contracts", &myContracts, Do{}); err != nil {
		return nil, err
	}
	return myContracts.Data, nil
}

type Accept struct {
}

func (c Client) Accept(id string) (Accept, error) {
	var accept Accept
	if err := c.do("my/contracts/{{.id}}/accept", &accept, Do{Method: "POST", Template: map[string]string{
		"id": id,
	}}); err != nil {
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

func (c Client) MyShips() ([]Ship, error) {
	var myShips MyShips
	if err := c.do("my/ships", &myShips, Do{}); err != nil {
		return nil, err
	}
	return myShips.Data, nil
}

type MyShipsBuy struct {
	Data Ship `json:"data"`
}

func (c Client) MyShipsBuy(waypoint, shipType string) (Ship, error) {
	var myShipsBuy MyShipsBuy
	if err := c.do("my/ships", &myShipsBuy, Do{Method: "POST", Payload: map[string]any{
		"shipType":       shipType,
		"waypointSymbol": waypoint,
	}}); err != nil {
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

func (c Client) MyShipsOrbit(ship string) (ShipOrbit, error) {
	var myShipsOrbit MyShipsOrbit
	if err := c.do("my/ships/{{.ship}}/orbit", &myShipsOrbit, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{}}); err != nil {
		return ShipOrbit{}, err
	}
	return myShipsOrbit.Data, nil
}

type MyShipsDock struct {
}

func (c Client) MyShipsDock(ship string) (MyShipsDock, error) {
	var myShipsDock MyShipsDock
	if err := c.do("my/ships/{{.ship}}/dock", &myShipsDock, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{}}); err != nil {
		return MyShipsDock{}, err
	}
	return myShipsDock, nil
}

type MyShipsRefuel struct {
}

func (c Client) MyShipsRefuel(ship string) (MyShipsRefuel, error) {
	var myShipsRefuel MyShipsRefuel
	if err := c.do("my/ships/{{.ship}}/refuel", &myShipsRefuel, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{}}); err != nil {
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

func (c Client) MyShipsExtract(ship string) (MyShipsExtract, error) {
	var myShipsExtract MyShipsExtract
	if err := c.do("my/ships/{{.ship}}/extract", &myShipsExtract, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{}}); err != nil {
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

func (c Client) MyShipsCargo(ship string) (Cargo, error) {
	var myShipsCargo MyShipsCargo
	if err := c.do("my/ships/{{.ship}}/cargo", &myShipsCargo, Do{Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{}}); err != nil {
		return Cargo{}, err
	}
	return myShipsCargo.Data, nil
}

type MyShipsNavigate struct {
}

func (c Client) MyShipsNavigate(ship, symbol string) (MyShipsNavigate, error) {
	var myShipsNavigate MyShipsNavigate
	if err := c.do("my/ships/{{.ship}}/navigate", &myShipsNavigate, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{
		"waypointSymbol": symbol,
	}}); err != nil {
		return MyShipsNavigate{}, err
	}
	return myShipsNavigate, nil
}

type MyShipsSell struct {
}

func (c Client) MyShipsSell(ship, symbol string, units int) (MyShipsSell, error) {
	var myShipsSell MyShipsSell
	if err := c.do("my/ships/{{.ship}}/sell", &myShipsSell, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{
		"symbol": symbol,
		"units":  units,
	}}); err != nil {
		return MyShipsSell{}, err
	}
	return myShipsSell, nil
}

type MyShipsJettison struct {
}

func (c Client) MyShipsJettison(ship, symbol string, units int) (MyShipsJettison, error) {
	var myShipsJettison MyShipsJettison
	if err := c.do("my/ships/{{.ship}}/jettison", &myShipsJettison, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{
		"symbol": symbol,
		"units":  units,
	}}); err != nil {
		return MyShipsJettison{}, err
	}
	return myShipsJettison, nil
}

const (
	baseUrl = "https://api.spacetraders.io/v2"
)

type Do struct {
	Method   string
	Template map[string]string
	Payload  map[string]any
}

func (c Client) do(pathTemplate string, value any, cfg Do) error {
	parsedTemplate, err := template.New("pathTemplate").Parse(pathTemplate)
	if err != nil {
		return err
	}
	var builder strings.Builder
	if err = parsedTemplate.Execute(&builder, cfg.Template); err != nil {
		return err
	}
	path := builder.String()
	url := fmt.Sprintf("%s/%s", baseUrl, path)
	payload, err := json.Marshal(cfg.Payload)
	if err != nil {
		return err
	}
	if cfg.Method == "" {
		cfg.Method = "GET"
	}
	req, err := http.NewRequest(cfg.Method, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
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
	return json.Unmarshal(body, value)
}
