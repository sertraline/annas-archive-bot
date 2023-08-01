package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func main() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)

	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)

	if err != nil {
		log.Fatal(err)
		return
	}
	b.Use(middleware.Logger())

	b.Handle("/books", BookPaginator)

	b.Handle(&bookBtnNext, GetNextPage)
	b.Handle(&bookBtnPrev, GetPrevPage)
	b.Handle(&bookBtnReset, ResetPage)
	b.Handle(&bookBtnBack, BackPage)
	b.Handle(&bookBtnDownload, DownloadItem)

	b.Start()
}
