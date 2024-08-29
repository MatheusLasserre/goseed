/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"goseed/log"
	"goseed/sql"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goseed",
	Short: "A seed tool for sql databases",
	Long:  `Select a database, a table, and i'll goseed.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: startSeed,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goseed.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("database", "d", "", "use database")
	rootCmd.Flags().StringP("table", "t", "", "from table")
	rootCmd.Flags().StringP("host", "p", "", "Database Connection String. Example: -p \"root:goseed@tcp(localhost:3306)/\"")
	rootCmd.Flags().Int64P("size", "s", 0, "Seed size")
	rootCmd.Flags().Int64P("chunkSize", "c", 0, "How many rows to insert at a time. Default: 100. Recommended: 100.")
	rootCmd.Flags().IntP("max-connections", "m", 0, "How many connections to use. The dafault value is the result of SHOW VARIABLES LIKE 'max_connections'. If max_connections is 0 or not found, the default value is 1.")

}

func startSeed(cmd *cobra.Command, args []string) {
	start := time.Now()
	dbName, err := cmd.Flags().GetString("database")
	if err != nil {
		log.Fatal("failed to detect database:" + err.Error())
	}
	if dbName == "" {
		log.Fatal("Database flag is required. (goseed -u database or goseed --use database)")
		return
	}
	table, err := cmd.Flags().GetString("table")
	if err != nil {
		log.Error("failed to detect table:" + err.Error())
		return
	}
	seedSize, err := cmd.Flags().GetInt64("size")
	if err != nil {
		log.Error("failed to detect seed size:" + err.Error())
		return
	}
	if seedSize == 0 {
		log.Error("failed to detect seed size or input is 0")
		return
	}
	chunkSize, err := cmd.Flags().GetInt64("chunkSize")
	if err != nil {
		log.Error("failed to detect chunk size:" + err.Error())
		return
	}
	if chunkSize == 0 {
		log.Info("failed to detect chunk size or input is 0. Using default chunk size: 100")
		chunkSize = 100
	}
	connStr, err := cmd.Flags().GetString("host")
	if err != nil {
		log.Error("failed to detect connection string:" + err.Error())
		return
	}
	if connStr == "" {
		log.Error("failed to detect connection string or input is empty")
		return
	}
	fmt.Println("Database selected:", dbName)
	fmt.Println("Table selected:", table)
	fmt.Println("Seed Size:", seedSize)
	fmt.Println("Chunk Size:", chunkSize)
	db, err := sql.Connect(connStr)
	if err != nil {
		log.Fatal("failed to connect to database:" + err.Error())
		return
	}
	// This piece of code is for testing purposes and should be removed
	err = Setup(db)
	if err != nil {
		log.Fatal("failed to setup database:" + err.Error())
		return
	}
	maxConn, err := cmd.Flags().GetInt("max-connections")
	if err != nil || maxConn == 0 {
		maxConn, err = db.DbStore.GetMaxConnections()
		if err != nil {
			log.Fatal("failed to get max connections: " + err.Error())
			return
		}
	}
	db.DB.SetMaxOpenConns(maxConn)
	db.DB.SetMaxIdleConns(maxConn)
	fmt.Printf("Max Connections: %v\n", maxConn)
	fields, err := db.DbStore.GetTableFields(dbName, table)
	if err != nil {
		log.Fatal("failed to get table fields:" + err.Error())
		return
	}
	insertMap := db.DbStore.GenerateInsertionMap(fields, seedSize)
	wg := &sync.WaitGroup{}
	start2 := time.Now()
	err = db.DbStore.BatchInsertFromMap(insertMap, fields, table, chunkSize, dbName, maxConn, wg)
	wg.Wait()
	if err != nil {
		log.Fatal("failed to batch insert from map:" + err.Error())
		return
	}
	log.Info("Generating SQL Value Strings and Sending Batches took: " + time.Since(start2).String())

	count, err := db.DbStore.SelectCount(table, dbName)
	if err != nil {
		log.Fatal("failed to select count:" + err.Error())
		return
	}
	log.Info("TABLE " + table + " count: " + strconv.FormatInt(count, 10))

	log.Info("Seed took: " + time.Since(start).String())
}

func Setup(db *sql.Store) error {
	err := db.DbStore.Setup()
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	_, err = db.DB.Exec("USE goseed;")
	if err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}
	err = db.PersonStore.Setup()
	if err != nil {
		return fmt.Errorf("failed to get all person: %w", err)
	}
	return nil
}

type DescribeTable struct {
	Field   string  `db:"Field"`
	Type    string  `db:"Type"`
	Null    string  `db:"Null"`
	Key     *string `db:"Key"`
	Default *string `db:"Default"`
	Extra   *string `db:"Extra"`
}
