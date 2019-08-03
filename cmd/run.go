// Copyright © 2018 coinpaprika.com
// Copyright @ 2019 Veles Core 
//
// Licensed under the Apache License, version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"net/http"

	"github.com/coinpaprika/coinpaprika-api-go-client/coinpaprika"
	"github.com/velescore/telegram-bot/telegram"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	debug   bool
	token   string
	metrics int

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run coinpaprika bot",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}

	commandsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "velescore",
		Subsystem: "telegram_bot",
		Name:      "commands_proccessed",
		Help:      "The total number of processed commands",
	})
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&debug, "debug", "d", false, "enable debugging messages")
	runCmd.Flags().StringVarP(&token, "token", "t", "", "telegram API token")
	runCmd.Flags().IntVarP(&metrics, "metrics", "m", 9900, "metrics port (default :9900) endpoint: /metrics")
	runCmd.MarkFlagRequired("token")

	prometheus.MustRegister(commandsProcessed)
}

func run() error {
	log.SetLevel(log.ErrorLevel)
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	log.Debug("starting telegram-bot")

	bot, err := telegram.NewBot(telegram.BotConfig{
		Token:          token,
		Debug:          debug,
		UpdatesTimeout: 60,
	})

	if err != nil {
		return err
	}

	updates, err := bot.GetUpdatesChannel()
	if err != nil {
		return err
	}
	go func(updates tgbotapi.UpdatesChannel) {
		for u := range updates {
			log.Debugf("Got message: %v", u)

			if u.Message == nil || !u.Message.IsCommand() {
				log.Debug("Received non-message or non-command")
				continue
			}
			commandsProcessed.Inc()

			text := `Please use one of the commands:

			/h or /help 	  display help message
			/p <symbol> 		info about coin price
			/s <symbol> 		info about supply
			/c <symbol> 		info about price change
			/a <symbol>			info about ATH

			`
			log.Debugf("received command: %s", u.Message.Command())
			switch u.Message.Command() {
			/*
			case "author":
				text = "https://github.com/coinpaprika/telegram-bot"
			case "source":
				text = "https://github.com/velescore/telegram-bot"
			*/
			case "p":
				if text, err = commandPrice(u.Message.CommandArguments()); err != nil {
					text = "invalid coin name|ticker|symbol, please try again"
					log.Error(err)
				}
			case "s":
				if text, err = commandSupply(u.Message.CommandArguments()); err != nil {
					text = "invalid coin name|ticker|symbol, please try again"
					log.Error(err)
				}
			/*case "v":
				if text, err = commandVolume(u.Message.CommandArguments()); err != nil {
					text = "invalid coin name|ticker|symbol, please try again"
					log.Error(err)
				}
			case "m":
				if text, err = commandMarketCap(u.Message.CommandArguments()); err != nil {
					text = "invalid coin name|ticker|symbol, please try again"
					log.Error(err)
				}*/
			case "a":
				if text, err = commandAthPrice(u.Message.CommandArguments()); err != nil {
					text = "invalid coin name|ticker|symbol, please try again"
					log.Error(err)
				}
			case "c":
				if text, err = commandPriceChange(u.Message.CommandArguments()); err != nil {
					text = "invalid coin name|ticker|symbol, please try again"
					log.Error(err)
				}
			}

			err := bot.SendMessage(telegram.Message{
				ChatID:    int(u.Message.Chat.ID),
				Text:      text,
				MessageID: u.Message.MessageID,
			})

			if err != nil {
				log.Error(err)
			}
		}

	}(updates)

	log.Debugf("launching metrics endpoints :%d/metrics", metrics)
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(fmt.Sprintf(":%d", metrics), http.DefaultServeMux)
}

func commandPrice(argument string) (string, error) {
	log.Debugf("processing command /p with argument :%s", argument)

	ticker, err := getTickerByQuery(argument)
	if err != nil {
		return "", errors.Wrap(err, "command /p")
	}

	priceUSD := ticker.Quotes["USD"].Price
	priceBTC := ticker.Quotes["BTC"].Price
	volumeUSD := ticker.Quotes["USD"].Volume24h
	marketCapUSD := ticker.Quotes["USD"].MarketCap
	marketCapBTC := ticker.Quotes["BTC"].MarketCap
	if ticker.Name == nil || ticker.ID == nil || priceUSD == nil || priceBTC == nil || volumeUSD == nil || marketCapUSD == nil || marketCapBTC == nil {
		return "", errors.Wrap(errors.New("missing data"), "command /p")
	}

	return fmt.Sprintf(`%s price:
		 %.4f $
		 %.8f ₿
	%s marketcap:
		 %.f $
		 %.f ₿
	%s volume:
		 %.f $
  http://coinpaprika.com/coin/%s`,
		*ticker.Name, *priceUSD, *priceBTC, *ticker.Symbol, *marketCapUSD, *marketCapBTC, *ticker.Symbol, *volumeUSD, *ticker.ID), nil
}
/*
func commandMarketCap(argument string) (string, error) {
	log.Debugf("processing command /m with argument :%s", argument)

	ticker, err := getTickerByQuery(argument)
	if err != nil {
		return "", errors.Wrap(err, "command /m")
	}

	marketCapUSD := ticker.Quotes["USD"].MarketCap
	marketCapBTC := ticker.Quotes["BTC"].MarketCap
	if ticker.Name == nil || marketCapUSD == nil || marketCapBTC == nil {
		return "", errors.Wrap(errors.New("missing data"), "command /m")
	}

	return fmt.Sprintf("%s marketcap information \n %.2f USD \n %.2f BTC", *ticker.Name, *marketCapUSD, *marketCapBTC), nil
}
*/
func commandAthPrice(argument string) (string, error) {
	log.Debugf("processing command /a with argument :%s", argument)

	ticker, err := getTickerByQuery(argument)
	if err != nil {
		return "", errors.Wrap(err, "command /a")
	}

	athUSD := ticker.Quotes["USD"].ATHPrice
	athBTC := ticker.Quotes["BTC"].ATHPrice
	downFromAth := ticker.Quotes["USD"].PercentFromPriceATH
	athDate := ticker.Quotes["USD"].ATHDate
	if ticker.Name == nil || athDate == nil || athUSD == nil || athBTC == nil || downFromAth == nil {
		return "", errors.Wrap(errors.New("missing data"), "command /a")
	}

	return fmt.Sprintf(`%s ATH info:
		 %.4f $
		 %.8f ₿
		 %s
		 Down since ATH %.2f %%`,
		*ticker.Name, *athUSD, *athBTC, *athDate, *downFromAth), nil
}

