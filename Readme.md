# Employee Manager
This is a simple employee manager application that allows you to add, delete, and update employees. It is using golang(1.22.3) and a postgres database. 

## Installation
1. Clone the repository
2. Run `go mod tidy` to install the dependencies
3. Set the following environment variables:
    - DB_URL: The url to your postgres database
    - PORT: The port you want the server to run on
4. Run `make run` to start the server
5. The server should be running on the port you specified. For example, if you set the port to 8080, you can access the server at `http://localhost:8080/swagger/index.html`

## Documentation
We use swag to generate the documentation. Run `make gen-swag` to generate the documentation.

## Testing
Run `make test` to run the tests

## Snapshot Testing
We use `cupaloy` to do snapshot testing. Run `make update-snapshot` to update the snapshots.
