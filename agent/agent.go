package agent

import (
	"fmt"
	"time"

	"github.com/kssilveira/spacetraders-agent-go/client"
)

type Agent struct {
	Client client.Client
}

func (a Agent) All() error {
	headquarters, err := a.getHeadquarters()
	if err != nil {
		return err
	}
	ship, err := a.maybeBuyShip(headquarters)
	if err != nil {
		return err
	}
	for {
		contractID, symbolToDeliver, err := a.acceptContract()
		if err != nil {
			return err
		}
		found := false
		for symbol, deliver := range symbolToDeliver {
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
			if err := a.navigateAndExtract(headquarters, ship, symbolToDeliver); err != nil {
				return err
			}
			isDone, units, err = a.isDone(ship)
			if err != nil {
				return err
			}
		}
		if err := a.deliver(contractID, ship, units, symbolToDeliver); err != nil {
			return err
		}
	}
	return nil
}

func (a Agent) getHeadquarters() (string, error) {
	agent, err := a.Client.MyAgent()
	if err != nil {
		return "", err
	}
	headquarters := agent.Headquarters
	waypoint, err := a.Client.Waypoint(headquarters)
	if err != nil {
		return "", err
	}
	fmt.Printf("%#v\n", waypoint)
	return headquarters, nil
}

func (a Agent) acceptContract() (string, map[string]client.Deliver, error) {
	contracts, err := a.Client.MyContracts()
	if err != nil {
		return "", nil, err
	}
	contract := contracts[0]
	if !contract.Accepted {
		accepted, err := a.Client.Accept(contract.ID)
		if err != nil {
			return "", nil, err
		}
		fmt.Printf("%#v\n", accepted)
	}
	res := map[string]client.Deliver{}
	for _, deliver := range contract.Terms.Deliver {
		res[deliver.TradeSymbol] = deliver
	}
	return contract.ID, res, nil
}

func (a Agent) maybeBuyShip(headquarters string) (string, error) {
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
	refuel, err := a.Client.MyShipsRefuel(ship)
	if err != nil {
		return "", err
	}
	fmt.Printf("%#v\n", refuel)
	return ship, nil
}

func (a Agent) excavator() (string, error) {
	ships, err := a.Client.MyShips()
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
	shipyardWaypoints, err := a.Client.WaypointsWithFilter(headquarters, "traits=SHIPYARD")
	if err != nil {
		return "", err
	}
	for _, shipyardWaypoint := range shipyardWaypoints {
		shipyard, err := a.Client.WaypointShipyard(shipyardWaypoint.Symbol)
		if err != nil {
			return "", err
		}
		for _, ship := range shipyard.Ships {
			if ship.Type != "SHIP_MINING_DRONE" {
				continue
			}
			got, err := a.Client.MyShipsBuy(shipyardWaypoint.Symbol, "SHIP_MINING_DRONE")
			if err != nil {
				return "", err
			}
			return got.Symbol, nil
		}
	}
	return "", fmt.Errorf("failed to buy ship")
}

func (a Agent) navigateAndExtract(headquarters, ship string, symbolToDeliver map[string]client.Deliver) error {
	orbit, err := a.Client.MyShipsOrbit(ship)
	if err != nil {
		return err
	}
	asteroidWaypoints, err := a.Client.WaypointsWithFilter(headquarters, "type=ENGINEERED_ASTEROID")
	if err != nil {
		return err
	}
	asteroid := asteroidWaypoints[0].Symbol
	if orbit.Nav.WaypointSymbol != asteroid {
		navigate, err := a.Client.MyShipsNavigate(ship, asteroid)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", navigate)
	}
	if err := a.dock(ship); err != nil {
		return err
	}
	refuel, err := a.Client.MyShipsRefuel(ship)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", refuel)
	market, err := a.Client.WaypointMarket(orbit.Nav.WaypointSymbol)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", market)
	if err := a.extract(ship, symbolToDeliver); err != nil {
		return err
	}
	return nil
}

func (a Agent) extract(ship string, symbolToDeliver map[string]client.Deliver) error {
	isOrbit := false
	var err error
	for {
		isOrbit, err = a.sell(ship, symbolToDeliver, isOrbit)
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
			orbit, err := a.Client.MyShipsOrbit(ship)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", orbit)
		}
		extract, err := a.Client.MyShipsExtract(ship)
		if err != nil {
			return err
		}
		fmt.Printf("sleep %d\n", extract.Error.Data.Cooldown.RemainingSeconds)
		time.Sleep(time.Duration(extract.Error.Data.Cooldown.RemainingSeconds) * time.Second)
	}
	return nil
}

func (a Agent) sell(ship string, symbolToDeliver map[string]client.Deliver, isOrbit bool) (bool, error) {
	cargo, err := a.Client.MyShipsCargo(ship)
	if err != nil {
		return false, err
	}
	for _, item := range cargo.Inventory {
		if _, ok := symbolToDeliver[item.Symbol]; ok {
			continue
		}
		if isOrbit {
			isOrbit = false
			dock, err := a.Client.MyShipsDock(ship)
			if err != nil {
				return false, err
			}
			fmt.Printf("%#v\n", dock)
		}
		sell, err := a.Client.MyShipsSell(ship, item.Symbol, item.Units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%#v\n", sell)
		jettison, err := a.Client.MyShipsJettison(ship, item.Symbol, item.Units)
		if err != nil {
			return false, err
		}
		fmt.Printf("%#v\n", jettison)
	}
	return isOrbit, nil
}

func (a Agent) isDone(ship string) (bool, int, error) {
	cargo, err := a.Client.MyShipsCargo(ship)
	if err != nil {
		return false, 0, err
	}
	if cargo.Units == cargo.Capacity && len(cargo.Inventory) == 1 {
		return true, cargo.Units, nil
	}
	return false, 0, nil
}

func (a Agent) deliver(contractID, ship string, units int, symbolToDeliver map[string]client.Deliver) error {
	for trade, deliver := range symbolToDeliver {
		orbit, err := a.Client.MyShipsOrbit(ship)
		if err != nil {
			return err
		}
		if orbit.Nav.WaypointSymbol != deliver.DestinationSymbol {
			navigate, err := a.Client.MyShipsNavigate(ship, deliver.DestinationSymbol)
			if err != nil {
				return err
			}
			fmt.Printf("%#v\n", navigate)
		}
		if err := a.dock(ship); err != nil {
			return err
		}
		deliver, err := a.Client.MyContractsDeliver(contractID, ship, trade, units)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", deliver)
	}
	return nil
}

func (a Agent) dock(ship string) error {
	dock, err := a.Client.MyShipsDock(ship)
	if err != nil {
		return err
	}
	fmt.Printf("sleep %d\n", dock.Error.Data.SecondsToArrival)
	time.Sleep(time.Duration(dock.Error.Data.SecondsToArrival) * time.Second)
	dock, err = a.Client.MyShipsDock(ship)
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", dock)
	return nil
}
