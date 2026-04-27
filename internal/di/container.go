package di

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/resoul/api/internal/config"
	"github.com/resoul/api/internal/domain"
	infradb "github.com/resoul/api/internal/infrastructure/db"
	"github.com/resoul/api/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Container is the single composition root for the application.
// It is constructed once in cmd/serve.go and closed on shutdown.
// Handlers and services receive only the specific fields they need —
// never the full Container.
type Container struct {
	Config         *config.Config
	DB             *gorm.DB
	AccountService domain.AccountService
	StorageClient  *minio.Client
}

func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.Init(ctx)

	db, err := openDB(cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	storageClient, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}

	accountRepo := infradb.NewAccountRepository(db)
	accountSvc := service.NewAccountService(accountRepo, storageClient)

	return &Container{
		Config:         cfg,
		DB:             db,
		AccountService: accountSvc,
		StorageClient:  storageClient,
	}, nil
}

func (c *Container) Close() error {
	if c == nil || c.DB == nil {
		return nil
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// openDB opens a PostgreSQL connection with sane pool defaults.
func openDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
