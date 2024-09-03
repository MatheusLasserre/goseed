package generators

import (
	"errors"
	"fmt"
	"goseed/log"
	"goseed/schemas"
	"goseed/utils"
	"maps"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

func GenerateFieldMap(fields []schemas.TableFields, pKeys int, idx int) ([]map[string]schemas.InsertionMap, error) {
	// loopCount := pKeys
	// if pKeys <= 0 {
	// 	loopCount = 1
	// }
	curPKey := 1
	mapArraySize := utils.PowerInt(2, pKeys-1)
	mapArrIdx := 0
	mapArray := make([]map[string]schemas.InsertionMap, mapArraySize)
	fIdx := 0
	tmpMap := make(map[string]schemas.InsertionMap, len(fields))
	err := GenerateCompositeMaps(mapArray, tmpMap, fields, idx, fIdx, &curPKey, len(fields), &mapArrIdx)
	if err != nil {
		log.Fatal("failed to generate insertion map: " + err.Error())
		return nil, fmt.Errorf("failed to generate insertion map: %w", err)
	}
	// for i := 0; i < loopCount; i++ {
	// 	result := make(map[string]schemas.InsertionMap, pKeys)
	// 	for _, v := range fields {
	// 		strValue, intValue, err := GenerateTableFieldValue(v, idx)
	// 		if err != nil {
	// 			log.Fatal("failed to generate insertion map: " + err.Error())
	// 			return nil, fmt.Errorf("failed to generate insertion map: %w", err)
	// 		}
	// 		if intValue != nil {
	// 			im := schemas.InsertionMap{
	// 				StrValue: "",
	// 				IntValue: intValue,
	// 			}
	// 			result[v.Field] = im
	// 			continue
	// 		}

	// 		result[v.Field] = schemas.InsertionMap{
	// 			StrValue: strValue,
	// 			IntValue: nil,
	// 		}
	// 	}
	// 	mapArray[i] = result
	// }
	return mapArray, nil
}

func GenerateCompositeMaps(mapArr []map[string]schemas.InsertionMap, curMap map[string]schemas.InsertionMap, fields []schemas.TableFields, outerIdx int, fIdx int, curPKey *int, mapLength int, mapArrIdx *int) error {
	if fIdx == mapLength {
		mapArr[*mapArrIdx] = maps.Clone(curMap)
		*mapArrIdx++
		return nil
	}
	if *fields[fIdx].Key != "PRI" || *fields[fIdx].Key == "PRI" && *curPKey == 1 {
		if *fields[fIdx].Key == "PRI" && *curPKey == 1 {
			*curPKey = 0
		}

		strValue, intValue, err := GenerateTableFieldValue(fields[fIdx], outerIdx)
		if err != nil {
			log.Fatal("failed to generate insertion map: " + err.Error())
			return fmt.Errorf("failed to generate insertion map: %w", err)
		}
		if intValue != nil {
			curMap[fields[fIdx].Field] = schemas.InsertionMap{
				StrValue: "",
				IntValue: intValue,
			}
			err = GenerateCompositeMaps(mapArr, curMap, fields, outerIdx, fIdx+1, curPKey, mapLength, mapArrIdx)
			if err != nil {
				return fmt.Errorf("failed to generate insertion map: %w", err)
			}
			return nil
		}
		curMap[fields[fIdx].Field] = schemas.InsertionMap{
			StrValue: strValue,
			IntValue: nil,
		}
		err = GenerateCompositeMaps(mapArr, curMap, fields, outerIdx, fIdx+1, curPKey, mapLength, mapArrIdx)
		if err != nil {
			return fmt.Errorf("failed to generate insertion map: %w", err)
		}
		return nil
	}

	if *fields[fIdx].Key == "PRI" {
		for i := 0; i < 2; i++ {
			strValue, intValue, err := GenerateTableFieldValue(fields[fIdx], outerIdx+i)
			if err != nil {
				log.Fatal("failed to generate insertion map: " + err.Error())
				return fmt.Errorf("failed to generate insertion map: %w", err)
			}
			if intValue != nil {
				curMap[fields[fIdx].Field] = schemas.InsertionMap{
					StrValue: "",
					IntValue: intValue,
				}
				err = GenerateCompositeMaps(mapArr, curMap, fields, outerIdx, fIdx+1, curPKey, mapLength, mapArrIdx)
				if err != nil {
					return fmt.Errorf("failed to generate insertion map: %w", err)
				}
				continue
			}
			curMap[fields[fIdx].Field] = schemas.InsertionMap{
				StrValue: strValue,
				IntValue: nil,
			}
			err = GenerateCompositeMaps(mapArr, curMap, fields, outerIdx, fIdx+1, curPKey, mapLength, mapArrIdx)
			if err != nil {
				return fmt.Errorf("failed to generate insertion map: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf("failed to generate insertion map: Failed to build map. This error should never happen, please open an issue")
}

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

func CountPrimaryKeys(fields []schemas.TableFields) int {
	count := 0
	for _, v := range fields {
		if v.Key != nil {
			if (*v.Key) == "PRI" {
				count++
			}
		}
	}
	return count
}

var GenerateValue = NewValuesGenerator()

func NewValuesGenerator() *ValuesGenerator {
	return &ValuesGenerator{}
}

var supportedStringTypes = [...]TypesFormat{
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
		if strings.ToLower(strSlice[0]) == v.Name {
			if field.Key != nil {
				if (*field.Key) == "PRI" {
					return uuid.NewString(), nil
				}
			}
			if len(strSlice) > 1 {
				typeLength, err := strconv.ParseInt(strings.Replace(strSlice[1], ")", "", -1), 10, 64)
				if err != nil {
					log.Fatal("failed to parse type " + v.Name + ": " + err.Error())
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

var SupportedNumericTypes = [...]TypesFormat{
	{"bit", 1, 64},
	{"tinyint", -128, 127},
	{"smallint", -32768, 32767},
	{"mediumint", -8388608, 8388607},
	{"int", -2147483648, 2147483647},
	{"bigint", -9223372036854775808, 9223372036854775807},
	{"float", 650535, 650535},
	{"double", 650535, 650535},
	{"decimal", 650535, 650535},
	{"numeric", 0, 65535},
	{"boolean", 0, 1},
	{"bool", 0, 1},
}

func (v *ValuesGenerator) GenerateNumericTypes(field schemas.TableFields, index int) (int64, error) {

	strSlice := strings.Split(field.Type, "(")
	for _, v := range SupportedNumericTypes {
		if strings.ToLower(strSlice[0]) == v.Name {
			if field.Key != nil {
				if (*field.Key) == "PRI" {
					return int64(index + 1), nil
				}
			}
			multiplier := int64(1)
			if rand.IntN(2) == 1 && v.minLength > 0 {
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
			if strings.Contains(field.Type, "unsigned") {
				maxLength = maxLength + v.maxLength - 1
				multiplier = 1
			}

			value := rand.Int64N(maxLength) * multiplier

			return value, nil

		}
	}

	return 0, errors.New("failed to generate table field value for type: " + field.Type)
}

var supportedDateTypes = [...]TypesFormat{
	{"date", 0, 0},
	{"datetime", 0, 0},
	{"timestamp", 0, 0},
	{"time", 0, 0},
	{"year", 0, 0},
}

func (v *ValuesGenerator) GenerateDateTypes(field schemas.TableFields) (string, error) {
	strSlice := strings.Split(field.Type, "(")
	for _, v := range supportedDateTypes {
		if strings.ToLower(strSlice[0]) == v.Name {
			randDate := generateRandomUnixTime(946692000, 1893466800)
			if field.Type == "date" {
				return randDate.Format("2006-01-02"), nil
			}
			if field.Type == "datetime" {
				return randDate.Format("2006-01-02 15:04:05"), nil
			}
			if field.Type == "timestamp" {
				return randDate.Format("2006-01-02 15:04:05"), nil
			}
			if field.Type == "time" {
				return randDate.Format("15:04:05"), nil
			}
		}

	}
	return "", errors.New("failed to generate table field value for type: " + field.Type)
}

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func generateRandomUnixTime(min, max int) time.Time {
	return time.Unix(int64(randRange(min, max)), 0)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func GenerateRandomString(max int) string {
	tSize := rand.IntN(max) + 1
	b := make([]byte, tSize)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := tSize-1, rand.Int64(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int64(), letterIdxMax
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

type TypesFormat struct {
	Name      string
	minLength int64
	maxLength int64
}
