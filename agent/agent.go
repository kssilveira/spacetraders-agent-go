package agent

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/kssilveira/spacetraders-agent-go/client"
	"github.com/kssilveira/spacetraders-agent-go/token"
)

type State struct {
	Credits         int
	SymbolToDeliver map[string]client.Deliver
	SymbolToCargo   map[string]client.Inventory
	Capacity        int
}

type Agent struct {
	Client client.Client
	State  State
}

func (a *Agent) Run(args []string) error {
	headquarters, err := a.getHeadquarters()
	if err != nil {
		return err
	}
	if len(args) == 1 {
		return a.fulfillContracts(headquarters)
	}
	waypoints, err := a.waypoints(headquarters)
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", len(waypoints))
	return nil
}

var (
	interestingTrait = map[string]any{
		"MINERAL_DEPOSITS":        nil,
		"COMMON_METAL_DEPOSITS":   nil,
		"PRECIOUS_METAL_DEPOSITS": nil,
		"RARE_METAL_DEPOSITS":     nil,
		"ICE_CRYSTALS":            nil,
		"METHANE_POOLS":           nil,
		"EXPLOSIVE_GASES":         nil,
		"UNSTABLE_COMPOSITION":    nil,
		"STRIPPED":                nil,
		"MARKETPLACE":             nil,
		"SHIPYARD":                nil,
		"UNCHARTED":               nil,
	}
)

func (a *Agent) waypoints(headquarters string) ([]client.Waypoint, error) {
	page := 1
	res := []client.Waypoint{}
	hqWaypoint, err := a.Client.Waypoint(headquarters)
	if err != nil {
		return nil, err
	}
	for {
		waypoints, err := a.Client.Waypoints(headquarters, fmt.Sprintf("page=%d&limit=20", page))
		if err != nil {
			return nil, err
		}
		if len(waypoints) == 0 {
			break
		}
		res = append(res, waypoints...)
		page++
	}
	for i, waypoint := range res {
		waypoint.Distance = int(math.Hypot(float64(hqWaypoint.X-waypoint.X), float64(hqWaypoint.Y-waypoint.Y)))
		for _, trait := range waypoint.Traits {
			if trait.Symbol == "MARKETPLACE" {
				market, err := a.Client.Market(waypoint.Symbol)
				if err != nil {
					return nil, err
				}
				waypoint.Exports = market.Exports
				waypoint.Imports = market.Imports
				waypoint.Exchange = market.Exchange
			}
			if trait.Symbol == "SHIPYARD" {
				shipyard, err := a.Client.Shipyard(waypoint.Symbol)
				if err != nil {
					return nil, err
				}
				fmt.Printf("%#v\n", shipyard)
			}
		}
		res[i] = waypoint
	}
	slices.SortFunc(res, func(a, b client.Waypoint) int {
		return cmp.Compare(a.Distance, b.Distance)
	})
	fmt.Printf("%-*s %-*s %4s %4s %3s traits\n", len("X1-UN88-EE5F"), "symbol", len("ENGINEERED_ASTEROID"), "type", "x", "y", "d")
	for _, waypoint := range res {
		fmt.Printf("%-*s %-*s %4d %4d %3d %s\n", len("X1-UN88-EE5F"), waypoint.Symbol, len("ENGINEERED_ASTEROID"), waypoint.Type, waypoint.X, waypoint.Y, waypoint.Distance, symbolsFromTraits(waypoint.Traits, interestingTrait))
		for _, list := range []string{symbolsFromItems("exports", waypoint.Exports), symbolsFromItems("imports", waypoint.Imports), symbolsFromItems("exchange", waypoint.Exchange)} {
			if list == "" {
				continue
			}
			fmt.Printf("%s\n", list)
		}
	}
	fmt.Printf("%-*s %-*s %4s %4s %3s traits\n", len("X1-UN88-EE5F"), "symbol", len("ENGINEERED_ASTEROID"), "type", "x", "y", "d")
	return res, nil
}

func symbolsFromTraits(items []client.Trait, interesting map[string]any) string {
	res := []string{}
	for _, item := range items {
		if interesting != nil {
			if _, ok := interesting[item.Symbol]; !ok {
				continue
			}
		}
		res = append(res, item.Symbol)
	}
	return strings.Join(res, ", ")
}

func symbolsFromItems(name string, items []client.Item) string {
	res := []string{}
	for _, item := range items {
		res = append(res, item.Symbol)
	}
	if len(res) == 0 {
		return ""
	}
	return fmt.Sprintf("%-*s %s", len("X1-UN88-EE5F"), name, strings.Join(res, ", "))
}

