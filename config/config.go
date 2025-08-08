// config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv" // 确保导入了这个包
)

// DBConfig holds database connection details.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Load loads configuration from environment variables or .env files.
// Priority (highest to lowest):
// 1. OS Environment Variables
// 2. .env.{ENV} file (e.g., .env.test, .env.production) - loaded first if ENV is set
// 3. Default .env file
func Load() (*DBConfig, error) {
	// 1. Check for a specific environment variable (e.g., ENV)
	//    This determines which specific .env file to load preferentially.
	env := os.Getenv("ENV")
	// We don't set a default here because we want to allow .env file to be loaded
	// even if ENV is not explicitly set.

	// 2. Load the environment-specific .env file FIRST (if ENV is set and file exists)
	//    Using Overload ensures its values take precedence over .env (if loaded later)
	//    but can still be overridden by OS environment variables.
	if env != "" {
		envFile := fmt.Sprintf(".env.%s", env)
		// Use Overload to ensure variables from this file are set,
		// potentially overriding system env vars if godotenv allows,
		// but more importantly, overriding the default .env loaded later.
		// Note: Standard practice is OS Env > File Env. godotenv.Load usually
		// doesn't override existing OS Env Vars. Let's load specific env file first.
		if err := godotenv.Overload(envFile); err != nil {
			fmt.Printf("Info: No specific env file %s found or error overloading it: %v\n", envFile, err)
			// It's okay if the specific env file doesn't exist
		} else {
			fmt.Printf("Overloaded environment variables from %s\n", envFile) // Indicate precedence
		}
	}

	// 3. Load the default .env file SECOND (if it exists)
	//    Values loaded here will NOT override those from .env.{ENV} (due to loading order and Overload above)
	//    but WILL override any defaults in getEnv calls if .env.{ENV} wasn't loaded or didn't have the var.
	//    It will also NOT override OS Environment Variables.
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist
		fmt.Println("Info: No default .env file found or error loading it:", err)
	} else {
		if env == "" {
			fmt.Println("Loaded default .env file") // Only print this if no specific env was targeted
		} else {
			fmt.Println("Loaded default .env file (as fallback)")
		}
	}

	// 4. Load configuration from environment variables (including those just loaded from files)
	//    OS Environment Variables have the highest priority implicitly (handled by os.Getenv)
	//    If not found in OS Env, values from .env.{ENV} (if loaded and Overloaded) or .env (if loaded) are used.
	//    If not found in any file, the hardcoded defaults are used.
	cfg := &DBConfig{
		Host:     getEnv("DB_HOST", "localhost"), // Default if not in any .env or OS env
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "123456"), // Default, consider security!
		DBName:   getEnv("DB_NAME", "nyc"),        // Default if not set
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
	return cfg, nil
}

// getEnv gets an environment variable or returns a default value.
// It checks OS environment variables. If a variable was loaded from a .env file
// using godotenv.Load/Overload, it becomes part of the OS environment and will be found here.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectionString builds the PostgreSQL connection string for pgxpool.
func (c *DBConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// PoolConfig holds connection pool configuration.
type PoolConfig struct {
	MinConns int32
	MaxConns int32
}

// LoadPoolConfig loads pool configuration from environment variables.
func LoadPoolConfig() *PoolConfig {
	return &PoolConfig{
		MinConns: int32(getEnvInt("DB_MIN_CONN_SIZE", 5)),  // Default if not set
		MaxConns: int32(getEnvInt("DB_MAX_CONN_SIZE", 20)), // Default if not set
	}
}

// getEnvInt gets an environment variable as an integer or returns a default value.
func getEnvInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if val, err := strconv.Atoi(valueStr); err == nil {
			return val
		}
		// Log warning if parsing fails?
		fmt.Printf("Warning: Failed to parse %s='%s' as integer, using default %d\n", key, valueStr, defaultValue)
	}
	return defaultValue
}
