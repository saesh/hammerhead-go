# hammerhead

A Go client library for the [Hammerhead Karoo](https://www.hammerhead.io) API.

## Installation

```bash
go get github.com/saesh/hammerhead-go
```

Requires Go 1.21+. No external dependencies.

## Authentication

The Hammerhead API uses OAuth 2.0. You'll need a client ID and secret from Hammerhead, and an access token obtained through the authorization code flow.

### Step 1 — Redirect the user

```go
import hammerhead "github.com/saesh/hammerhead-go"

// A placeholder token is required to construct the client; no API call is made here.
client, err := hammerhead.NewClient("placeholder")
if err != nil {
    log.Fatal(err)
}

authURL := client.AuthorizeURL(
    "your-client-id",
    "https://yourapp.com/oauth/callback",
    []hammerhead.Scope{
        hammerhead.ScopeActivityRead,
        hammerhead.ScopeRouteRead,
        hammerhead.ScopeRouteWrite,
    },
    "random-csrf-state",
)

// Redirect the user to authURL
```

### Step 2 — Exchange the code for a token

After the user authorizes your app, Hammerhead redirects to your `redirect_uri` with a `code` query parameter.

```go
token, err := client.ExchangeToken(ctx,
    r.URL.Query().Get("code"),
    "your-client-id",
    "your-client-secret",
    "https://yourapp.com/oauth/callback",
)
if err != nil {
    // handle error
}

// Store token.AccessToken and token.RefreshToken securely
fmt.Println("User ID:", token.UserID)
fmt.Println("Expires in:", token.ExpiresIn, "seconds")
```

### Step 3 — Create an authenticated client

```go
client, err = hammerhead.NewClient(token.AccessToken)
if err != nil {
    log.Fatal(err)
}
```

### Refreshing tokens

```go
newToken, err := client.RefreshToken(ctx,
    storedRefreshToken,
    "your-client-id",
    "your-client-secret",
)
```

## Usage

### Activities

```go
client, err := hammerhead.NewClient("your-access-token")
if err != nil {
    log.Fatal(err)
}
ctx := context.Background()

// List activities (paginated)
list, err := client.ListActivities(ctx, &hammerhead.ActivityListOptions{
    Page:      1,
    PerPage:   20,
    StartDate: "2025-01-01", // YYYY-MM-DD, optional
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("%d total activities\n", list.TotalItems)
for _, a := range list.Activities {
    fmt.Printf("  %s — %s (%.1f km)\n", a.ID, a.Name, a.Distance/1000)
}

// Get a single activity
activity, err := client.GetActivity(ctx, "1000.activity.abcd")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Type:", activity.ActivityType)
fmt.Println("Description:", activity.Description)

// Download the FIT file
fitBytes, err := client.GetActivityFile(ctx, "1000.activity.abcd")
if err != nil {
    log.Fatal(err)
}

if err := os.WriteFile("activity.fit", fitBytes, 0644); err != nil {
    log.Fatal(err)
}
```

### Routes

```go
// List routes
list, err := client.ListRoutes(ctx, &hammerhead.RouteListOptions{
    Page:    1,
    PerPage: 50,
})

// Upload a new route from a file
f, err := os.Open("my-route.gpx")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

route, err := client.CreateRoute(ctx, "my-route.gpx", f)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Created route:", route.ID)

// Supported formats: .gpx, .fit, .tcx, .kml, .kmz

// Update an existing route
f2, _ := os.Open("updated-route.gpx")
defer f2.Close()

route, err = client.UpdateRoute(ctx, "1000.route.abcd", "updated-route.gpx", f2)

// Delete a route
err = client.DeleteRoute(ctx, "1000.route.abcd")
```

### Workouts

```go
// Upload a workout
f, err := os.Open("intervals.zwo")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

workout, err := client.CreateWorkout(ctx, "intervals.zwo", f, "2025-06-15")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Created workout:", workout.ID)

// Supported formats: .fit, .zwo
// Pass "" as plannedDate to omit it

// Update a workout
f2, _ := os.Open("updated.zwo")
defer f2.Close()

workout, err = client.UpdateWorkout(ctx, "1000.workout.abcd", "updated.zwo", f2, "")

// Delete a workout
err = client.DeleteWorkout(ctx, "1000.workout.abcd")
```

## Error handling

Non-2xx responses are returned as `*hammerhead.APIError`, which exposes the HTTP status code.

```go
activity, err := client.GetActivity(ctx, "unknown-id")
if err != nil {
    var apiErr *hammerhead.APIError
    if errors.As(err, &apiErr) {
        fmt.Println("Status:", apiErr.StatusCode) // e.g. 404
        fmt.Println("Message:", apiErr.Message)
    }
    log.Fatal(err)
}
```

## Available scopes

| Constant | Scope | Description |
|---|---|---|
| `ScopeActivityRead` | `activity:read` | Read activities and receive webhook notifications |
| `ScopeRouteRead` | `route:read` | Read routes |
| `ScopeRouteWrite` | `route:write` | Create and update routes |
| `ScopeWorkoutWrite` | `workout:write` | Create and update workouts |

## Activity types

`ActivityTypeRide`, `ActivityTypeEBike`, `ActivityTypeMountainBike`, `ActivityTypeGravel`, `ActivityTypeEMountainBike`, `ActivityTypeVelomobile`

## Deauthorizing a user

```go
err := client.Deauthorize(ctx, "your-client-id", "your-client-secret", accessToken)
```

This removes the user's account link, deletes all imported routes and workouts, and revokes their tokens.

## Advanced configuration

```go
// Use a custom HTTP client (e.g. with a timeout)
hc := &http.Client{Timeout: 30 * time.Second}
client, err := hammerhead.NewClient("your-access-token", hammerhead.WithHTTPClient(hc))
if err != nil {
    log.Fatal(err)
}

// Override base URLs (useful for testing)
client, err = hammerhead.NewClient("token",
    hammerhead.WithBaseURL("http://localhost:8080"),
    hammerhead.WithAuthURL("http://localhost:8080"),
)
if err != nil {
    log.Fatal(err)
}
```

## License

MIT
