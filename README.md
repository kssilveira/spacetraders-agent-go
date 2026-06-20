# spacetraders-agent-go

https://spacetraders.io agent in Go.

## TODO

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/survey'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/extract' \
 --header 'Content-Type: application/json' \
 --data '{
    "survey": "null"
   }'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/contracts/:contractId/fulfill'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/negotiate/contract' \
 --header 'Authorization: Bearer '
```

```
openapi-generator generate \
 -i https://spacetraders.io/SpaceTraders.json \
 -o packages/spacetraders-sdk \
 -g typescript-axios \
 --additional-properties=npmName="spacetraders-sdk" \
 --additional-properties=npmVersion="2.3.0" \
 --additional-properties=supportsES6=true \
 --additional-properties=withSeparateModelsAndApi=true \
 --additional-properties=modelPackage="models" \
 --additional-properties=apiPackage="api"
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/register' \
 --header 'Authorization: Bearer ACCOUNT_TOKEN' \
 --header 'Content-Type: application/json' \
 --data '{
    "symbol": "",
    "faction": ""
   }'
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

```
curl --request PATCH \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/nav' \
 --header 'Content-Type: application/json' \
 --data '{
    "flightMode": ""
   }'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/warp' \
 --header 'Content-Type: application/json' \
 --data '{
    "systemSymbol": ""
   }'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/jump' \
 --header 'Content-Type: application/json' \
 --data '{
    "systemSymbol": ""
   }'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/refuel' \
 --header 'Content-Type: application/json' \
 --data '{
    "fromCargo": ""
   }'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/siphon'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/repair'
```

```
curl 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/repair'
```

```
curl --request POST \
 --url 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/scrap'
```

```
curl 'https://api.spacetraders.io/v2/my/ships/:shipSymbol/scrap'
```
