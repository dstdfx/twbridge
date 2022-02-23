# twbridge

This repository contains an implementation of Telegram <-> Whatsapp bridge.  
It allows you to receive incoming Whatsapp messages in Telegram chat and reply to them.

## How it works

All text messages that you receive in Whatsapp chats are being forwarded to Telegram chat with the bot.  
Incoming text messages have the following format:
```text
From: Test User [jid: testuser@gmail.com]
= = = = = = = = = = = =
Message: hello, world!
```
Reply to a message can be done by simply [replying](https://telegram.org/blog/replies-mentions-hashtags#replies) to a specific message.

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

Run locally:

```bash
export TELEGRAM_API_TOKEN=<your-telegram-bot-token>; ./twbridge
```

Run in docker:

```bash
docker run -d --env TELEGRAM_API_TOKEN="<YOUR-TELEGRAM_API_TOKEN>" ghcr.io/dstdfx/twbridge:latest
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

