# sidan-backend

## How to run

`make`: start database locally

`go run src/sidan-backend.go`: start service locally

The service should connect automatically to the local database. To
change configfile (from the default `config/local.yaml`), you can set
the `CONFIG_FILE` env parameter pointing to an the new config-file.

## API Documentation

**Swagger UI**: http://localhost:8080/swagger/index.html

**Get JWT Token**: http://localhost:8080/swagger-auth (login with Google OAuth2)

**Generate Swagger docs**: `make swagger` (regenerates from code annotations)

### How to Use Swagger UI:
1. Visit `/swagger-auth` to login with Google
2. Copy the JWT token that appears
3. Go to `/swagger/index.html` 
4. Click "Authorize" button (top right)
5. Paste token and click "Authorize"
6. Try API endpoints!

Note: Auth endpoints (`/auth/*`) are excluded from Swagger as they implement the OAuth2 flow used by `/swagger-auth`

## Structure

under `/`:
- `db/`: database migrations
- `static/`: static files served by (for instance) `file` endpoint.
- `config/`: different configs for different cases.
- `apis/`: files defining old and new apis.
- code under `src/`:
  - `src/router/`: defines the endpoints
  - `src/database/`: defines models and database apis.
  - `src/config/`: define and parses the config

## Old APIs

http://chalmerslosers.com:9000/service-inspector/Clac3

Download wsdl to `apis/Clac3.xml`, then run `node wsdl_to_swagger.js`
to generate the swagger yaml.


# Non-functional Requirements

- [X] Language: Golang
- [ ] Test locally
- [ ] Tests for APIs
- [x] Docker + env variables for choosing local/test/prod configs
- [x] Other config in yaml files

- [ ] (optional) New tables and migrate data.

# Functional requirements

- [ ] Content-Type: application/json

- [x] /auth/ - for authorization and forgot password etc.
- [x] /file/ - for image uploads
- [ ] /mail/ - for mailing the reset password
- [ ] /notify/ - for sending notifications
- [ ] /db/ - all database queries and calls

- GET for querying
- PUT for creation
- POST for modification
- DELETE for removal
- (optional) HEAD on API should give back a header with a description string.
- (optional) GET on base URIs should give back a documented list on available APIs.

# API

**Interactive documentation**: Navigate to `/swagger/index.html` when the server is running.

All endpoints except `/auth/*` are auto-documented from code annotations. See `docs/AUTH.md` for authentication flow.

## /auth/

For oAuth2 flows, please see
https://developers.google.com/oauthplayground/
https://auth0.com/blog/backend-for-frontend-pattern-with-auth0-and-dotnet/
https://stackoverflow.com/a/77469099

## /db/

These functions operate on the database.

Check under `src/database/models/` to see what format the JSON request
should have.

## /mail/

### PUT /mail

    Send new mail.

    JSON payload structure:

    {"from_email": "", "to_emails": [], "message": "", "title": ""}

## /file/

### GET /file/RESOURCE

    Fetch a file resource from the `/static/` directory.

### PUT /file/image

    Upload a new image file via form data.

    Supported image types:
    - gif
    - png
    - jpeg

    Max size 10 MB

    curl -F ‘data=@path/to/local/file' -X put
