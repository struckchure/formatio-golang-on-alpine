# formatio-golang-on-alpine

## PostgreSQL

Build the database image with non-secret database settings:

```sh
docker build -f Dockerfile.postgres \
	--build-arg POSTGRES_USER=app \
	--build-arg POSTGRES_DB=app \
	--build-arg POSTGRES_PASSWORD=change-me \
	-t formatio-postgres .
```

For production, pass the password at runtime through an environment file
instead. Build arguments can be retained in image history:

```sh
printf 'POSTGRES_PASSWORD=%s\n' 'change-me' > .env.postgres
docker run --env-file .env.postgres \
	-p 5432:5432 \
	--name formatio-postgres \
	formatio-postgres
```

The Go app accepts `DATABASE_URL`, or builds a connection from
`POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, and
`POSTGRES_DB`. It runs `schema.sql` automatically on startup.

Posts are managed through these endpoints. Request bodies use the following
shape: `{"title":"A title","content":"Post body"}`.

```text
POST   /posts
GET    /posts
GET    /posts/:id
PUT    /posts/:id
DELETE /posts/:id
```
