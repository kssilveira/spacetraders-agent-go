# spacetraders-agent-go

https://spacetraders.io agent in Go using https://spacetraders.io/openapi.

## TODO

- change main to stop buying a new ship and instead use the main ship

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

