# spacetraders-agent-go

https://spacetraders.io agent in Go.

## TODO

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/register' \
 --header 'Authorization: Bearer ACCOUNT_TOKEN' \
 --header 'Content-Type: application/json' \
 --data '{
    "symbol": "INSERT_CALLSIGN_HERE",
    "faction": "COSMIC"
   }'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/negotiate/contract' \
 --header 'Authorization: Bearer '
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/contracts/:contractId/fulfill'
```

```
curl --request  \
 --url 'https://api.spacetraders.io/v2/factions'
```

```
curl --request  \
 --url 'https://api.spacetraders.io/v2/systems'
```

```
curl --request  \
 --url 'https://api.spacetraders.io/v2/systems/X1-KY89/waypoints'
```