func commandSupply(argument string) (string, error) {
	log.Debugf("processing command /s with argument :%s", argument)

	ticker, err := getTickerByQuery(argument)
	if err != nil {
		return "", errors.Wrap(err, "command /s")
	}

	if ticker.Name == nil || ticker.MaxSupply == nil || ticker.TotalSupply == nil  || ticker.CirculatingSupply == nil || ticker.Symbol == nil {
		return "", errors.Wrap(errors.New("missing data"), "command /s")
	}

	return fmt.Sprintf(`%s supply info:
		max supply: %d %s
		total supply: %d %s
		circ. supply: %d %s`,
		*ticker.Name, *ticker.MaxSupply, *ticker.Symbol, *ticker.TotalSupply, *ticker.Symbol, *ticker.CirculatingSupply, *ticker.Symbol), nil
}
/*
func commandMarkets(argument string) (string, error) {
	log.Debugf("processing command /e with argument :%s", argument)

	ticker, err := getTickerByQuery(argument)
	market = string
	if err != nil {
		return "", errors.Wrap(err, "command /e")
	}

	if ticker.Name == nil || ticker.ID == nil {
		return "", errors.Wrap(errors.New("missing data"), "command /e")
	}

	return fmt.Sprintf("%s is trading on: %d \n\n http://coinpaprika.com/coin/%s", market*ExchangeName, market*Pair, market*ReportedVolume24hShare ), nil
}
*/
func commandPriceChange(argument string) (string, error) {
	log.Debugf("processing command /c with argument :%s", argument)

	ticker, err := getTickerByQuery(argument)
	if err != nil {
		return "", errors.Wrap(err, "command /c")
	}

	priceChange1h := ticker.Quotes["USD"].PercentChange1h
	priceChange12h := ticker.Quotes["USD"].PercentChange12h
	priceChange24h := ticker.Quotes["USD"].PercentChange24h
	priceChange7d := ticker.Quotes["USD"].PercentChange7d
	priceChange30d := ticker.Quotes["USD"].PercentChange30d
	priceChange1y := ticker.Quotes["USD"].PercentChange1y
	if ticker.Name == nil || priceChange1h == nil || priceChange12h == nil || priceChange24h == nil {
		return "", errors.Wrap(errors.New("missing data"), "command /c")
	}

	return fmt.Sprintf(`%s price change:
		 1h:  %.2f %%
		 12h:  %.2f %%
		 24h:  %.2f %%
		 7d:  %.2f %%
		 30d:  %.2f %%
		 1y:  %.2f %%`,
		 *ticker.Name, *priceChange1h, *priceChange12h, *priceChange24h, *priceChange7d, *priceChange30d, *priceChange1y), nil
}
/*
func commandVolume(argument string) (string, error) {
	log.Debugf("processing command /v with argument :%s", argument)

	ticker, err := getTickerByQuery(argument)
	if err != nil {
		return "", errors.Wrap(err, "command /v")
	}

	volumeUSD := ticker.Quotes["USD"].Volume24h
	if ticker.Name == nil || ticker.ID == nil || volumeUSD == nil {
		return "", errors.Wrap(errors.New("missing data"), "command /v")
	}

	return fmt.Sprintf("%s 24h volume: %.2f USD", *ticker.Name, *volumeUSD), nil
}
*/
func getTickerByQuery(query string) (*coinpaprika.Ticker, error) {
	paprikaClient := coinpaprika.NewClient(nil)

	searchOpts := &coinpaprika.SearchOptions{Query: query, Categories: "currencies", Modifier: "symbol_search"}
	result, err := paprikaClient.Search.Search(searchOpts)
	if err != nil {
		return nil, errors.Wrap(err, "query:"+query)
	}

	log.Debugf("found %d results for query by symbol :%s", len(result.Currencies), query)
	if len(result.Currencies) <= 0 {
		//search by name:
		searchOpts = &coinpaprika.SearchOptions{Query: query, Categories: "currencies"}
		result, err = paprikaClient.Search.Search(searchOpts)
		if err != nil {
			return nil, errors.Wrap(err, "query:"+query)
		}
		log.Debugf("found %d results for query by name :%s", len(result.Currencies), query)

		if len(result.Currencies) <= 0 {
			return nil, errors.Errorf("invalid coin name|ticker|symbol")
		}
	}
	if result.Currencies[0].ID == nil {
		return nil, errors.New("missing id for a coin")
	}

	log.Debugf("best match for query :%s is: %s", query, *result.Currencies[0].ID)

	tickerOpts := &coinpaprika.TickersOptions{Quotes: "USD,BTC"}
	ticker, err := paprikaClient.Tickers.GetByID(*result.Currencies[0].ID, tickerOpts)
	if err != nil {
		return nil, errors.Wrap(err, "query:"+query)
	}

	return ticker, nil
}
