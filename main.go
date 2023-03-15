//go:generate go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=web
//go:generate go install golang.org/x/text/cmd/gotext@latest
//go:generate gotext -srclang=en update -out=catalog_gen.go -lang=en,ru
package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v7"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"source.toby3d.me/toby3d/hub/internal/domain"
	hubhttprelivery "source.toby3d.me/toby3d/hub/internal/hub/delivery/http"
	hubucase "source.toby3d.me/toby3d/hub/internal/hub/usecase"
	"source.toby3d.me/toby3d/hub/internal/middleware"
	subscriptionmemoryrepo "source.toby3d.me/toby3d/hub/internal/subscription/repository/memory"
	subscriptionucase "source.toby3d.me/toby3d/hub/internal/subscription/usecase"
	topicmemoryrepo "source.toby3d.me/toby3d/hub/internal/topic/repository/memory"
	topicucase "source.toby3d.me/toby3d/hub/internal/topic/usecase"
	"source.toby3d.me/toby3d/hub/internal/urlutil"
)

var logger = log.New(os.Stdout, "hub", log.LstdFlags)

//go:embed web/static/*
var static embed.FS

func init() {
	message.Set(language.English, "%d subscribers",
		plural.Selectf(1, "%d",
			"one", "%d subscriber",
			"other", "%d subscribers",
		),
	)
	message.Set(language.Russian, "%d subscribers",
		plural.Selectf(1, "%d",
			"one", "%d подписчик",
			"few", "%d подписчика",
			"many", "%d подписчиков",
			"other", "%d подписчика",
		),
	)
}

func main() {
	ctx := context.Background()

	config := new(domain.Config)
	if err := env.Parse(config, env.Options{Prefix: "HUB_"}); err != nil {
		logger.Fatalln(err)
	}

	static, err := fs.Sub(static, filepath.Join("web"))
	if err != nil {
		logger.Fatalln(err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
	subscriptions := subscriptionmemoryrepo.NewMemorySubscriptionRepository()
	topics := topicmemoryrepo.NewMemoryTopicRepository()
	topicService := topicucase.NewTopicUseCase(topics, client)
	subscriptionService := subscriptionucase.NewSubscriptionUseCase(subscriptions, topics, client)
	hubService := hubucase.NewHubUseCase(topics, subscriptions, client, config.BaseURL)

	handler := hubhttprelivery.NewHandler(hubhttprelivery.NewHandlerParams{
		Hub:           hubService,
		Subscriptions: subscriptionService,
		Topics:        topicService,
		Matcher:       matcher,
		Name:          config.Name,
	})

	server := &http.Server{
		Addr: config.Bind,
		Handler: http.HandlerFunc(middleware.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			head, _ := urlutil.ShiftPath(r.URL.Path)

			switch head {
			case "":
				handler.ServeHTTP(w, r)
			case "static":
				http.FileServer(http.FS(static)).ServeHTTP(w, r)
			}
		}).Intercept(middleware.LogFmt())),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		ErrorLog:     logger,
	}

	go hubService.ListenAndServe(ctx)

	logger.Printf("started %s on %s: %s", config.Name, config.Bind, config.BaseURL.String())
	if err = server.ListenAndServe(); err != nil {
		logger.Fatalln(err)
	}
}
