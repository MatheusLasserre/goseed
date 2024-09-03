package store

import (
	"fmt"
	"goseed/generators"
	"goseed/log"
	"goseed/schemas"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

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

func (s *DbStore) GetTableFields(database, table string) ([]schemas.TableFields, error) {
	result := []schemas.TableFields{}
	s.Select(&result, fmt.Sprintf("SELECT COLUMN_NAME AS 'Field', COLUMN_TYPE AS `Type`, IS_NULLABLE AS `NULL`, COLUMN_KEY AS `Key`,COLUMN_DEFAULT AS `Default`, EXTRA AS `Extra` FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s';", database, table))
	slices.SortStableFunc(result, func(i, j schemas.TableFields) int {
		if *i.Key == *j.Key {
			iv, jv := 0, 0
			jtype := strings.Split(j.Type, "(")
			itype := strings.Split(i.Type, "(")
			for _, v := range generators.SupportedNumericTypes {
				if strings.ToLower(jtype[0]) == v.Name {
					jv = 1
				}
				if strings.ToLower(itype[0]) == v.Name {
					iv = 1
				}
			}
			if iv < jv {
				return 1
			}
			if iv > jv {
				return -1
			}
		}
		if *i.Key == "PRI" {
			return -1
		}
		if *j.Key == "PRI" {
			return 1
		}
		return 0
	})
	return result, nil
}

func (s *DbStore) GenerateInsertionMap(fields []schemas.TableFields, table string, seedSize int64, chunkSize int64, maxConn int, dbName string, wg *sync.WaitGroup) error {
	pKeys := generators.CountPrimaryKeys(fields)
	mapArray := make([]map[string]schemas.InsertionMap, chunkSize)
	ln := seedSize - 1
	maxLimit := maxConn - 10
	if maxLimit < 1 {
		maxLimit = 1
	}
	limiter := make(chan int, maxLimit)
	var decrementer int64 = 0
	for idx := int64(0); idx < seedSize; {
		result, err := generators.GenerateFieldMap(fields, pKeys, int(idx))

		if err != nil {
			log.Fatal("failed to generate insertion map: " + err.Error())
		}
		for j, v := range result {
			mapArray[idx+int64(j)-decrementer] = v
			if (idx+1+int64(j))%chunkSize == 0 && (idx+1+int64(j)) >= chunkSize || idx+int64(j) == ln {
				wg.Add(1)
				limiter <- 1
				go func(mArr []map[string]schemas.InsertionMap) {
					fmt.Printf("mArr: %+v\n", mArr)
					err := s.BatchInsertFromMap(mArr, fields, table, chunkSize, dbName, maxConn)
					if err != nil {
						log.Fatal("failed to batch insert from map: " + err.Error())
					}
					<-limiter
					wg.Done()
				}(mapArray)
				mapArray = make([]map[string]schemas.InsertionMap, chunkSize)
				decrementer = decrementer + chunkSize
				if idx+int64(j) == ln {
					break
				}
			}
		}
		idx = idx + int64(len(result))
	}

	return nil
}

func (s *DbStore) BatchInsertFromMap(bArr []map[string]schemas.InsertionMap, fields []schemas.TableFields, table string, chunkSize int64, dbName string, maxConn int) error {
	columnsString := "("
	var valuesString = []string{}
	fieldsOrder := []string{}
	for _, v := range fields {
		columnsString = columnsString + v.Field + ", "
		fieldsOrder = append(fieldsOrder, v.Field)
	}
	columnsString = strings.TrimSuffix(columnsString, ", ") + ")"
	utilString := ""
	for _, v := range bArr {
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
	}
	// log.Warn("INSERT INTO " + dbName + "." + table + " " + columnsString + " VALUES " + strings.TrimSuffix(strings.Join(valuesString[:], ", "), ", ") + ";")
	_, err := s.Exec("INSERT INTO " + dbName + "." + table + " " + columnsString + " VALUES " + strings.TrimSuffix(strings.Join(valuesString[:], ", "), ", ") + ";")
	if err != nil {
		log.Fatal("failed to batch insert from map: " + err.Error())
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
