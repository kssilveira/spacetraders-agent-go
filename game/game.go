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
	headquarters, err := getStringField(agent, "headquarters")
	if err != nil {
		return "", err
	}
	waypoint, err := g.waypoint(headquarters)
	if err != nil {
		return "", err
	}
	fmt.Printf("%d\n", len(waypoint))
	return headquarters, nil
}

func (g Game) acceptContract() (map[string]any, error) {
	contracts, err := g.myContracts()
	if err != nil {
		return nil, err
	}
	contract, err := getMap(contracts[0])
	if err != nil {
		return nil, err
	}
	isAccepted, err := getBoolField(contract, "accepted")
	if err != nil {
		return nil, err
	}
	if !isAccepted {
		id, err := getStringField(contract, "id")
		if err != nil {
			return nil, err
		}
		accepted, err := g.accept(id)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%d\n", len(accepted))
	}
	terms, err := getMapField(contract, "terms")
	if err != nil {
		return nil, err
	}
	delivers, err := getListField(terms, "deliver")
	if err != nil {
		return nil, err
	}
	res := map[string]any{}
	for i := range delivers {
		deliver, err := getMap(delivers[i])
		if err != nil {
			return nil, err
		}
		tradeSymbol, err := getStringField(deliver, "tradeSymbol")
		if err != nil {
			return nil, err
		}
		res[tradeSymbol] = nil
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
	for i := range ships {
		ship, err := getMap(ships[i])
		if err != nil {
			return "", err
		}
		registration, err := getMapField(ship, "registration")
		if err != nil {
			return "", err
		}
		role, err := getStringField(registration, "role")
		if err != nil {
			return "", err
		}
		if role != "EXCAVATOR" {
			continue
		}
		symbol, err := getStringField(ship, "symbol")
		if err != nil {
			return "", err
		}
		return symbol, nil
	}
	return "", nil
}

func (g Game) doBuyShip(headquarters string) (string, error) {
	shipyardWaypoints, err := g.waypointsWithFilter(headquarters, "traits=SHIPYARD")
	if err != nil {
		return "", err
	}
	for i := range shipyardWaypoints {
		shipyardWaypoint, err := getMap(shipyardWaypoints[i])
		if err != nil {
			return "", err
		}
		symbol, err := getStringField(shipyardWaypoint, "symbol")
		if err != nil {
			return "", err
		}
		shipyard, err := g.waypointAction(symbol, "shipyard")
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
			got, err := g.myShipsBuy(symbol, "SHIP_MINING_DRONE")
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
	asteroidWaypoint, err := getMap(asteroidWaypoints[0])
	if err != nil {
		return err
	}
	asteroid, err := getStringField(asteroidWaypoint, "symbol")
	if err != nil {
		return err
	}
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

func (g Game) myAgent() (map[string]any, error) {
	return g.do("my/agent", "GET", nil, nil)
}

func (g Game) waypoint(waypoint string) (map[string]any, error) {
	return g.do("systems/{{.system}}/waypoints/{{.waypoint}}", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
	}, nil)
}

func (g Game) waypointsWithFilter(waypoint, filter string) ([]any, error) {
	data, err := g.do("systems/{{.system}}/waypoints?{{.filter}}", "GET", map[string]string{
		"system": strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"filter": filter,
	}, nil)
	if err != nil {
		return nil, err
	}
	return getListField(data, "data")
}

func (g Game) waypointAction(waypoint, action string) (map[string]any, error) {
	return g.do("systems/{{.system}}/waypoints/{{.waypoint}}/{{.action}}", "GET", map[string]string{
		"system":   strings.Join(strings.Split(waypoint, "-")[:2], "-"),
		"waypoint": waypoint,
		"action":   action,
	}, nil)
}

func (g Game) myContracts() ([]any, error) {
	data, err := g.do("my/contracts", "GET", nil, nil)
	if err != nil {
		return nil, err
	}
	return getListField(data, "data")
}

func (g Game) accept(id string) (map[string]any, error) {
	return g.do("my/contracts/{{.id}}/accept", "POST", map[string]string{
		"id": id,
	}, nil)
}

func (g Game) myShips() ([]any, error) {
	data, err := g.do("my/ships", "GET", nil, nil)
	if err != nil {
		return nil, err
	}
	return getListField(data, "data")
}

func (g Game) myShipsBuy(waypoint, shipType string) (map[string]any, error) {
	return g.do("my/ships", "POST", nil, map[string]any{
		"shipType":       shipType,
		"waypointSymbol": waypoint,
	})
}

func (g Game) myShipsAction(ship, action, method string) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/{{.action}}", method, map[string]string{
		"ship":   ship,
		"action": action,
	}, map[string]any{})
}

func (g Game) myShipsNavigate(ship, symbol string) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/navigate", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"waypointSymbol": symbol,
	})
}

func (g Game) myShipsSell(ship, symbol string, units int) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/sell", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"symbol": symbol,
		"units":  units,
	})
}

func (g Game) myShipsJettison(ship, symbol string, units int) (map[string]any, error) {
	return g.do("my/ships/{{.ship}}/jettison", "POST", map[string]string{
		"ship": ship,
	}, map[string]any{
		"symbol": symbol,
		"units":  units,
	})
}

const (
	baseUrl = "https://api.spacetraders.io/v2"
)

func (g Game) do(pathTemplate, method string, templateData map[string]string, payloadData map[string]any) (map[string]any, error) {
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
