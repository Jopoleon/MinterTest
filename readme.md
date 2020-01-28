## Minter test task.

Text of the task in `task.txt`

### Building

`$ go build .`

### Start in Docker 

`$ docker-compose up -d --build`

### Start on localhost

You will need postgres nod to start application 

`$ go run .`

###Configuration

Use `.env` file to configure various environment variables 

```.env
DB_HOST=dbpostgres # uncomment for docker-compose usage
#DB_HOST=localhost # for local db
PARSING_WORKERS_AMOUNT=4 #amount of concurrent parsing workers
DB_DRIVER=postgres
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=postgres
DB_PORT=5432
```