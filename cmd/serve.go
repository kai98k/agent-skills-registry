//go:build server

package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liuyukai/agentskills/internal/database"
	"github.com/liuyukai/agentskills/internal/server"
	"github.com/liuyukai/agentskills/internal/storage"
	"github.com/spf13/cobra"
)

func init() {
	serveCmd.Flags().IntP("port", "p", 8000, "HTTP listen port")
	serveCmd.Flags().String("db-driver", envOrDefault("AGENTSKILLS_DB_DRIVER", "sqlite"), "Database driver: sqlite or postgres")
	serveCmd.Flags().String("db-dsn", envOrDefault("AGENTSKILLS_DB_DSN", "./data/agentskills.db"), "Database connection string")
	serveCmd.Flags().String("storage-driver", envOrDefault("AGENTSKILLS_STORAGE_DRIVER", "local"), "Storage driver: local or s3")
	serveCmd.Flags().String("storage-path", envOrDefault("AGENTSKILLS_STORAGE_PATH", "./data/bundles"), "Local storage path")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the AgentSkills HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		if p := os.Getenv("AGENTSKILLS_PORT"); p != "" {
			fmt.Sscanf(p, "%d", &port)
		}
		dbDriver, _ := cmd.Flags().GetString("db-driver")
		dbDSN, _ := cmd.Flags().GetString("db-dsn")
		storageDriver, _ := cmd.Flags().GetString("storage-driver")
		storagePath, _ := cmd.Flags().GetString("storage-path")

		// Initialize database
		var db database.Database
		switch dbDriver {
		case "sqlite":
			db = database.NewSQLite(dbDSN)
		default:
			return fmt.Errorf("unsupported db driver: %s (postgres support coming soon)", dbDriver)
		}
		if err := db.Open(); err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()
		if err := db.Migrate(); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
		log.Printf("Database: %s (%s)", dbDriver, dbDSN)

		// Initialize storage
		var store storage.Storage
		switch storageDriver {
		case "local":
			store = storage.NewLocalStorage(storagePath)
		default:
			return fmt.Errorf("unsupported storage driver: %s (s3 support coming soon)", storageDriver)
		}
		if err := store.Init(); err != nil {
			return fmt.Errorf("init storage: %w", err)
		}
		log.Printf("Storage: %s (%s)", storageDriver, storagePath)

		// Start server
		srv := server.New(db, store, port)

		// Graceful shutdown
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGTERM)

		go func() {
			if err := srv.Start(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()

		<-done
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbDriver, _ := cmd.Flags().GetString("db-driver")
		dbDSN, _ := cmd.Flags().GetString("db-dsn")

		var db database.Database
		switch dbDriver {
		case "sqlite":
			db = database.NewSQLite(dbDSN)
		default:
			return fmt.Errorf("unsupported db driver: %s", dbDriver)
		}
		if err := db.Open(); err != nil {
			return err
		}
		defer db.Close()

		if err := db.Migrate(); err != nil {
			return err
		}
		log.Println("Migration completed successfully.")
		return nil
	},
}

func init() {
	migrateCmd.Flags().String("db-driver", envOrDefault("AGENTSKILLS_DB_DRIVER", "sqlite"), "Database driver")
	migrateCmd.Flags().String("db-dsn", envOrDefault("AGENTSKILLS_DB_DSN", "./data/agentskills.db"), "Database DSN")
	rootCmd.AddCommand(migrateCmd)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
