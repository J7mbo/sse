SSE
---

An example Server-Sent events server.

Reads from RabbitMQ and sends messages to the corresponding /topic/userid/nonce combination.

Uses a middleware to check the user + nonce combination exists in a postgres db.

Subscribe with `EventSource` in JS to `/subscribe/topic/userid/nonce`, where all three params are v4 uuids. Then place a
 payload of type `FoundContract` (this is my implementation) in RabbitMQ, and it will be published. Check [internal
 /rabitmq/message](internal/rabbitmq/message) for the required message payload.
 
You will need to set up RabbitMQ and Postgres yourself as they are not a part of this service.

Of interest
--

- [Go](https://imgflip.com/memetemplate/You-Dont-Say), implicit interfaces, channels, some "idiomatic go" in case
 that's your thing
- TLS everywhere, make, ansible, docker, docker-compose
- Logfiles are output in a format that can be parsed by Logstash / Filebeat etc (for Kibana)
- Mockery, testify etc

Prerequisites
--

- [Docker, docker-compose](https://docs.docker.com/get-docker/).
- [Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html), and `ansible-playbook` on
 your host machine.
- RabbitMQ running somewhere that you can connect to (configure in [#Configuration](#Configuration)), with TLS.
- [Mockery](https://github.com/vektra/mockery) if you want to generate mocks with `make mock`
- [Patience](https://en.wikipedia.org/wiki/Patience). Ensure RabbitMQ is on the same docker network ;)

Usage
--

Run: `make` to see your options.

You'll need to place the user in the users db table first, which is just a list of users that are allowed to connect. 
Schema needs to have `uuid` and `sse_token` strings. Save a new user into your `users` table with these values to allow 
a connection. Alternatively, just remove the call to postgres in the middleware wont need a database.

```postgresql
CREATE DATABASE app;
CREATE TABLE users (uuid varchar(36), sse_token varchar(36));
INSERT INTO users (uuid, sse_token) VALUES('bb6819a5-c346-4924-ab09-d48b47fdf087', 'bb6819a5-c346-4924-ab09-d48b47fdf087');
```

Add the following JavaScript to connect:

```javascript
let topic = "some-topic";
let uuid = "bb6819a5-c346-4924-ab09-d48b47fdf087";
let token = "bb6819a5-c346-4924-ab09-d48b47fdf087"

let url = `https://localhost:8001/subscribe/${topic}/${uuid}/${token}`;

let es1 = new EventSource(url);
es1.onmessage = function (event) {
    console.log(event.data);
};
```

Place the following message in rabbitmq, and it will appear in the front-end console:

```json
{
    "userid":  "bb6819a5-c346-4924-ab09-d48b47fdf087", 
    "filename":  "some data here", 
    "filepath":  "some data here",
    "finished": false, 
    "type": "foundcontract"
}
```

The general flow I'm using is a custom one. In my implementation, the service which provides the nonce to the front-end
also shares its database with the SSE server. This makes SSE less of a microservice. However, the SoC is still there and
the service can still be scaled, albeit around the Postgres dependency which can also be scaled.

Configuration
--

See [infrastructure/dev.env](infrastructure/dev.env). If you don't pass these environment variables then they default to
those in [internal/config.go](internal/config.go), allowing two local environments: locally and in a docker container.

TLS
--

Configure the certificates directory under `SSE_CERTS_DIR` in `dev.env`.
Ensure the following files exist and are readable:

For the SSE server:
- `${SSE_CERTS_DIR}/cert.pem`
- `${SSE_CERTS_DIR}/key.pem`
- `${SSE_CERTS_DIR}/ca.pem`

For the connection to Rabbitmq:
- `${SSE_CERTS_DIR}/server_certificate.pem`
- `${SSE_CERTS_DIR}/server_key.pem`
- `${SSE_CERTS_DIR}/ca_certificate.pem`

For the connection to Postgres:
- `${SSE_CERTS_DIR}/postgresca.crt`

Ansible takes care of the SSE certs for you, but you need to retrieve the postgres and rabbitmq certs yourself.
If you're using mkcert to generate the RabbitMQ certs, you need the `-client` parameter.