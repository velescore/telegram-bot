[![Release](https://img.shields.io/github/release/velescore/telegram-bot.svg?branch=master)](https://github.com/velescore/telegram-bot/releases) 
[![Build Status](https://travis-ci.org/velescore/telegram-bot.svg?branch=master)](https://travis-ci.org/velescore/telegram-bot) 
[![go-doc](https://godoc.org/github.com/velescore/telegram-bot?status.svg)](https://godoc.org/github.com/velescore/telegram-bot) 
[![Go Report Card](https://goreportcard.com/badge/github.com/velescore/telegram-bot)](https://goreportcard.com/report/github.com/velescore/telegram-bot) 
[![Followers](https://img.shields.io/twitter/follow/velescore.svg?style=social&label=Follow)](https://twitter.com/velescore)

# Velescore telegram bot

## Commands

```
	      	/h or /help 	  display help message
		/p <symbol> 	  info about coin price
		/s <symbol> 	  info about supply
		/c <symbol> 	  info about price change
		/a <symbol>	  info about ATH
```   

## Telegram address 
https://t.me/Velescore
## Binary releases
https://github.com/velescore/telegram-bot/releases

## Building project from source

```
git clone git@github.com:velesscore/telegram-bot.git
cd telegram-bot/
make 
```

## Running bot
Basic usage: ```./telegram-bot run -t "telegram_bot_api_key"```

Where telegram_bot_api_key can be generated as described https://core.telegram.org/bots#creating-a-new-bot 


Additional parameters are described in help section:
```./telegram-bot run --help```

By default [/metrics](http://localhost:9900/metrics) endopoint is launched which is compatibile with https://prometheus.io/

## Version checking
```./telegram-bot version```
