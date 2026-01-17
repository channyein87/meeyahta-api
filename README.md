# meeyahta-api

Simple Go HTTP server that proxies trip requests to the Transport NSW API.

## Setup

1. Create a `config.json` file in the project root:

```json
{ "apikey": "<transport_nsw_api_key>" }
```

2. Run the server:

```bash
go run .
```

The server listens on port `3000`.

## Trip endpoint

`POST /api/trip`

Request body:

```json
{
  "origin": "10101102",
  "destination": "10101137"
}
```

The server:
- Builds the upstream query with the current Australia/Sydney date (`itdDate`) and time (`itdTime`).
- Sends the request to `https://api.transport.nsw.gov.au/v1/tp/trip` with the API key from `config.json`.
- Flattens each journey into a list of trips, extracting the origin and destination station names (without the `"Station"` suffix) and formatting departure/arrival times in Australia/Sydney time (e.g., `08:00 PM`).

Response:

```json
{
  "trips": [
    {
      "origin": "Redfern",
      "destination": "Town Hall",
      "departureTime": "07:00 PM",
      "arrivalTime": "07:30 PM"
    }
  ]
}
```
