# go andrewwillette.com
My personal website which hosts my CV and music recordings.

## Installation and startup
* Ensure `go` is available in your `$PATH`.
* Execute `go install`
* Execute `go build .`
* Execute `./willette_api`

## Database management
The rest api uses a sqlite database for persistence of users and music data.

The database location is conditional. If the directory `/goApp/db` exists, it is used as the location for the sqlite database. This directory is created in the Docker image, and is used to persist the sqlite database between docker deployments. The below example assumes `~/db` is the desired location for the database on the hostOS.

```bash
docker run -p 9099:9099 -d -v ~/db:/goApp/db -e NEW_RELIC_LICENSE=<nrLicense> imagename
```
