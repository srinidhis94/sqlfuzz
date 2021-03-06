package action

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/PumpkinSeed/sqlfuzz/drivers"
	"github.com/brianvoe/gofakeit/v5"
	_ "github.com/lib/pq"
	"github.com/rs/xid"
)

// Insert is inserting a random generated data into the chosen table
func Insert(db *sql.DB, fields []drivers.FieldDescriptor, driver drivers.Driver, table string) error {
	var f = make([]string, 0, len(fields))
	var values = make([]interface{}, 0, len(fields))
	for _, field := range fields {
		// Has default value. No need to insert this field manually.
		if field.HasDefaultValue {
			continue
		}
		f = append(f, field.Field)
		values = append(values, generateData(driver, field))
	}
	query := driver.Insert(f, table)

	_, err := db.Exec(query, values...)
	return err
}

// generateData generates random data based on the field
func generateData(driver drivers.Driver, fieldDescriptor drivers.FieldDescriptor) interface{} {
	field := driver.MapField(fieldDescriptor)
	switch field.Type {
	case drivers.String:
		if field.Length > 19 {
			return xid.New().String()
		}
		if field.Length > 0 {
			return randomString(field.Length)
		}
		return randomString(20)
	case drivers.Int16:
		return gofakeit.Number(1, 32766)
	case drivers.Int32:
		return gofakeit.Number(1, 2147483647)
	case drivers.Float:
		max := 2147483647
		if fieldDescriptor.Precision.Valid && fieldDescriptor.Scale.Valid {
			max = int(math.Pow10(fieldDescriptor.Precision.Int - fieldDescriptor.Scale.Int))
		}
		return gofakeit.Number(1, max)
	case drivers.Blob:
		return base64.StdEncoding.EncodeToString([]byte(randomString(12)))
	case drivers.Text:
		return randomString(12)
	case drivers.Enum:
		return field.Enum[gofakeit.Number(0, len(field.Enum)-1)]
	case drivers.Bool:
		if gofakeit.Number(1, 200)%2 == 0 {
			return true
		}
		return false
	case drivers.Json:
		return fmt.Sprintf(
			`{"%s": "%s", "%s": "%s"}`,
			gofakeit.Password(true, true, false, false, false, 6),
			gofakeit.Password(true, true, false, false, false, 6),
			gofakeit.Password(true, true, false, false, false, 6),
			gofakeit.Password(true, true, false, false, false, 6),
		)
	case drivers.Time:
		return time.Date(
			gofakeit.Number(1970, 2038),
			time.Month(gofakeit.Number(0, 12)),
			gofakeit.Day(),
			gofakeit.Hour(),
			gofakeit.Minute(),
			gofakeit.Second(),
			gofakeit.NanoSecond(),
			time.UTC)
	case drivers.Year:
		return gofakeit.Number(1901, 2155)
	case drivers.XML:
		xml, err := gofakeit.XML(&gofakeit.XMLOptions{
			Type:          "single",
			RootElement:   "xml",
			RecordElement: "record",
			RowCount:      2,
			Indent:        true,
			Fields: []gofakeit.Field{
				{Name: "first_name", Function: "firstname"},
				{Name: "last_name", Function: "lastname"},
				{Name: "password", Function: "password", Params: map[string][]string{"special": {"false"}}},
			},
		})
		if err != nil {
			return nil
		}
		return string(xml)
	case drivers.UUID:
		return gofakeit.UUID()
	case drivers.BinaryString:
		return binaryString(int(field.Length))
	case drivers.Unknown:
		log.Printf("unknown field type: %s\n", fieldDescriptor.Field)
		return nil
	}

	return nil
}

// randomString generates a length size random string
func randomString(length int16) string {
	var charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

func binaryString(length int) string {
	var str []string
	for i := 0; i < length; i++ {
		str = append(str, strconv.Itoa(gofakeit.Number(0, 1)))
	}
	return strings.Join(str, "")
}
