# Flight Tracker API
#### A simple flight tracking application built with Go, Gin, and Redis.

## Requirements:
- Docker
- Make (This comes installed with most Linux distributions and macOS versions)

## Walkthrough:
### JWT Auth
This app has JWT authentication. You can use the token obtained from the login endpoint to access protected routes. For example, to get a secret message, you can use the following command:

```bash
   curl -X GET http://localhost/secret/ \
    -H "Authorization: Bearer $JWT_TOKEN"
```

This shouldn't work if you don't have the JWT_TOKEN set. You can set it by running the login command below.

### Login
```bash
   export JWT_TOKEN=$(curl -s -X POST http://localhost/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin", "password":"admin"}' | jq -r '.token')
```

In this script you query the `/login` endpoint with the username and password of the admin user. The response is parsed using `jq` to extract the token, which is then stored in the `JWT_TOKEN` environment variable. This token can be used for authentication in subsequent requests to the API.


### JWT Auth Rematch

```bash
   curl -X GET http://localhost/secret/ \
    -H "Authorization: Bearer $JWT_TOKEN"
```

Now you can use the token to access the secret endpoint. The server will verify the token and return the secret message if the token is valid.

### Get all flights
```bash
    curl "http://localhost/api/flights" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $JWT_TOKEN"
```
This will return a list of all the cached flights in the system.

