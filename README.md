[![Build Status](https://travis-ci.org/yonush/pwrcost.svg?branch=master)](https://travis-ci.org/yonush/pwrcost)

# Introduction

This aim of this project is to create a basic notes application - based off [pwrcost](https://github.com/yonush/pwrcost). It uses mostly built-in Go's packages such as `net/http`, `database/sql`, `strconv`, `html/template` and uses third party packages `gorilla/mux` for the router, `jackc/pgx` for the PostegreSQL driver and `icza/session ` for session manangement. 

This application was created to demonstrate the following:</p>

-   Go HTML templates with [W3.CSS](https://www.w3schools.com/w3css/w3css_examples.asp) stylesheet
-   RESTful API with gorilla/mux
-   Datastorage using PostgreSQL - tested with 15.4.1
-   Session mannagement with icza/session 
-   Create, Read, Update and Delete Notes/Tasks
-   Searching for Notes/Tasks based on attributes - Validation for Text Patterns from assignment brief was attempted, but not executed in application. 
-   Share function for Notes/Tasks
-   Count occurrences of specified text snippets within a single Note/Task - Validation for Text Patterns from assignment brief was attempted, but not executed in application. 

The application is standalone requiring no additional WAMP/LAMP dependencies. It has been built and tested on Windows. If the application requires rebuilding, there are two batch files provided. The rebuild assumes there is an existing Go installation on the host computer. The provided batch files will download the required 3rd party packages for the build process.

## Building

This application uses the Go programming language - where the latest was [Go 1.21](https://go.dev/dl/) as of writing this application. If you do not have Go installed on your system, you can acquire a copy from [Go.dev](https://go.dev/dl/). The go1.21.0.windows-amd64.msi was used to build this application.

### Database configuration

The app assumes a database exists - ESD. Edit the _dbsetup.go_ to change the default database name. Database defaults in the _dbsetup.go_ are shown below.

```go
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
	dbname   = "postgres"
)
```

To run the server on your Windows system:

1. Run `buildpkg.cmd` in the root of the repo to build the binary (`notes.exe`) using non vendored packages
1. Run `buildvendor.cmd` in the root of the repo to build the binary (`notes.exe`) with the vendor
1. Run the binary `notes.exe` or used the run.cmd (has env variable set)
1. Browse to [http://localhost:8080](http://localhost:8080) to test the application out. If port 8080 does not work, you can start the app as follows using your console, where \*\*\*\* is your chosen port number.

    ```
        > ./notes.exe ****
    ```

### Non Windows

Testing has been performed on WSL & Linux but not MacOS. However, the commands in buildpkg.cmd and buildvendor.cmd can be run manually to build and run this demo.

#### Build by pkg

```bash
export GO111MODULE="on"
export GOFLAGS="-mod=mod"
go mod download
:: strip debug info during build
go build -ldflags="-s -w" .

```

#### Build by vendor

```bash
export GO111MODULE="on"
export GOFLAGS="-mod=vendor"
go mod vendor
:: strip debug info during build
go build -ldflags="-s -w"
```

### Dependencies

The application uses the following Go packages to build;

-   [Datastore: PostgreSQL driver](https://github.com/jackc/pgx/)
-   [HTTP router: Gorilla mux](https://github.com/gorilla/mux)

## Datastore

This version application requires a separate database to function - PostgreSQL. A few CSV files are imported from the local data folder. This will be imported when the application is run for the first time. Thereafter the application will use the database each time it is executed.

## Sample screens

![Creating](statics/images/create.png "create")
![Creating](statics/images/list.png "create")

## Session management

The application uses the [icza/session](https://github.com/icza/session) module to handle some basic sessions for the authentication.

Alternate session management be found at the following sources:

-   [swithek/sessionup](https://github.com/swithek/sessionup) Simple, yet effective HTTP session management and identification package
-   [gorilla/sessions](https://github.com/gorilla/sessions) - provides cookie and filesystem sessions and infrastructure for custom session backends along with the gorilla mux
-   [alexedwards/scs](https://github.com/alexedwards/scs) - session management with over 19 different datastores
-   [](https://github.com/adam-hanna/sessions) - A dead simple, highly performant, highly customizable sessions service for go http servers using redis as a datastore.
