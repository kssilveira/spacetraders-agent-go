# spacetraders-agent-go

https://spacetraders.io agent in Go using https://spacetraders.io/openapi.

## Features

### Fulfill contracts

```
$ go run main.go -- contracts |& tee ~/contracts
```

### List ships

```
$ go run main.go -- ships |& tee ~/ships

ships

  KAUE5-1, COMMAND, X1-UN88-EE5F, IN_ORBIT, FRAME_FRIGATE, REACTOR_FISSION_I, ENGINE_ION_DRIVE_II, MODULE_CARGO_HOLD_II, MODULE_CREW_QUARTERS_I, MODULE_CREW_QUARTERS_I, MODULE_MINERAL_PROCESSOR_I, MODULE_GAS_PROCESSOR_I, MOUNT_SENSOR_ARRAY_II, MOUNT_GAS_SIPHON_II, MOUNT_MINING_LASER_II, MOUNT_SURVEYOR_II

  KAUE5-2, SATELLITE, X1-UN88-H52, IN_ORBIT, FRAME_PROBE, REACTOR_SOLAR_I, ENGINE_IMPULSE_DRIVE_I

  KAUE5-3, EXCAVATOR, X1-UN88-EE5F, IN_ORBIT, FRAME_DRONE, REACTOR_CHEMICAL_I, ENGINE_IMPULSE_DRIVE_I, MODULE_CARGO_HOLD_I, MODULE_MINERAL_PROCESSOR_I, MOUNT_MINING_LASER_I
```

### List waypoints

```
$ go run main.go -- waypoints |& tee ~/waypoints

X1-UN88-EE5F ENGINEERED_ASTEROID  -24    4  29 COMMON_METAL_DEPOSITS, STRIPPED, MARKETPLACE
exchange     FUEL

X1-UN88-H54  MOON                 -44   11  43 MARKETPLACE
exports      JEWELRY
imports      GOLD, SILVER, PRECIOUS_STONES, DIAMONDS
exchange     FUEL

X1-UN88-H53  MOON                 -44   11  43 MARKETPLACE
exchange     ICE_WATER, AMMONIA_ICE, QUARTZ_SAND, SILICON_CRYSTALS, FUEL

X1-UN88-H51  PLANET               -44   11  43 MARKETPLACE
exports      IRON, ALUMINUM, COPPER
imports      DRUGS, IRON_ORE, ALUMINUM_ORE, COPPER_ORE
exchange     FUEL

X1-UN88-H52  MOON                 -44   11  43 MARKETPLACE, SHIPYARD
imports      SHIP_PLATING, SHIP_PARTS
exchange     FUEL
types        SHIP_MINING_DRONE, SHIP_SURVEYOR
ships
  SHIP_MINING_DRONE, FRAME_DRONE, REACTOR_CHEMICAL_I, ENGINE_IMPULSE_DRIVE_I, MODULE_CARGO_HOLD_I, MODULE
_MINERAL_PROCESSOR_I, MOUNT_MINING_LASER_I
  SHIP_SURVEYOR, FRAME_DRONE, REACTOR_CHEMICAL_I, ENGINE_IMPULSE_DRIVE_I, MOUNT_SURVEYOR_I
```

## TODO

- remove prints

- do survey

```
curl https://api.spacetraders.io/v2/my/ships/KAUE5-1/survey \
  --request POST
```

- use survey

```
curl https://api.spacetraders.io/v2/my/ships/KAUE5-1/extract/survey \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{
  "signature": "X1-UN88-EE5F-3EC0C9",
  "symbol": "X1-UN88-EE5F",
  "deposits": [
    {
      "symbol": "IRON_ORE"
    }
  ],
  "expiration": "2026-06-23T18:07:58.422Z",
  "size": "LARGE"
}'
```

- go to the closest ASTEROID or ENGINEERED_ASTEROID that is not STRIPPED

- find the closest market that buys each of the extracted items

- go to the markets to sell

- calculate profit given the fuel cost

- remember the profit per item and throw away items with negative profit

