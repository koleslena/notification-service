# notification-service

## Configuration

In configs, copy example and rename it into `config.yaml`
Environments variables are in app.env file

## Local build & run

Run migration script for db
```bash
go build -o migrate ./migrate/migrate.go # build
./migrate/migrate # run 
```
Setup url and token for sending service in app.env file
Run notification service
```bash
go build -o ./main ./main.go # build
./main # run 
``` 

### Docker-compose setup

There is an example `docker-compose.yaml` file to provide an overall idea on how to run the service alongside with the other infrastructure components.
