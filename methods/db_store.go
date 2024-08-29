package methods

import (
	"errors"
	"fmt"
	"goseed/log"
	"goseed/schemas"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/google/uuid"
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
			strValue, intValue, err := GenerateTableFieldValue(v, int(i))
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

var GenerateValue = NewValuesGenerator()

func GenerateTableFieldValue(fields schemas.TableFields, index int) (string, schemas.NumberNil, error) {
	if fields.Extra != nil {
		if len(*fields.Extra) > 0 {
			if *fields.Extra == "auto_increment" {
				return "auto_increment", NumberImpl{number: int64(index + 1)}, nil
			}
		}
	}
	strRes, err := GenerateValue.GenerateStringTypes(fields)
	if err == nil {
		return strRes, nil, nil
	}
	numRes, err := GenerateValue.GenerateNumericTypes(fields, index)
	if err == nil {
		return "", NumberImpl{number: numRes}, nil
	}
	dateRes, err := GenerateValue.GenerateDateTypes(fields)
	if err == nil {
		return dateRes, nil, nil
	}
	return "", nil, errors.New("Failed to generate table field value for type: " + fields.Type)
}

type NumberImpl struct {
	number int64
}

func (n NumberImpl) Number() int64 {
	return n.number
}

type IValuesGenerator interface {
	GenerateStringTypes(field schemas.TableFields, index int) (string, error)
	GenerateNumericTypes(field schemas.TableFields) (int, error)
	GenerateDateTimeTypes(field schemas.TableFields) (string, error)
}

type ValuesGenerator struct {
}

func NewValuesGenerator() *ValuesGenerator {
	return &ValuesGenerator{}
}

var supportedStringTypes = [...]supportedStrings{
	{"varchar", 0, 65535},
	{"char", 0, 255},
	{"tinytext", 0, 255},
	{"text", 0, 65535},
	{"mediumtext", 0, 16777215},
	{"longtext", 0, 4294967295},
}

func (v *ValuesGenerator) GenerateStringTypes(field schemas.TableFields) (string, error) {
	for _, v := range supportedStringTypes {
		strSlice := strings.Split(field.Type, "(")
		if strings.ToLower(strSlice[0]) == v.name {
			if field.Key != nil {
				if (*field.Key) == "PRI" {
					return uuid.NewString(), nil
				}
			}
			if len(strSlice) > 1 {
				typeLength, err := strconv.ParseInt(strings.Replace(strSlice[1], ")", "", -1), 10, 64)
				if err != nil {
					log.Fatal("failed to parse type " + v.name + ": " + err.Error())
					return "", err
				}
				if typeLength >= v.minLength && typeLength <= v.maxLength {

					return GenerateRandomString(int(typeLength)), nil
				}
			}
		}
	}

	return "", errors.New("failed to generate table field value for type: " + field.Type)
}

type supportedStrings struct {
	name      string
	minLength int64
	maxLength int64
}

var supportedNumericTypes = [...]supportedStrings{
	{"bit", 1, 64},
	{"tinyint", -128, 127},
	{"smallint", -32768, 32767},
	{"mediumint", -8388608, 8388607},
	{"int", -2147483648, 2147483647},
	{"bigint", -9223372036854775808, 9223372036854775807},
	{"float", 0, 65535},
	{"double", 0, 65535},
	{"decimal", 0, 65535},
	{"numeric", 0, 65535},
	{"boolean", 0, 1},
	{"bool", 0, 1},
}

func (v *ValuesGenerator) GenerateNumericTypes(field schemas.TableFields, index int) (int64, error) {
	if field.Key != nil {
		if (*field.Key) == "PRI" {
			return int64(index + 1), nil
		}
	}

	for _, v := range supportedNumericTypes {
		strSlice := strings.Split(v.name, "(")
		if strings.ToLower(strSlice[0]) == field.Type {
			multiplier := int64(1)
			if rand.Int63n(10) > 5 && v.minLength < 0 {
				multiplier = -1
			}
			if v.minLength == 0 && v.maxLength == 1 {
				if multiplier == -1 {
					return int64(0), nil
				} else {
					return int64(1), nil
				}
			}
			maxLength := v.maxLength
			if strings.Contains(v.name, "unsigned") {
				maxLength = maxLength + v.maxLength - 1
				multiplier = 1
			}

			value := rand.Int63n(maxLength) * multiplier

			return value, nil

		}
	}

	return 0, errors.New("failed to generate table field value for type: " + field.Type)
}

var supportedDateTypes = [...]supportedStrings{
	{"date", 0, 0},
	{"datetime", 0, 0},
	{"timestamp", 0, 0},
	{"time", 0, 0},
	{"year", 0, 0},
}

func (v *ValuesGenerator) GenerateDateTypes(field schemas.TableFields) (string, error) {
	for _, v := range supportedDateTypes {
		strSlice := strings.Split(v.name, "(")
		if strings.ToLower(strSlice[0]) == field.Type {
			if field.Type == "date" {
				return time.Now().Format("2006-01-02"), nil
			}
			if field.Type == "datetime" {
				return time.Now().Format("2006-01-02 15:04:05"), nil
			}
			if field.Type == "timestamp" {
				return time.Now().Format("2006-01-02 15:04:05"), nil
			}
			if field.Type == "time" {
				return time.Now().Format("15:04:05"), nil
			}
		}

	}
	return "", errors.New("failed to generate table field value for type: " + field.Type)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
