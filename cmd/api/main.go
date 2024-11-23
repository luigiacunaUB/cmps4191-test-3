//Luigi Acuna
//CMPS4191 Test 3
//Novemeber 23 2024

// Luigi Acuna
// CMPS4191 Test 3 Advanced Web Dev
// October 30 2024
package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	//"github.com/luigiacunaUB/cmps4191-test-3/internal/data"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port       int    //port number to access signin page
	enviroment string //enviroment the signin page will be on
	db         struct {
		dsn string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type applicationDependencies struct {
	config serverConfig
	logger *slog.Logger
	//Add Models for various tables (also dont foreget to add DB conn to each )
	//ProductModel data.ProductModel
	//ReviewModel  data.ReviewModel
}

func main() {
	var settings serverConfig

	//Settings ports and enviroment info
	flag.IntVar(&settings.port, "port", 4000, "Server Port")
	flag.StringVar(&settings.enviroment, "env", "development", "Enviroment(development|staging|)")
	flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://admin:password123@localhost/amazon?sslmode=disable", "PostgreSQL DSN")

	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")
	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")
	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	appInstance := &applicationDependencies{
		config: settings,
		logger: logger,
		//ProductModel: data.ProductModel{DB: db},
		//ReviewModel:  data.ReviewModel{DB: db},
	}

	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(settings serverConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
