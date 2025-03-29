package main

import (
	"flag"
	"log"
	"log/slog"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/micahco/mono/internal/data/postgres"
	"github.com/micahco/mono/internal/mailer"
)

var logger *slog.Logger

type config struct {
	dev  bool
	port int
	db   struct {
		dsn string
	}
	limiter struct {
		enabled bool
		rps     int
	}
	smtp struct {
		port     int
		host     string
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

func main() {
	var cfg config

	flag.BoolVar(&cfg.dev, "dev", false, "Development mode")
	flag.IntVar(&cfg.port, "port", getEnvInt("API_PORT"), "API server port")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")

	flag.IntVar(&cfg.smtp.port, "smtp-port", getEnvInt("SMTP_PORT"), "SMTP port")
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("API_SMTP_SENDER"), "SMTP sender")

	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", getEnvBool("API_LIMITER_ENABLED"), "Enable rate limiter")
	flag.IntVar(&cfg.limiter.rps, "limiter-rps", getEnvInt("API_LIMITER_RPS"), "Rate limiter maximum requests per second")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		if val == "" {
			val = os.Getenv("API_CORS_TRUSTED_ORIGINS")
		}

		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

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

	app := &application{
		config: cfg,
		db:     *pg.DB,
		logger: logger,
		mailer: m,
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

func getEnvBool(key string) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return false
	}

	return strings.ToLower(val) == "true"
}
