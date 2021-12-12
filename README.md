# twbridge

This repository contains an implementation of Telegram <-> Whatsapp bridge.  
It allows you to receive incoming Whatsapp messages in Telegram chat and reply to them.

## Requirements 

To run this app you need a Telegram bot created, check [this manual](https://core.telegram.org/bots#3-how-do-i-create-a-bot)
if you've never done it before.  
Once you managed to create it you'll have an API token which will be used to run this app.

## Build

Use the following command to build a binary:

```bash
make build
```

Use the following command to build a Docker image:

```bash
docker build -t twbridge .
```

## Running

```bash
export TELEGRAM_API_TOKEN=<your-telegram-bot-token>; ./twbridge
```

## Testing

Use the following command to run unit-tests and linters:

```sh
make tests
```

Use the following command to run unit-tests only:

```sh
make unittests
```

Use the following command to run golangci-lint:

```sh
make golangci-lint
```

## TODO
