package modules

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/araddon/dateparse"
)

type Module interface {
	Name() string
	ParseFlags(args []string) error
	Run(apiClient *datadog.APIClient, ctx context.Context) error
}

type ToFromStruct struct {
	From int64
	To   int64
}

func getToFrom(from string, to string) (int64, int64, error) {
	// Parse the from and to times, save them as Unix timestamps
	fromTime, err := dateparse.ParseAny(from)
	if err != nil {
		return 0, 0, err
	}
	toTime, err := dateparse.ParseAny(to)
	if err != nil {
		return 0, 0, err
	}
	fmt.Printf("Start time: %s\n", fromTime.Format(time.RFC3339))
	fmt.Printf("End time: %s\n", toTime.Format(time.RFC3339))
	return int64(fromTime.Unix()), int64(toTime.Unix()), nil
}

func GetStructKeysAsCSV(s interface{}) string {
	// Convert the struct fields to a CSV string.
	// This is useful for writing a header row to a CSV file.
	var csv string
	v := reflect.ValueOf(s)
	for i := 0; i < v.NumField(); i++ {
		csv += fmt.Sprintf("%v,", v.Type().Field(i).Name)
	}
	return csv
}
