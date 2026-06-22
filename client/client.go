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
	AccountToken string
	AgentToken   string
	Client       *http.Client
}

type Register struct {
	Token string `json:"token"`
}

type RegisterRes struct {
	Data Register `json:"data"`
}

func (c Client) Register(symbol, faction string) (Register, error) {
	var res RegisterRes
	if err := c.do("register", &res, Do{IsAccount: true, Method: "POST", Payload: map[string]any{
		"symbol":  symbol,
		"faction": faction,
	}}); err != nil {
		return Register{}, err
	}
	return res.Data, nil
}

type AgentError struct {
	Code int `json:"code"`
}

type Agent struct {
	Headquarters string `json:"headquarters"`
	Credits      int    `json:"credits"`
}

type AgentRes struct {
	Data  Agent      `json:"data"`
	Error AgentError `json:"error"`
}

func (c Client) Agent() (AgentRes, error) {
	var res AgentRes
	if err := c.do("my/agent", &res, Do{}); err != nil {
		return AgentRes{}, err
	}
	return res, nil
}

type Trait struct {
	Symbol string `json:"symbol"`
}

type Waypoint struct {
	Symbol         string `json:"symbol"`
	Type           string `json:"type"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	Distance       int
	Traits         []Trait `json:"traits"`
	HasMarketplace bool
	Exports        []Item
	Imports        []Item
	Exchange       []Item
}

type WaypointRes struct {
	Data Waypoint `json:"data"`
}

func (c Client) Waypoint(waypoint string) (Waypoint, error) {
	var res WaypointRes
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}", &res, Do{Template: map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}}); err != nil {
		return Waypoint{}, err
	}
	return res.Data, nil
}

type WaypointsRes struct {
	Data []Waypoint `json:"data"`
}

func (c Client) Waypoints(waypoint, filter string) ([]Waypoint, error) {
	var res WaypointsRes
	separator := ""
	if filter != "" {
		separator = "?"
	}
	if err := c.do("systems/{{.system}}/waypoints{{.separator}}{{.filter}}", &res, Do{Template: map[string]string{
		"system":    strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"separator": separator,
		"filter":    filter,
	}}); err != nil {
		return nil, err
	}
	return res.Data, nil
}

type ShipyardShip struct {
	Type string `json:"type"`
}

type Shipyard struct {
	Ships []ShipyardShip `json:"ships"`
}

type ShipyardRes struct {
	Data Shipyard `json:"data"`
}

func (c Client) Shipyard(waypoint string) (Shipyard, error) {
	var res ShipyardRes
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}/shipyard", &res, Do{Template: map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}}); err != nil {
		return Shipyard{}, err
	}
	return res.Data, nil
}

type Item struct {
	Symbol string `json:"symbol"`
}

type Market struct {
	Exports  []Item `json:"exports"`
	Imports  []Item `json:"imports"`
	Exchange []Item `json:"exchange"`
}

type MarketRes struct {
	Data Market `json:"data"`
}

func (c Client) Market(waypoint string) (Market, error) {
	var res MarketRes
	if err := c.do("systems/{{.system}}/waypoints/{{.waypoint}}/market", &res, Do{Template: map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}}); err != nil {
		return Market{}, err
	}
	return res.Data, nil
}

type Deliver struct {
	TradeSymbol       string `json:"tradeSymbol"`
	DestinationSymbol string `json:"destinationSymbol"`
	UnitsRequired     int    `json:"unitsRequired"`
	UnitsFulfilled    int    `json:"unitsFulfilled"`
}

type Terms struct {
	Deliver []Deliver `json:"deliver"`
}

type Contract struct {
	ID        string `json:"id"`
	Accepted  bool   `json:"accepted"`
	Fulfilled bool   `json:"fulfilled"`
	Terms     Terms  `json:"terms"`
}

type ContractsRes struct {
	Data []Contract `json:"data"`
}

func (c Client) Contracts() ([]Contract, error) {
	var res ContractsRes
	if err := c.do("my/contracts", &res, Do{}); err != nil {
		return nil, err
	}
	return res.Data, nil
}

type DeliverRes struct {
}

func (c Client) Deliver(ID, ship, trade string, units int) (DeliverRes, error) {
	var res DeliverRes
	if err := c.do("my/contracts/{{.id}}/deliver", &res, Do{Method: "POST", Template: map[string]string{
		"id": ID,
	}, Payload: map[string]any{
		"shipSymbol":  ship,
		"tradeSymbol": trade,
		"units":       units,
	}}); err != nil {
		return DeliverRes{}, err
	}
	return res, nil
}

type AcceptRes struct {
}

func (c Client) Accept(id string) (AcceptRes, error) {
	var res AcceptRes
	if err := c.do("my/contracts/{{.id}}/accept", &res, Do{Method: "POST", Template: map[string]string{
		"id": id,
	}}); err != nil {
		return AcceptRes{}, err
	}
	return res, nil
}

type FulfillRes struct {
}

func (c Client) Fulfill(id string) (FulfillRes, error) {
	var res FulfillRes
	if err := c.do("my/contracts/{{.id}}/fulfill", &res, Do{Method: "POST", Template: map[string]string{
		"id": id,
	}}); err != nil {
		return FulfillRes{}, err
	}
	return res, nil
}

type Registration struct {
	Role string `json:"role"`
}

type Ship struct {
	Symbol       string       `json:"symbol"`
	Registration Registration `json:"registration"`
}

type ShipsRes struct {
	Data []Ship `json:"data"`
}

func (c Client) Ships() ([]Ship, error) {
	var res ShipsRes
	if err := c.do("my/ships", &res, Do{}); err != nil {
		return nil, err
	}
	return res.Data, nil
}

type BuyData struct {
	Ship Ship `json:"ship"`
}

type BuyRes struct {
	Data BuyData `json:"data"`
}

func (c Client) Buy(waypoint, shipType string) (Ship, error) {
	var res BuyRes
	if err := c.do("my/ships", &res, Do{Method: "POST", Payload: map[string]any{
		"shipType":       shipType,
		"waypointSymbol": waypoint,
	}}); err != nil {
		return Ship{}, nil
	}
	return res.Data.Ship, nil
}

type NegotiateRes struct {
}

func (c Client) Negotiate(ship string) (NegotiateRes, error) {
	var res NegotiateRes
	if err := c.do("my/ships/{{.ship}}/negotiate/contract", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}}); err != nil {
		return NegotiateRes{}, err
	}
	return res, nil
}

type OrbitErrorData struct {
	SecondsToArrival int `json:"secondsToArrival"`
}

type OrbitError struct {
	Data OrbitErrorData `json:"data"`
}

type Nav struct {
	WaypointSymbol string `json:"waypointSymbol"`
}

type Orbit struct {
	Nav Nav `json:"nav"`
}

type OrbitRes struct {
	Data  Orbit      `json:"data"`
	Error OrbitError `json:"error"`
}

func (c Client) Orbit(ship string) (OrbitRes, error) {
	var res OrbitRes
	if err := c.do("my/ships/{{.ship}}/orbit", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}}); err != nil {
		return OrbitRes{}, err
	}
	return res, nil
}

type DockErrorData struct {
	SecondsToArrival int `json:"secondsToArrival"`
}

type DockError struct {
	Data DockErrorData `json:"data"`
}

type DockRes struct {
	Error DockError `json:"error"`
}

func (c Client) Dock(ship string) (DockRes, error) {
	var res DockRes
	if err := c.do("my/ships/{{.ship}}/dock", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}}); err != nil {
		return DockRes{}, err
	}
	return res, nil
}

type RefuelRes struct {
}

func (c Client) Refuel(ship string) (RefuelRes, error) {
	var res RefuelRes
	if err := c.do("my/ships/{{.ship}}/refuel", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}}); err != nil {
		return RefuelRes{}, err
	}
	return res, nil
}

type ExtractErrorCooldown struct {
	RemainingSeconds int `json:"remainingSeconds"`
}

type ExtractErrorData struct {
	Cooldown ExtractErrorCooldown `json:"cooldown"`
}

type ExtractError struct {
	Data ExtractErrorData `json:"data"`
}

type ExtractRes struct {
	Error ExtractError `json:"error"`
}

func (c Client) Extract(ship string) (ExtractRes, error) {
	var res ExtractRes
	if err := c.do("my/ships/{{.ship}}/extract", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}}); err != nil {
		return ExtractRes{}, err
	}
	return res, nil
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

type CargoRes struct {
	Data Cargo `json:"data"`
}

func (c Client) Cargo(ship string) (Cargo, error) {
	var res CargoRes
	if err := c.do("my/ships/{{.ship}}/cargo", &res, Do{Template: map[string]string{
		"ship": ship,
	}}); err != nil {
		return Cargo{}, err
	}
	return res.Data, nil
}

type NavigateRes struct {
}

func (c Client) Navigate(ship, symbol string) (NavigateRes, error) {
	var res NavigateRes
	if err := c.do("my/ships/{{.ship}}/navigate", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{
		"waypointSymbol": symbol,
	}}); err != nil {
		return NavigateRes{}, err
	}
	return res, nil
}

type SellRes struct {
}

func (c Client) Sell(ship, symbol string, units int) (SellRes, error) {
	var res SellRes
	if err := c.do("my/ships/{{.ship}}/sell", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{
		"symbol": symbol,
		"units":  units,
	}}); err != nil {
		return SellRes{}, err
	}
	return res, nil
}

type JettisonRes struct {
}

func (c Client) Jettison(ship, symbol string, units int) (JettisonRes, error) {
	var res JettisonRes
	if err := c.do("my/ships/{{.ship}}/jettison", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{
		"symbol": symbol,
		"units":  units,
	}}); err != nil {
		return JettisonRes{}, err
	}
	return res, nil
}

type RefineError struct {
	Code int `json:"code"`
}

type RefineRes struct {
	Error RefineError `json:"error"`
}

func (c Client) Refine(ship, produce string) (RefineRes, error) {
	var res RefineRes
	if err := c.do("my/ships/{{.ship}}/refine", &res, Do{Method: "POST", Template: map[string]string{
		"ship": ship,
	}, Payload: map[string]any{
		"produce": produce,
	}}); err != nil {
		return RefineRes{}, err
	}
	return res, nil
}

type Do struct {
	IsAccount bool
	Method    string
	Template  map[string]string
	Payload   map[string]any
}

func (c Client) do(pathTemplate string, value any, cfg Do) error {
	url, err := getURL(pathTemplate, cfg)
	if err != nil {
		return err
	}
	if cfg.Payload == nil {
		cfg.Payload = map[string]any{}
	}
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
	token := c.AgentToken
	if cfg.IsAccount {
		token = c.AccountToken
	}
	req.Header.Add("Authorization", "Bearer "+token)
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
	fmt.Printf(" => %d %s\n", resp.StatusCode, url)
	fmt.Println(prettyJSON.String())
	fmt.Printf(" => %d %s\n", resp.StatusCode, url)
	return json.Unmarshal(body, value)
}

const (
	baseUrl = "https://api.spacetraders.io/v2"
)

func getURL(pathTemplate string, cfg Do) (string, error) {
	parsedTemplate, err := template.New("pathTemplate").Parse(pathTemplate)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err = parsedTemplate.Execute(&builder, cfg.Template); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", baseUrl, builder.String()), nil
}
