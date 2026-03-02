# OOT Rando Explorer

This repo contains a web app for self-hosting to share Randomizer seeds from Ship of Harkinian with your friends.

## Running

TODO: dockerfile registry and example docker compose file

## Building

1. **Prerequisites**
   - Go 1.24 or later
   - TailwindCSS CLI, either from NPM or standalone
   - Docker (optional)

2. **Build Locally**

```sh
git clone --recursive https://github.com/bsinky/oot-rando-explorer.git
cd oot-rando-explorer
# To compile the binary
go build -o sohrando .
# Then build the CSS, assuming use of the standalone tailwind CLI
./tailwindcss -i ./assets/input.css -o ./assets/index.css
```

3. **Build with Docker**

A multi-stage Dockerfile is provided that compiles assets and the Go binary into a convenient Docker image.

```sh
make deploy
# or
docker build -t oot-rando-explorer:latest .
```

## Running Tests

You can run the unit tests with:

```sh
go test ./...
```

## Environment Variables

The behavior of the application can be modified using environment variables:

- `PORT` - the port the web server should run on, defaults to 8080
- `OOTRANDODB` - which type of database to use, set this to either `sqlite` or `postgres`
- `OOTRANDODBURI` – the URI used to open the database. If not set, the application defaults to `sqlite.db`.
  Examples:
  `sqlite.db` (default), `postgres://user:pass@host/dbname?sslmode=disable`.
- `OOTRANDOSESSIONSECRET` - the string to use when encrypting user sessions, this should be set to a unique password

## Command-Line Options

When the `sohrando` binary is executed with arguments, instead of running the web server it enters command mode and performs one of the following administrative operations before exiting:

```text
usage: sohrando <command> [args]
commands:
  make-admin <username>
  make-moderator <username>
  reset-password <username> <new_password>
```

These mainly exist to run commands against the database when running within a container, like so:

```text
docker compose exec <container-name> sohrando make-admin youruser
```

## Running the Server

To start the web server, just run the `sohrando` binary:

```sh
./sohrando
```

It will listen on port 8080 by default.

## Updating to add new Ship of Harkinian versions

```sh
# First ensure the Shipwright submodule is up to date
git submodule update --remote
# Then run the script
./get-soh-versions.sh
```

If any new versions were found, it will update the list in `randoseed/versions.txt` for you to review and commit the changes if everything seems correct. For a lot of Ship updates, this is likely all that is needed.

If the JSON format Ship uses has changed significantly, more code changes will be needed to support the new spoilerlog format.