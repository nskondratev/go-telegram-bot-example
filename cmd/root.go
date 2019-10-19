/*
Copyright © 2019 Nikita Kondratev <highflyer16@yandex.ru>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/translate"
	"github.com/jackc/pgx"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nskondratev/go-telegram-translator-bot/internal/app/handler/command"
	"github.com/nskondratev/go-telegram-translator-bot/internal/app/handler/voice"
	"github.com/nskondratev/go-telegram-translator-bot/internal/app/middleware"
	"github.com/nskondratev/go-telegram-translator-bot/internal/bot"
	"github.com/nskondratev/go-telegram-translator-bot/internal/logger"
	"github.com/nskondratev/go-telegram-translator-bot/internal/pg"
	googleS2T "github.com/nskondratev/go-telegram-translator-bot/internal/speech2text/google"
	googleT2S "github.com/nskondratev/go-telegram-translator-bot/internal/text2speech/google"
	googleTr "github.com/nskondratev/go-telegram-translator-bot/internal/texttranslate/google"
	usersPgStore "github.com/nskondratev/go-telegram-translator-bot/internal/users/pg"
	"github.com/nskondratev/go-telegram-translator-bot/internal/voicetranslate"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-telegram-translator-bot",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		appCtx, appCancel := context.WithCancel(context.Background())
		logger := getLogger()
		db := getDB()
		usersStore := usersPgStore.New(db)

		// Init translation services
		googleS2TClient, err := speech.NewClient(appCtx)
		if err != nil {
			logger.Fatal().
				Err(err).
				Msg("Failed to create Google Cloud Speech to text API client")
		}
		googleTranslateClient, err := translate.NewClient(appCtx)
		if err != nil {
			logger.Fatal().
				Err(err).
				Msg("Failed to create Google Cloud Translate API client")
		}
		googleT2SClient, err := texttospeech.NewClient(appCtx)
		if err != nil {
			logger.Fatal().
				Err(err).
				Msg("Failed to create Google Cloud Text to speech API client")
		}
		s2t := googleS2T.New(googleS2TClient)
		tr := googleTr.New(googleTranslateClient)
		t2s := googleT2S.New(googleT2SClient)
		vs := voicetranslate.New(s2t, tr, t2s)

		// Init bot
		b, err := bot.New(logger, viper.GetString("API_KEY"))
		if err != nil {
			logger.Fatal().
				Err(err).
				Msg("failed to create bot")
		}

		msgHandler := bot.
			NewChain(
				command.New(&b).Middleware,
			).
			Then(voice.New(&b, vs))

		h := bot.
			NewChain(
				middleware.LogTimeExecution,
				middleware.LogUserInfo,
				middleware.SetUser(usersStore),
			).
			Then(msgHandler)

		b.Handle(h)

		go func() {
			err := b.RunUpdateChannel(appCtx)

			if err != nil {
				logger.Fatal().
					Err(err).
					Msg("error in bot update channel listener")
			}
		}()

		logger.Info().Msg("Start telegram bot application")

		// Wait for interruption
		ic := make(chan os.Signal, 1)
		signal.Notify(ic, os.Interrupt, syscall.SIGTERM)

		<-ic
		logger.Info().Msg("application is interrupted. Stopping appCtx...")
		appCancel()
		time.Sleep(500 * time.Millisecond)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./conf.yml", "config file (default is $HOME/.go-telegram-translator-bot.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level in lowercase")
	rootCmd.PersistentFlags().String("db-conn", "", "connection string to PostgreSQL database")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().String("api-key", "", "Telegram bot API key")

	if err := viper.BindPFlag("LOG_LEVEL", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Fatal(err)
	}
	if err := viper.BindPFlag("API_KEY", rootCmd.Flags().Lookup("api-key")); err != nil {
		log.Fatal(err)
	}

	if err := viper.BindPFlag("DB_CONN", rootCmd.PersistentFlags().Lookup("db-conn")); err != nil {
		log.Fatal(err)
	}

	viper.SetDefault("LOG_LEVEL", "info")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".go-telegram-translator-bot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".go-telegram-translator-bot")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func getLogger() zerolog.Logger {
	l, err := logger.New(viper.GetString("LOG_LEVEL"), os.Stdout)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
	}

	return l
}

func getDB() *pgx.ConnPool {
	db, err := pg.New(viper.GetString("DB_CONN"), 10)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err.Error())
	}
	return db
}
