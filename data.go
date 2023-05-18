package uos

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB provides access to the Gorm database connection.
var DB *gorm.DB

func setupDataAccess() {
	Log.InfoContext("open SQLite database file", LogContext{"file": Config.Database.File})

	dbAccess, err := gorm.Open(
		sqlite.Open(Config.Database.File),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		Log.PanicError("could not initialize data access", err)
	}
	DB = dbAccess

	Log.Info("register framework models")
	RegisterDBModels(
		AppUser{},
	)
}

func cleanupDataAccess() {
	// do nothing
}

// RegisterDBModels executes Gorm auto-migration for all specified models.
// Determines (and caches) the list of DB columns of each model.
// Panics if anything fails.
func RegisterDBModels(models ...interface{}) {
	// migrate
	err := DB.AutoMigrate(models...)
	if err != nil {
		Log.PanicError("DB models auto-migration failed", err)
	}

	// analyze
	for _, m := range models {
		analyzeModel(m)
	}
}

type dbModelInfo struct {
	table string

	columns   []string
	columnMap map[string]string
}

var dbModelInfos = map[string]dbModelInfo{}

func analyzeModel(model interface{}) {
	s, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		Log.PanicError("could not analyze DB model", err)
	}

	info := dbModelInfo{
		table:     s.Table,
		columns:   []string{},
		columnMap: map[string]string{},
	}

	for _, field := range s.Fields {
		info.columns = append(info.columns, field.DBName)
		info.columnMap[field.DBName] = field.Name
	}

	Log.DebugContext("analyzed DB model", LogContext{"name": s.Name, "columns": info.columns})
	dbModelInfos[s.Name] = info
}

func dbColumns(name string) []string {
	return dbModelInfos[name].columns
}

func dbEntryCount(name string) (int64, error) {
	if name == "" {
		return -1, nil
	}

	var count int64
	return count, DB.Table(dbModelInfos[name].table).Where("deleted_at IS NULL").Count(&count).Error
}

// DBExtract returns a table of the specified columns extracted from the given list of models.
// The model must be registered using RegisterDBModels.
// Panics if "name" specifies an unknown model or "models" is not a slice.
func DBExtract(name string, models interface{}, columns []string) TableData {
	info, ok := dbModelInfos[name]
	if !ok {
		Log.PanicContext("model not registered", LogContext{"name": name})
	}

	list := reflect.ValueOf(models)
	if list.Kind() != reflect.Slice {
		Log.Panic("invalid call to DBExtract - models must be a slice")
	}

	table := make([][]interface{}, list.Len())
	for i := 0; i < list.Len(); i++ {
		table[i] = make([]interface{}, len(columns)+1)
		entry := list.Index(i).Interface()

		table[i][0] = reflect.ValueOf(entry).FieldByName(info.columnMap["id"]).Interface()
		for j, c := range columns {
			table[i][j+1] = reflect.ValueOf(entry).FieldByName(info.columnMap[c]).Interface()
		}
	}

	return table
}

// DBExtractForm fills form items with values based on the specified model.
// The model must be registered using Register DBModels.
// Panics if "name" specifies an unknown model.
func DBExtractForm(name string, model interface{}, items FormItems) FormItems {
	info, ok := dbModelInfos[name]
	if !ok {
		Log.PanicContext("model not registered", LogContext{"name": name})
	}

	for pos, i := range items {
		items[pos].Value = fmt.Sprintf(
			"%v",
			reflect.ValueOf(model).FieldByName(info.columnMap[i.Name]).Interface(),
		)
	}

	return items
}
