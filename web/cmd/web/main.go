package main

import (
	"encoding/gob"
	"flag"
	"log"
	"log/slog"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/lmittmann/tint"
	"github.com/micahco/mono/lib/data/postgres"
	"github.com/micahco/mono/lib/mailer"
)

var logger *slog.Logger

type config struct {
	dev  bool
	port int
	url  string
	db   struct {
		dsn string
	}
	smtp struct {
		port     int
		host     string
		username string
		password string
		sender   string
	}
}

func main() {
	var cfg config

	flag.BoolVar(&cfg.dev, "dev", false, "Development mode")
	flag.IntVar(&cfg.port, "port", getEnvInt("WEB_PORT"), "web server port")
	flag.StringVar(&cfg.url, "url", os.Getenv("WEB_URL"), "base url for building links")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")

	flag.IntVar(&cfg.smtp.port, "smtp-port", getEnvInt("SMTP_PORT"), "SMTP port")
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("WEB_SMTP_SENDER"), "SMTP sender")

	flag.Parse()

	// Logger
	h := newSlogHandler(cfg.dev)
	logger = slog.New(h)
	// Create error log for http.Server
	errLog := slog.NewLogLogger(h, slog.LevelError)

	// DB
	pg, err := postgres.NewPostgresDB(cfg.db.dsn)
	if err != nil {
		fatal(err)
	}
	defer pg.Close()

	// Mailer
	sender := &mail.Address{
		Name:    "Do Not Reply",
		Address: cfg.smtp.sender,
	}
	m, err := mailer.New(
		cfg.smtp.host,
		cfg.smtp.port,
		cfg.smtp.username,
		cfg.smtp.password,
		sender,
	)
	if err != nil {
		fatal(err)
	}

	// Session manager
	sm := scs.New()
	sm.Store = pgxstore.New(pg.Pool)
	sm.Lifetime = 12 * time.Hour
	gob.Register(uuid.UUID{})
	gob.Register(FormErrors{})

	// Base URL
	baseURL, err := url.Parse(cfg.url)
	if err != nil {
		fatal(err)
	}

	app := &application{
		config:         cfg,
		db:             *pg.DB,
		logger:         logger,
		mailer:         m,
		sessionManager: sm,
		formDecoder:    form.NewDecoder(),
		validate:       validator.New(),
		baseURL:        baseURL,
	}

	err = app.serve(errLog)
	if err != nil {
		fatal(err)
	}
}

func newSlogHandler(dev bool) slog.Handler {
	if dev {
		// Development text hanlder
		return tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:  true,
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})
	}

	// Production use JSON handler with default opts
	return slog.NewJSONHandler(os.Stdout, nil)
}

func fatal(err error) {
	logger.Error("fatal", slog.Any("err", err))
	os.Exit(1)
}

func getEnvInt(key string) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return 0
	}

	v, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("getEnvInt(%v): %v", key, err)
	}

	return v
}
