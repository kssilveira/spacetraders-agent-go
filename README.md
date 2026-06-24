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

symbol       type                   x    y   d traits

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

### List waypoints with filter

#### Asteroids

```
$ go run main.go -- waypoints type=ASTEROID |& tee ~/asteroids

symbol       type                   x    y   d traits
X1-UN88-B10  ASTEROID              31  308 285 MINERAL_DEPOSITS
X1-UN88-B9   ASTEROID            -191  252 294 COMMON_METAL_DEPOSITS
X1-UN88-B35  ASTEROID             154  277 296 COMMON_METAL_DEPOSITS
X1-UN88-B38  ASTEROID               2  325 300 MINERAL_DEPOSITS
X1-UN88-B12  ASTEROID            -316   56 314 COMMON_METAL_DEPOSITS
```

#### Marketplaces

```
$ go run main.go -- waypoints traits=MARKETPLACE |& tee ~/marketplaces

symbol       type                   x    y   d traits

X1-UN88-A2   MOON                  -3   25   0 MARKETPLACE, SHIPYARD
imports      SHIP_PLATING, SHIP_PARTS
exchange     FUEL
types        SHIP_PROBE, SHIP_LIGHT_SHUTTLE, SHIP_LIGHT_HAULER

X1-UN88-A1   PLANET                -3   25   0 MARKETPLACE
imports      FOOD, MEDICINE, CLOTHING, EQUIPMENT, JEWELRY, HOLOGRAPHICS
exchange     FUEL

X1-UN88-A4   ORBITAL_STATION       -3   25   0 MARKETPLACE
exports      NOVEL_LIFEFORMS
imports      LAB_INSTRUMENTS, EQUIPMENT
exchange     FUEL
```

#### Shipyards

```
$ go run main.go -- waypoints traits=SHIPYARD |& tee ~/shipyards

symbol       type                   x    y   d traits

X1-UN88-A2   MOON                  -3   25   0 MARKETPLACE, SHIPYARD
types        SHIP_PROBE, SHIP_LIGHT_SHUTTLE, SHIP_LIGHT_HAULER

X1-UN88-H52  MOON                 -44   11  43 MARKETPLACE, SHIPYARD
types        SHIP_MINING_DRONE, SHIP_SURVEYOR
ships
  SHIP_MINING_DRONE, FRAME_DRONE, REACTOR_CHEMICAL_I, ENGINE_IMPULSE_DRIVE_I, MODULE_CARGO_HOLD_I, MODULE_MINERAL_PROCESSOR_I, MOUNT_MINING_LASER_I
  SHIP_SURVEYOR, FRAME_DRONE, REACTOR_CHEMICAL_I, ENGINE_IMPULSE_DRIVE_I, MOUNT_SURVEYOR_I

X1-UN88-C41  ORBITAL_STATION      -31 -150 177 MARKETPLACE, SHIPYARD
types        SHIP_PROBE, SHIP_SIPHON_DRONE
```

## TODO

- go to the markets to sell

- calculate profit given the fuel cost

- remember the profit per item and throw away items with negative profit

- handle expired survey
