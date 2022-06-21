
# Zebrahook

Delegate your webhook needs by using Zebrahook.

Zabrahook is a simple but complete system to fully handle webhooks, inspired by Stripe Webhooks, it depends only on PostgreSQL.


## Features
- **Easy to use**
  - Register webhook endpoints and send an event via REST API, depends only on PostgreSQL.
- **Event type wildcard support**
  - Allows to register a webhook endpoint that is subscribed to a subsets of events (e.g. `enabled_events: ["charges.*", "shipment.warehouse.A3001.*"]`)
- **Automatic backoff retries and circuit breaker**
  - If a webhook endpoint is not responding Zebrahook will automatically retry up to 3 times using an exponential backoff strategy (configurable)
- **Configurable**
  - Using a json config file or via CLI arguments you can customize request headers, backoff strategy and many more options
- **Secure**
  - Each registered webhook endpoints have a different secret key, event content is signed (HMAC SHA256) and timestamped to prevent replay attacks.

Integrate the REST API in your backend using the [**OpenAPI Spec**](https://generator3.swagger.io/index.html?url=https://raw.githubusercontent.com/nya1/zebrahook/main/gen/http/openapi3.yaml)


## Architecture

HTTP Server
  - This is the REST API that you will call to register new webhook endpoints and submit events


Worker, two different types of workers are used to better scale and delivery events
  - Event mapping worker
    - Used to map events to webhook endpoints

  - Dispatcher worker
    - This is the worker that will perform the HTTP request to the webhook endpoint and delivery the event

Checkout the [sequence diagram](./docs/architecture.md) to learn more

## CLI

You can easily start zebrahook using docker:

```bash
docker run -v $PWD/config.json:/config.json -p 3000:3000 -t ghcr.io/nya1/zebrahook --server
```

### Start the server

```bash
zebrahook --server
```

#### Additional Options

- `--http-port <port>`: HTTP port to listen

### Start worker

```bash
zebrahook --worker <type>
```

**NOTE:** `<type>` must be `eventMapping` or `dispatcher`


### Other flags

#### Setup database

Zebrahook needs a database with some tables and indexes, to setup your database you can run the following command:

```bash
zebrahook --setup
```

#### Create new API Key

Allows to create a new API key used to interact with Zebrahook, this is the only api key that you will need and it's considered an "admin" api key as it allows to register endpoints and send events.

```bash
zebrahook --new-api-key
```

Clear text api key will be printed on screen


## Configuration

A json file called `config.json` is needed as a configuration. An example configuration can be found [here](./config.example.json)


| JSON Field                                  | CLI Override  | Type    | Required | Default | Description                                                         |
|---------------------------------------------|---------------|---------|----------|---------|---------------------------------------------------------------------|
| `encryptionKey`                             | n/a           | string  | yes      |         | used internally to encrypt webhook secrets into the database    |
| `database.dsn`                              | n/a           | string  | yes      |         | full connection url to PostgreSQL                                   |
| `webhookRequest.timeoutSecs` | n/a           | number  | no       | 30       | maximum HTTP timeout in seconds         |
| `webhookRequest.userAgent` | n/a           | string  | no       | Zebrahook       | User-Agent header value     |
| `webhookRequest.signatureHeaderName` | n/a           | string  | no       | Zebrahook-Signature       | Name of the header that will contain the signature     |
| `logger.level`                              | `--log-level` | string  | no       | info    | log level, available values: debug, info, warn, error, fatal, panic |
| `logger.output.json`                        | `--log-json`  | boolean | no       | false   | if true output log as a json                                        |
| `backoffStrategy.baseSecs` | n/a           | number  | no       | 60       | the minimum seconds used in calculation of the exponential backoff (formula used: `baseSecs**nextAttemptCounter+random(0.0,1.0)`)                                 |
| `backoffStrategy.maxAttempts` | n/a           | number  | no       | 3       | maximum number of event delivery attempts                                 |
| `worker.eventMapping.parallelJob`           | n/a           | number  | no       | 1       | how many queue polling jobs to run in parallel                      |
| `worker.eventMapping.pollingIntervalSecs.min` | n/a           | number  | no       | 0.5     | minimum polling interval in seconds                                 |
| `worker.eventMapping.pollingIntervalSecs.max` | n/a           | number  | no       | 2       | maximum polling interval in seconds                                 |
| `worker.dispatcher.parallelJob`             | n/a           | number  | no       | 2       | how many queue polling jobs to run in parallel                      |
| `worker.dispatcher.pollingIntervalSecs.min` | n/a           | number  | no       | 0.5     | minimum polling interval in seconds                                 |
| `worker.dispatcher.pollingIntervalSecs.max` | n/a           | number  | no       | 2       | maximum polling interval in seconds                                 |


## Acknowledgements

- Built using [goa](https://github.com/goadesign/goa)
- For PostgreSQL queue using [pgq go module](https://github.com/btubbs/pgq)
- Crypto utils from [cryptopasta](https://github.com/gtank/cryptopasta)
