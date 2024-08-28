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
	err = Setup(db)
	if err != nil {
		log.Fatal("failed to setup database:" + err.Error())
		return
	}
	// Logic here
	err = db.DbStore.UseDatabase(dbName)
	if err != nil {
		log.Fatal("USE DATABASE failed:" + err.Error())
		return
	}

	fields, err := db.DbStore.GetTableFields(dbName, table)
	if err != nil {
		log.Fatal("failed to get table fields:" + err.Error())
		return
	}
	// insertMaps := make([]schemas.InsertionMap, SeedSize)
	insertMap := db.DbStore.GenerateInsertionMap(fields, seedSize)
	err = db.DbStore.BatchInsertFromMap(insertMap, fields, table, chunkSize)
	if err != nil {
		log.Fatal("failed to batch insert from map:" + err.Error())
		return
	}

	count, err := db.DbStore.SelectCount(table)
	if err != nil {
		log.Fatal("failed to select count:" + err.Error())
		return
	}
	log.Info("TABLE " + table + " count: " + strconv.FormatInt(count, 10))
	// Logic before here

	// testing(db)
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

func testing(db *sql.Store) error {
	db.PersonStore.Seed()
	_, err := db.PersonStore.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get all person: %w", err)
	}
	dbs := []string{}
	err = db.Select(&dbs, "SHOW DATABASES;")
	if err != nil {
		return fmt.Errorf("failed to get databases: %w", err)
	}
	fmt.Printf("SHOW DATABASES: %+q\n", dbs)

	tables := []string{}
	err = db.Select(&tables, "SHOW TABLES;")
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}
	fmt.Printf("SHOW TABLES: %+q\n", tables)
	descTable := []DescribeTable{}
	err = db.Select(&descTable, "DESC goseed.person;")
	if err != nil {
		return fmt.Errorf("failed to get describe table: %w", err)
	}
	fmt.Printf("DESC goseed.person: \n")
	for _, v := range descTable {
		fmt.Printf("	%+v\n", v)
	}
	fmt.Println()

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
