# Convert Interface to JSON Schema

```bash
go run ./... main.go
```

So an interface:

```go
type Weather interface {
    // Retrieves current weather for the given location
    GetWeather(
        location string, // City and country e.g. Bogotá, Colombia
    ) (temp int)
}
```

Gives this output:

```json
{
  "type": "function",
  "function": {
    "name": "GetWeather",
    "description": "Retrieves current weather for the given location",
    "parameters": {
      "type": "object",
      "properties": {
        "location": {
          "type": "string",
          "description": "City and country e.g. Bogotá, Colombia"
        }
      },
      "required": [
        "location"
      ],
      "additionalProperties": false
    },
    "strict": true
  }
}
```
