# Using docker compose

D2 chatbot uses MYSQL to store subscriber information and specifically `docker-compose` to do this.

## Starting mysql

```bash
# Start all docker containers specified as daemons.
$ docker-compose up -d

# Make sure everything started correctly
$ docker-compose ps

# Check logs if they didn't start
$ docker-compose logs -f
```

## Using mysql

Exec into the container running mysql, connecting the root with the password specified in the `docker-compose.yml`.

```bash
# Exec container
$ docker exec -it docker_mysql_1 sh

# Use mysql (enter password when prompted)
$ mysql -u root -p
```
