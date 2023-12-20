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

	"github.com/caarlos0/env/v10"
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	_ "modernc.org/sqlite"

	"source.toby3d.me/toby3d/hub/internal/domain"
	hubhttprelivery "source.toby3d.me/toby3d/hub/internal/hub/delivery/http"
	hubucase "source.toby3d.me/toby3d/hub/internal/hub/usecase"
	"source.toby3d.me/toby3d/hub/internal/middleware"
	subscriptionsqliterepo "source.toby3d.me/toby3d/hub/internal/subscription/repository/sqlite"
	subscriptionucase "source.toby3d.me/toby3d/hub/internal/subscription/usecase"
	topicsqliterepo "source.toby3d.me/toby3d/hub/internal/topic/repository/sqlite"
	topicucase "source.toby3d.me/toby3d/hub/internal/topic/usecase"
	"source.toby3d.me/toby3d/hub/internal/urlutil"
)

var logger = log.New(os.Stdout, "hub\t", log.LstdFlags)

//go:embed web/static/*
var static embed.FS

func main() {
	ctx := context.Background()

	static, err := fs.Sub(static, filepath.Join("web"))
	if err != nil {
		logger.Fatalln(err)
	}

	config := new(domain.Config)
	if err = env.ParseWithOptions(config, env.Options{
		Prefix:                "HUB_",
		UseFieldNameByDefault: true,
	}); err != nil {
		logger.Fatalln(err)
	}

	db, err := sqlx.Open("sqlite", config.DB)
	if err != nil {
		logger.Fatalf("cannot open database on path %s: %s", config.DB, err)
	}

	topics, err := topicsqliterepo.NewSQLiteTopicRepository(db)
	if err != nil {
		logger.Fatalln(err)
	}

	subscriptions, err := subscriptionsqliterepo.NewSQLiteSubscriptionRepository(db)
	if err != nil {
		logger.Fatalln(err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	matcher := language.NewMatcher(message.DefaultCatalog.Languages())
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
