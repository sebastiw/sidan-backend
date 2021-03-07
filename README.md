# sidan-backend

## How to run

`make`: start database locally

`go run src/sidan-backend.go`: start service locally

The service should connect automatically to the local database. To
change configfile (from the default `config/local.yaml`), you can set
the `CONFIG_FILE` env parameter pointing to an the new config-file.

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

- [ ] /auth/ - for authorization and forgot password etc.
- [ ] /file/ - for image uploads
- [ ] /mail/ - for mailing the reset password
- [ ] /notify/ - for sending notifications
- [ ] /db/ - all database queries and calls

- PUT for creation
- POST for modification
- DELETE for removal
- GET for querying
- (optional) HEAD on API should give back a header with a description string.
- (optional) GET on base URIs should give back a documented list on available APIs.
