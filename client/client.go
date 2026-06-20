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
	if err := c.do("my/agent", "GET", nil, nil, &myAgent); err != nil {
		return MyAgentData{}, err
	}
	return myAgent.Data, nil
}

type Waypoint struct {
	Symbol string `json:"symbol"`
}

func (c Client) Waypoint(waypoint string) (Waypoint, error) {
	var res Waypoint
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}", "GET", map[string]string{
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

func (c Client) WaypointsWithFilter(waypoint, filter string) ([]Waypoint, error) {
	var waypoints Waypoints
	if err := c.do("systems/{{.system}}/waypoints?{{.filter}}", "GET", map[string]string{
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

func (c Client) WaypointShipyard(waypoint string) (Shipyard, error) {
	var waypointShipyard WaypointShipyard
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}/shipyard", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}, nil, &waypointShipyard); err != nil {
		return Shipyard{}, err
	}
	return waypointShipyard.Data, nil
}

type WaypointMarket struct {
}

func (c Client) WaypointMarket(waypoint string) (WaypointMarket, error) {
	var waypointMarket WaypointMarket
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}/market", "GET", map[string]string{
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

func (c Client) MyContracts() ([]Contract, error) {
	var myContracts MyContracts
	fmt.Printf("%#v\n", myContracts)
	if err := c.do("my/contracts", "GET", nil, nil, &myContracts); err != nil {
		return nil, err
	}
	return myContracts.Data, nil
}

type Accept struct {
}

func (c Client) Accept(id string) (Accept, error) {
	var accept Accept
	if err := c.do("my/contracts/{{.id}}/accept", "POST", map[string]string{
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

func (c Client) MyShips() ([]Ship, error) {
	var myShips MyShips
	if err := c.do("my/ships", "GET", nil, nil, &myShips); err != nil {
		return nil, err
	}
	return myShips.Data, nil
}

type MyShipsBuy struct {
	Data Ship `json:"data"`
}

func (c Client) MyShipsBuy(waypoint, shipType string) (Ship, error) {
	var myShipsBuy MyShipsBuy
	if err := c.do("my/ships", "POST", nil, map[string]any{
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

func (c Client) MyShipsOrbit(ship string) (ShipOrbit, error) {
	var myShipsOrbit MyShipsOrbit
	if err := c.do("my/ships/{{.ship}}/orbit", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsOrbit); err != nil {
		return ShipOrbit{}, err
	}
	return myShipsOrbit.Data, nil
}

type MyShipsDock struct {
}

func (c Client) MyShipsDock(ship string) (MyShipsDock, error) {
	var myShipsDock MyShipsDock
	if err := c.do("my/ships/{{.ship}}/dock", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsDock); err != nil {
		return MyShipsDock{}, err
	}
	return myShipsDock, nil
}

type MyShipsRefuel struct {
}

func (c Client) MyShipsRefuel(ship string) (MyShipsRefuel, error) {
	var myShipsRefuel MyShipsRefuel
	if err := c.do("my/ships/{{.ship}}/refuel", "POST", map[string]string{
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

func (c Client) MyShipsExtract(ship string) (MyShipsExtract, error) {
	var myShipsExtract MyShipsExtract
	if err := c.do("my/ships/{{.ship}}/extract", "POST", map[string]string{
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

func (c Client) MyShipsCargo(ship string) (Cargo, error) {
	var myShipsCargo MyShipsCargo
	if err := c.do("my/ships/{{.ship}}/cargo", "GET", map[string]string{
		"ship": ship,
	}, map[string]any{}, &myShipsCargo); err != nil {
		return Cargo{}, err
	}
	return myShipsCargo.Data, nil
}

type MyShipsNavigate struct {
}

func (c Client) MyShipsNavigate(ship, symbol string) (MyShipsNavigate, error) {
	var myShipsNavigate MyShipsNavigate
	if err := c.do("my/ships/{{.ship}}/navigate", "POST", map[string]string{
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

func (c Client) MyShipsSell(ship, symbol string, units int) (MyShipsSell, error) {
	var myShipsSell MyShipsSell
	if err := c.do("my/ships/{{.ship}}/sell", "POST", map[string]string{
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

func (c Client) MyShipsJettison(ship, symbol string, units int) (MyShipsJettison, error) {
	var myShipsJettison MyShipsJettison
	if err := c.do("my/ships/{{.ship}}/jettison", "POST", map[string]string{
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

func (c Client) do(pathTemplate, method string, templateData map[string]string, payloadData map[string]any, v any) error {
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
