package methods

import (
	"fmt"
	"goseed/generators"
	"goseed/log"
	"goseed/schemas"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

func NewDbStore(db *sqlx.DB) *DbStore {
	return &DbStore{DB: db}
}

type DbStore struct {
	*sqlx.DB
}

func (s *DbStore) Setup(relFilePath string) error {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get executable path: " + err.Error())
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	relFilePath = filepath.Join(pwd, relFilePath)
	setupFile, err := os.Open(relFilePath)
	if err != nil {
		log.Fatal("failed to open setup file: " + err.Error())
		return fmt.Errorf("failed to open setup file: %w", err)
	}
	sqlSetupString, err := io.ReadAll(setupFile)
	if err != nil {
		log.Fatal("failed to read setup file: " + err.Error())
		return fmt.Errorf("failed to read setup file: %w", err)
	}
	// scanner := bufio.NewScanner(setupFile)
	// for scanner.Scan() {
	// 	sqlSetupString = sqlSetupString + scanner.Text() + "\n"
	// }
	// setupFile.Close()
	sqlStringArray := strings.Split(string(sqlSetupString[:]), ";")
	for _, v := range sqlStringArray {
		if strings.TrimSpace(v) == "" {
			continue
		}
		_, err = s.Exec(v)
		if err != nil {
			return fmt.Errorf("failed to setup database: %w", err)
		}
	}
	// _, err = s.Exec(string(sqlSetupString[:]))
	// if err != nil {
	// 	return fmt.Errorf("failed to setup database: %w", err)
	// }
	return nil
}

func (s *DbStore) GetTableFields(database, table string) ([]schemas.TableFields, error) {
	result := []schemas.TableFields{}
	s.Select(&result, fmt.Sprintf("SELECT COLUMN_NAME AS 'Field', COLUMN_TYPE AS `Type`, IS_NULLABLE AS `NULL`, COLUMN_KEY AS `Key`,COLUMN_DEFAULT AS `Default`, EXTRA AS `Extra` FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s';", database, table))
	return result, nil
}

func (s *DbStore) GenerateInsertionMap(fields []schemas.TableFields, seedSize int64) []map[string]schemas.InsertionMap {
	log.Info("Generating Rows...")
	start := time.Now()
	mapArray := make([]map[string]schemas.InsertionMap, seedSize)
	// fmt.Println(" ")
	for i := int64(0); i < seedSize; i++ {
		// fmt.Printf("\033[1A\033[K Rows Generated: %v/%v\n", (i + 1), seedSize)
		result := make(map[string]schemas.InsertionMap)
		for _, v := range fields {
			strValue, intValue, err := generators.GenerateTableFieldValue(v, int(i))
			if err != nil {
				log.Fatal("failed to generate insertion map: " + err.Error())
			}
			if intValue != nil {
				im := schemas.InsertionMap{
					StrValue: "",
					IntValue: intValue,
				}
				result[v.Field] = im
				continue
			}

			result[v.Field] = schemas.InsertionMap{
				StrValue: strValue,
				IntValue: nil,
			}

		}
		mapArray[i] = result
	}
	log.Info("Generating Rows took: " + time.Since(start).String())

	return mapArray
}

func (s *DbStore) BatchInsertFromMap(bArr []map[string]schemas.InsertionMap, fields []schemas.TableFields, table string, chunkSize int64, dbName string, maxConn int, wg *sync.WaitGroup) error {
	log.Info("Generating SQL Column Mapping...")
	bArrLen := len(bArr)
	columnsString := "("
	var valuesString = []string{}
	fieldsOrder := []string{}
	for _, v := range fields {
		columnsString = columnsString + v.Field + ", "
		fieldsOrder = append(fieldsOrder, v.Field)
	}
	columnsString = strings.TrimSuffix(columnsString, ", ") + ")"
	utilString := ""
	log.Info("Generating SQL Value Strings and Sending Batches...")
	maxLimit := maxConn - 10
	if maxLimit < 1 {
		maxLimit = 1
	}
	limiter := make(chan int, maxLimit)
	for idx, v := range bArr {
		utilString = "("
		for _, v2 := range fieldsOrder {
			mapV, ok := v[v2]
			if ok {
				if mapV.StrValue != "" {
					utilString = utilString + "'" + mapV.StrValue + "', "
				} else {
					utilString = utilString + strconv.FormatInt(mapV.IntValue.Number(), 10) + ", "
				}
			}
		}
		utilString = strings.TrimSuffix(utilString, ", ") + ") "
		valuesString = append(valuesString, utilString)

		if int64(idx+1)%chunkSize == 0 && int64(idx+1) >= chunkSize || idx == bArrLen {
			wg.Add(1)
			limiter <- 1
			go func(sql string, wg *sync.WaitGroup, db *sqlx.DB) {
				_, err := db.Exec(sql)
				if err != nil {
					log.Fatal("failed to batch insert from map: " + err.Error())
				}

				<-limiter
				wg.Done()
			}("INSERT INTO "+dbName+"."+table+" "+columnsString+" VALUES "+strings.TrimSuffix(strings.Join(valuesString[:], ", "), ", ")+";", wg, s.DB)
			valuesString = []string{}
		}
	}
	return nil
}

func (s *DbStore) SelectCount(table string, dbName string) (int64, error) {
	var count int64
	err := s.Get(&count, "SELECT COUNT(*) FROM "+""+dbName+"."+table+";")
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}
	return count, nil
}

func (s *DbStore) GetMaxConnections() (int, error) {
	max := schemas.ShowVariables{}
	err := s.Get(&max, "SHOW VARIABLES LIKE 'max_connections';")
	if err != nil {
		log.Info("max_connections not found. Setting to default: 1")
		return 1, nil
	}
	maxValue, err := strconv.Atoi(max.Value)
	if err != nil {
		log.Fatal("failed to get max connections: " + err.Error())
		return 0, fmt.Errorf("failed to get max connections: %w", err)
	}
	if maxValue == 0 {
		log.Info("max_connections found is 0. Setting to default: 100")
		return 100, nil
	}
	if maxValue == 1 {
		return 1, nil
	}
	return maxValue, nil
}
