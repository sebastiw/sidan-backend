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

Check `swagger.json` for full API.

## /auth/{provider}

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

## /repo/fdroid/ — F-Droid app repository

The backend hosts a private [F-Droid](https://f-droid.org)-compatible repository at
`/repo/fdroid/`. Add it to the F-Droid client with the URL:
`https://api.chalmerslosers.com/repo/fdroid`

### Uploading an app for the first time

```sh
TOKEN="<your Bearer token>"
curl -X POST https://api.chalmerslosers.com/repo/fdroid/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "apk=@MyApp-release.apk" \
  -F "name=My App" \
  -F "summary=One line description" \
  -F "description=Full description of the app" \
  -F "license=MIT" \
  -F "source_code=https://github.com/yourorg/myapp" \
  -F "categories=Utilities"
```

The supplementary fields (`name`, `summary`, `description`, `license`, `source_code`,
`categories`) are optional — package name, version, SDK levels, and permissions are all
parsed automatically from the APK.

### Uploading a new version

Exactly the same command with the new APK file. The new version is appended to the index
alongside the old one. The F-Droid client will offer an update to users who have the app
installed. The supplementary metadata fields only need to be included if you want to update
them; otherwise they are taken from the previous upload of that package.

### Prerequisites

The server must have a Java keystore at the path configured in `fdroid.keystorePath`. Generate
it once and keep it safe — changing the key invalidates the repo for anyone who has already
added it:

```sh
keytool -genkey -v -keystore fdroid.keystore -alias fdroid \
  -keyalg RSA -keysize 4096 -validity 10000
```

`jarsigner` (from any JDK) must be on the server's PATH. The keystore file must not be
committed to git

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