func (a *Agent) fulfillContracts(headquarters string) error {
	ship, err := a.maybeBuyShip(headquarters)
	if err != nil {
		return err
	}
	for {
		contractID, err := a.acceptContract(ship)
		if err != nil {
			return err
		}
		found := false
		for symbol, deliver := range a.State.SymbolToDeliver {
			if deliver.UnitsFulfilled < deliver.UnitsRequired {
				fmt.Printf("%s %#v\n", symbol, deliver)
				found = true
				break
			}
		}
		if !found {
			break
		}
		isDone, units, err := a.isDone(ship)
		if err != nil {
			return err
		}
		if !isDone {
			if err := a.navigateAndExtract(headquarters, ship); err != nil {
				return err
			}
			isDone, units, err = a.isDone(ship)
			if err != nil {
				return err
			}
		}
		if err := a.deliver(contractID, ship, units); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) getHeadquarters() (string, error) {
	agent, err := a.Client.Agent()
	if err != nil {
		return "", err
	}
	a.State.Credits = agent.Data.Credits
	if agent.Error.Code != 0 {
		register, err := a.Client.Register("KAUE", "AEGIS")
		if err != nil {
			return "", err
		}
		if err := token.SetAgentToken(register.Token); err != nil {
			return "", err
		}
		a.Client.AgentToken = register.Token
		agent, err = a.Client.Agent()
		if err != nil {
			return "", err
		}
	}
	return agent.Data.Headquarters, nil
}

var (
	symbolToSource = map[string]string{
		"IRON": "IRON_ORE",
	}
)

func (a *Agent) acceptContract(ship string) (string, error) {
	contracts, err := a.Client.Contracts()
	if err != nil {
		return "", err
	}
	for _, contract := range contracts {
		if !contract.Accepted {
			accepted, err := a.Client.Accept(contract.ID)
			if err != nil {
				return "", err
			}
			fmt.Printf("%#v\n", accepted)
		}
		if !contract.Fulfilled {
			a.State.SymbolToDeliver = map[string]client.Deliver{}
			found := false
			for _, deliver := range contract.Terms.Deliver {
				a.State.SymbolToDeliver[deliver.TradeSymbol] = deliver
				if source, ok := symbolToSource[deliver.TradeSymbol]; ok {
					a.State.SymbolToDeliver[source] = deliver
				}
				if deliver.UnitsFulfilled < deliver.UnitsRequired {
					found = true
				}
			}
			if found {
				return contract.ID, nil
			}
			fulfilled, err := a.Client.Fulfill(contract.ID)
			if err != nil {
				return "", err
			}
			fmt.Printf("%#v\n", fulfilled)
		}
	}
	negotiated, err := a.Client.Negotiate(ship)
	if err != nil {
		return "", err
	}
	fmt.Printf("%#v\n", negotiated)
	return a.acceptContract(ship)
}

func (a *Agent) maybeBuyShip(headquarters string) (string, error) {
	symbol, err := a.excavator()
	if err != nil {
		return "", err
	}
	if symbol != "" {
		return symbol, nil
	}
	ship, err := a.doBuyShip(headquarters)
	if err != nil {
		return "", err
	}
	refuel, err := a.Client.Refuel(ship)
	if err != nil {
		return "", err
	}
	fmt.Printf("%#v\n", refuel)
	return ship, nil
}

func (a *Agent) excavator() (string, error) {
	ships, err := a.Client.Ships()
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

func (a *Agent) doBuyShip(headquarters string) (string, error) {
	shipyardWaypoints, err := a.Client.Waypoints(headquarters, "traits=SHIPYARD")
	if err != nil {
		return "", err
	}
	for _, shipyardWaypoint := range shipyardWaypoints {
		shipyard, err := a.Client.Shipyard(shipyardWaypoint.Symbol)
		if err != nil {
			return "", err
		}
		for _, ship := range shipyard.Ships {
			if ship.Type != "SHIP_MINING_DRONE" {
				continue
			}
			got, err := a.Client.Buy(shipyardWaypoint.Symbol, "SHIP_MINING_DRONE")
			if err != nil {
				return "", err
			}
			return got.Symbol, nil
		}
	}
	return "", fmt.Errorf("failed to buy ship")
}

func (a *Agent) navigateAndExtract(headquarters, ship string) error {
	orbit, err := a.orbit(ship)
	if err != nil {
		return err
	}
	asteroidWaypoints, err := a.Client.Waypoints(headquarters, "type=ENGINEERED_ASTEROID")
	if err != nil {
		return err
	}
	asteroid := asteroidWaypoints[0].Symbol
	if orbit.Data.Nav.WaypointSymbol != asteroid {
		navigate, err := a.Client.Navigate(ship, asteroid)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", navigate)
	}
	if err := a.dock(ship); err != nil {
		return err
	}
	refuel, err := a.Client.Refuel(ship)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", refuel)
	market, err := a.Client.Market(orbit.Data.Nav.WaypointSymbol)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", market)
	if err := a.extract(ship); err != nil {
		return err
	}
	return nil
}

func (a *Agent) extract(ship string) error {
	isOrbit := false
	var err error
	for {
		isOrbit, err = a.sell(ship, isOrbit)
		if err != nil {
			return err
		}
		isDone, _, err := a.isDone(ship)
		if err != nil {
			return err
		}
		if isDone {
			return nil
		}
		if !isOrbit {
			isOrbit = true
			orbit, err := a.Client.Orbit(ship)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", orbit)
		}
		extract, err := a.Client.Extract(ship)
		if err != nil {
			return err
		}
		if err := a.sleep(extract.Error.Data.Cooldown.RemainingSeconds); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) sell(ship string, isOrbit bool) (bool, error) {
	cargo, err := a.Client.Cargo(ship)
	if err != nil {
		return false, err
	}
	for _, item := range cargo.Inventory {
		a.State.SymbolToCargo = map[string]client.Inventory{}
		if _, ok := a.State.SymbolToDeliver[item.Symbol]; ok {
			a.State.SymbolToCargo[item.Symbol] = item
			continue
		}
		if isOrbit {
			isOrbit = false
			dock, err := a.Client.Dock(ship)
			if err != nil {
				return false, err
			}
			fmt.Printf("%#v\n", dock)
		}
		sell, err := a.Client.Sell(ship, item.Symbol, item.Units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%#v\n", sell)
		jettison, err := a.Client.Jettison(ship, item.Symbol, item.Units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%#v\n", jettison)
	}
	return isOrbit, nil
}

func (a *Agent) isDone(ship string) (bool, int, error) {
	cargo, err := a.Client.Cargo(ship)
	if err != nil {
		return false, 0, err
	}
	a.State.Capacity = cargo.Capacity
	if cargo.Units == cargo.Capacity && len(cargo.Inventory) == 1 {
		have := cargo.Inventory[0].Symbol
		want := a.State.SymbolToDeliver[have].TradeSymbol
		if want != have {
			refine, err := a.Client.Refine(ship, want)
			if err != nil {
				return false, 0, err
			}
			if refine.Error.Code != 0 {
				return false, 0, fmt.Errorf("%#v", refine)
			}
			return false, 0, nil
		}
		return true, cargo.Units, nil
	}
	return false, 0, nil
}

func (a *Agent) deliver(contractID, ship string, units int) error {
	for trade, deliver := range a.State.SymbolToDeliver {
		orbit, err := a.Client.Orbit(ship)
		if err != nil {
			return err
		}
		if orbit.Data.Nav.WaypointSymbol != deliver.DestinationSymbol {
			navigate, err := a.Client.Navigate(ship, deliver.DestinationSymbol)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", navigate)
		}
		if err := a.dock(ship); err != nil {
			return err
		}
		deliver, err := a.Client.Deliver(contractID, ship, trade, units)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", deliver)
	}
	return nil
}

func (a *Agent) dock(ship string) error {
	dock, err := a.Client.Dock(ship)
	if err != nil {
		return err
	}
	if dock.Error.Data.SecondsToArrival != 0 {
		if err := a.sleep(dock.Error.Data.SecondsToArrival); err != nil {
			return err
		}
		dock, err = a.Client.Dock(ship)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", dock)
	}
	return nil
}

func (a *Agent) orbit(ship string) (client.OrbitRes, error) {
	orbit, err := a.Client.Orbit(ship)
	if err != nil {
		return client.OrbitRes{}, err
	}
	if orbit.Error.Data.SecondsToArrival != 0 {
		if err := a.sleep(orbit.Error.Data.SecondsToArrival); err != nil {
			return client.OrbitRes{}, err
		}
		orbit, err = a.Client.Orbit(ship)
		if err != nil {
			return client.OrbitRes{}, err
		}
	}
	return orbit, nil
}

func (a *Agent) sleep(seconds int) error {
	if seconds == 0 {
		return nil
	}
	fmt.Printf("sleep %d\n", seconds)
	state, err := json.MarshalIndent(a.State, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("state %s\n", string(state))
	time.Sleep(time.Duration(seconds) * time.Second)
	return nil
}
