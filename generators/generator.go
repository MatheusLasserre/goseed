package generators

import (
	"errors"
	"goseed/log"
	"goseed/schemas"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

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

var supportedNumericTypes = [...]TypesFormat{
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
	if field.Key != nil {
		if (*field.Key) == "PRI" {
			return int64(index + 1), nil
		}
	}

	strSlice := strings.Split(field.Type, "(")
	for _, v := range supportedNumericTypes {
		if strings.ToLower(strSlice[0]) == v.name {
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
	for _, v := range supportedDateTypes {
		strSlice := strings.Split(v.name, "(")
		if strings.ToLower(strSlice[0]) == field.Type {
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
	name      string
	minLength int64
	maxLength int64
}
