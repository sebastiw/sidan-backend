# sidan-backend

## Old APIs

http://chalmerslosers.com:9000/service-inspector/Clac3

Download wsdl to `apis/Clac3.xml`, then run `node wsdl_to_swagger.js`
to generate the swagger yaml.


# Non-functional Requirements

- [ ] Language: Golang
- [ ] Test locally
- [ ] Tests for APIs
- [ ] Docker + env variables for choosing local/test/prod configs
- [ ] Other config in yaml files

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
