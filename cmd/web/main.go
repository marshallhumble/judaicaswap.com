package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	//Internal
	"judaicaswap.com/internal/models"

	//External
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	logger         *slog.Logger
	Share          models.ShareInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	config         models.ServerConfigInterface
	S3Client       *s3.Client
	S3Bucket       string
	S3Url          string
}

// MaxUploadSize defines the largest file that can be uploaded in the system
const MaxUploadSize = 2024 * 2024

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	dbPass, dbUser, dbHost, dbName, s3BucketName, s3UrlName, s3region, err := readFileEnvs(".env")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	//AWS Login
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(s3region))

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	awsClient := s3.NewFromConfig(cfg)

	addr := flag.String("addr", ":443", "HTTP network address")
	dsn := flag.String("dsn", dbUser+":"+dbPass+"@tcp("+dbHost+":3306)/"+dbName+"?parseTime=true", "MySQL data source name")

	flag.Parse()

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		logger:         logger,
		Share:          &models.ShareModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		config:         &models.ServerConfigModel{DB: db},
		S3Client:       awsClient,
		S3Bucket:       s3BucketName,
		S3Url:          s3UrlName,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", "addr", srv.Addr)

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB open the db and check if the tables exist
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// readFileEnvs pull the sensitive data details from the .ENV file that we are using for Docker init
func readFileEnvs(fileName string) (dbPass, dbUser, dbHost, dbName, s3bucket, s3url, s3region string, err error) {

	file, err := os.Open(fileName)
	if err != nil {
		return "", "", "", "", "", "", "", err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", "", "", "", "", "", err
	}

	text := string(data)

	dbName = getVariable(text, "DB_DATABASE")
	dbPass = getVariable(text, "DB_PASSWORD")
	dbUser = getVariable(text, "DB_USERNAME")
	dbHost = getVariable(text, "DB_HOST")
	s3bucket = getVariable(text, "S3BUCKET")
	s3url = getVariable(text, "S3URL")
	s3region = getVariable(text, "S3_REGION")

	return dbPass, dbUser, dbHost, dbName, s3bucket, s3url, s3region, nil
}

// getVariable get the variables from the ENV file, right now we are assuming they look like this:
// DB_USERNAME=username
// DB_PASSWORD=password
// DB_DATABASE=db_name
// S3BUCKET=s3bucketname
// S3URL=s3url
func getVariable(text, key string) string {

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if strings.Contains(line, key) {
			// Split the line into key-value pairs
			parts := strings.Split(line, "=")

			// Get the value of the variable
			return parts[1]
		}

	}
	return ""
}
