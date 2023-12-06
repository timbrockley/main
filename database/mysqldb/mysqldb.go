//------------------------------------------------------------

package mysqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/timbrockley/golang-main/system"
)

//------------------------------------------------------------

type MySQLdbStruct struct {
	//----------
	Host string
	//----------
	User                 string
	Password             string
	AllowNativePasswords bool
	//----------
	Database string
	//----------
	AutoCreate bool
	//----------
	DB *sql.DB
	//----------
}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------

//------------------------------------------------------------
// init
//------------------------------------------------------------

func init() {
	//------------------------------------------------------------
	system.LoadENVs()
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Connect method
//------------------------------------------------------------

func (conn *MySQLdbStruct) Connect(checkENV ...bool) error {
	//------------------------------------------------------------
	var err error
	//------------------------------------------------------------
	// if checkENV passed as true then check environment variables
	//------------------------------------------------------------
	if checkENV != nil && checkENV[0] {
		//----------
		if conn.Database == "" {
			conn.Database = os.Getenv("MYSQL_DATABASE")
		}
		//----------
		conn.Host = os.Getenv("MYSQL_HOST")
		//----------
		conn.User = os.Getenv("MYSQL_USER")
		conn.Password = os.Getenv("MYSQL_PWD")
		//----------
		conn.AllowNativePasswords = os.Getenv("MYSQL_ALLOW_NATIVE_PASSWORDS") == "true"
		//----------
	}
	//------------------------------------------------------------
	mysqlConfig := mysql.Config{
		//----------
		User:   conn.User,
		Passwd: conn.Password,
		//----------
		AllowNativePasswords: conn.AllowNativePasswords,
		//----------
		// DBName: conn.Database,
		//----------
	}
	//------------------------------------------------------------
	if conn.Host != "" {
		//----------
		mysqlConfig.Addr = conn.Host
		mysqlConfig.Net = "tcp"
		//----------
	}
	//------------------------------------------------------------
	if conn.Database != "" && !CheckDatabaseName(conn.Database) {
		return errors.New("invalid database name")
	}
	//------------------------------------------------------------
	conn.DB, err = sql.Open("mysql", mysqlConfig.FormatDSN())
	//----------
	if err == nil {
		//----------
		err = conn.DB.Ping()
		//----------
		if err == nil && conn.Database != "" {
			//----------
			if conn.AutoCreate {
				//----------
				_, err = conn.DB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", conn.Database))
				//----------
			}
			//----------
			if err == nil {
				//----------
				_, err = conn.DB.Exec(fmt.Sprintf("USE %s;", conn.Database))
				//----------
			}
			//----------
		}
		//----------
	}
	//------------------------------------------------------------
	return err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Connect - interface to connect method
//------------------------------------------------------------
// conn, err = Connect(MySQLdbStruct{ }, checkENV)
//------------------------------------------------------------

func Connect(conn MySQLdbStruct, checkENV ...bool) (MySQLdbStruct, error) {
	//------------------------------------------------------------
	return conn, conn.Connect(checkENV...)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Exec method
//------------------------------------------------------------

func (conn *MySQLdbStruct) Exec(query string, args ...any) (sql.Result, error) {
	//------------------------------------------------------------
	if conn.DB == nil {
		return nil, errors.New("not connected")
	}
	//------------------------------------------------------------
	return conn.DB.Exec(strings.TrimSpace(query), args...)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Query method
//------------------------------------------------------------

func (conn *MySQLdbStruct) Query(query string, args ...any) (*sql.Rows, error) {
	//------------------------------------------------------------
	if conn.DB == nil {
		return nil, errors.New("not connected")
	}
	//------------------------------------------------------------
	return conn.DB.Query(strings.TrimSpace(query), args...)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// QueryRow method
//------------------------------------------------------------

func (conn *MySQLdbStruct) QueryRow(query string, args ...any) *sql.Row {
	//------------------------------------------------------------
	return conn.DB.QueryRow(strings.TrimSpace(query), args...)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// Close method
//------------------------------------------------------------

func (conn *MySQLdbStruct) Close() error {
	//------------------------------------------------------------
	var err error
	//------------------------------------------------------------
	if conn.DB != nil {
		err = conn.DB.Close()
	}
	//------------------------------------------------------------
	conn.DB = nil
	//------------------------------------------------------------
	return err
	//------------------------------------------------------------
}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------

//------------------------------------------------------------
// TableExists method
//------------------------------------------------------------

func (conn *MySQLdbStruct) TableExists(tableName string) (bool, error) {
	//------------------------------------------------------------
	if conn.DB == nil {
		return false, errors.New("not connected")
	}
	//----------
	if conn.Database == "" {
		return false, errors.New("database name cannot be blank")
	}
	//----------
	if !CheckDatabaseName(conn.Database) {
		return false, errors.New("invalid database name")
	}
	//----------
	if tableName == "" {
		return false, errors.New("table name cannot be blank")
	}
	//----------
	if !CheckTableName(tableName) {
		return false, errors.New("invalid table name")
	}
	//------------------------------------------------------------
	var err error
	var rows *sql.Rows
	//------------------------------------------------------------
	rows, err = conn.DB.Query(fmt.Sprintf("SELECT COUNT(*) AS count FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA='%s' AND TABLE_NAME='%s' LIMIT 1;", conn.Database, tableName))
	//------------------------------------------------------------
	if err != nil {
		return false, err
	}
	//------------------------------------------------------------
	defer rows.Close()
	//------------------------------------------------------------
	var count int
	//----------
	for rows.Next() {
		//----------
		err = rows.Scan(&count)
		//----------
	}
	//----------
	if err != nil {
		return false, err
	}
	//------------------------------------------------------------
	return count == 1, nil
	//------------------------------------------------------------
}

//------------------------------------------------------------
// GetSQLTableInfo method
//------------------------------------------------------------

func (conn *MySQLdbStruct) GetSQLTableInfo(tableName string) (
	[]struct {
		Sequence int
		Name     string
		Type     string
	},
	map[string]string,
	error) {
	//------------------------------------------------------------
	if conn.DB == nil {
		return nil, nil, errors.New("not connected")
	}
	//----------
	if conn.Database == "" {
		return nil, nil, errors.New("database name cannot be blank")
	}
	//----------
	if !CheckDatabaseName(conn.Database) {
		return nil, nil, errors.New("invalid database name")
	}
	//----------
	if tableName == "" {
		return nil, nil, errors.New("table name cannot be blank")
	}
	//----------
	if !CheckTableName(tableName) {
		return nil, nil, errors.New("invalid table name")
	}
	//------------------------------------------------------------
	var err error
	var rows *sql.Rows
	//------------------------------------------------------------
	var columInfoRows []struct {
		Sequence int
		Name     string
		Type     string
	}
	//----------
	columnInfoMap := map[string]string{}
	//------------------------------------------------------------
	rows, err = conn.DB.Query("SELECT IFNULL(ORDINAL_POSITION, 0), IFNULL(COLUMN_NAME, ''), IFNULL(DATA_TYPE, '') FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=? AND TABLE_NAME=?;", conn.Database, tableName)
	//------------------------------------------------------------
	if err == nil {
		//------------------------------------------------------------
		defer rows.Close()
		//------------------------------------------------------------
		for rows.Next() {
			//----------
			var Sequence int
			var Name string
			var Type string
			//----------
			if err = rows.Scan(&Sequence, &Name, &Type); err != nil {
				break
			}
			//----------
			columInfoRows = append(columInfoRows, struct {
				Sequence int
				Name     string
				Type     string
			}{Sequence, Name, Type})
			//----------
			columnInfoMap[Name] = Type
			//----------
		}
		//------------------------------------------------------------
	}
	//------------------------------------------------------------
	return columInfoRows, columnInfoMap, err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// GetTableInfo method
//------------------------------------------------------------

func (conn *MySQLdbStruct) GetTableInfo(tableName string) (
	[]struct {
		Sequence int
		Name     string
		Type     string
	},
	map[string]string,
	error) {
	//------------------------------------------------------------
	if conn.DB == nil {
		return nil, nil, errors.New("not connected")
	}
	//----------
	if conn.Database == "" {
		return nil, nil, errors.New("database name cannot be blank")
	}
	//----------
	if !CheckDatabaseName(conn.Database) {
		return nil, nil, errors.New("invalid database name")
	}
	//----------
	if tableName == "" {
		return nil, nil, errors.New("table name cannot be blank")
	}
	//----------
	if !CheckTableName(tableName) {
		return nil, nil, errors.New("invalid table name")
	}
	//------------------------------------------------------------
	var err error
	var rows *sql.Rows
	//------------------------------------------------------------
	rows, err = conn.DB.Query(fmt.Sprintf("SELECT * FROM %s.%s LIMIT 1;", conn.Database, tableName))
	//------------------------------------------------------------
	if err != nil {
		return nil, nil, err
	}
	//------------------------------------------------------------
	defer rows.Close()
	//------------------------------------------------------------
	return conn.GetRowsInfo(rows)
	//------------------------------------------------------------
}

//------------------------------------------------------------
// GetRowsInfo method
//------------------------------------------------------------

func (conn *MySQLdbStruct) GetRowsInfo(rows *sql.Rows) (
	[]struct {
		Sequence int
		Name     string
		Type     string
	},
	map[string]string,
	error) {
	//------------------------------------------------------------
	var err error
	var colTypes []*sql.ColumnType
	//------------------------------------------------------------
	var columInfoRows []struct {
		Sequence int
		Name     string
		Type     string
	}
	//----------
	columnInfoMap := map[string]string{}
	//------------------------------------------------------------
	colTypes, err = rows.ColumnTypes()
	//----------
	if err == nil {
		//----------
		for index, column := range colTypes {
			//----------
			Name := column.Name()
			Type := column.DatabaseTypeName()
			//----------
			columInfoRows = append(columInfoRows, struct {
				Sequence int
				Name     string
				Type     string
			}{index + 1, Name, Type})
			//----------
			columnInfoMap[Name] = Type
			//----------
		}
		//----------
	}
	//------------------------------------------------------------
	return columInfoRows, columnInfoMap, err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// ScanRows method
//------------------------------------------------------------

func (conn *MySQLdbStruct) ScanRows(sqlRows *sql.Rows) ([]map[string]any, error) {
	//------------------------------------------------------------
	var records []map[string]any
	//------------------------------------------------------------
	_, columnTypes, _ := conn.GetRowsInfo(sqlRows)
	//------------------------------------------------------------
	columns, err := sqlRows.Columns()
	//------------------------------------------------------------
	if err == nil {
		//------------------------------------------------------------
		for sqlRows.Next() {
			//------------------------------------------------------------
			scans := make([]any, len(columns))
			//----------
			record := make(map[string]any)
			//----------
			for i := range scans {
				scans[i] = &scans[i]
			}
			//----------
			sqlRows.Scan(scans...)
			//----------
			for index, value := range scans {
				//----------
				value = string(value.([]byte))
				//----------
				Name := columns[index]
				Type := strings.ToUpper(columnTypes[Name])
				//----------
				switch Type {
				case "BIGINT", "BIT", "BIT VARYING", "INT", "INTEGER", "MEDIUMINT", "SERIAL", "SMALLINT", "SMALLSERIAL", "TINYINT":
					value = system.ConvertToInt(value)
				case "DEC", "DECIMAL", "DOUBLE", "DOUBLE PRECISION", "FIXED", "FLOAT", "NUMERIC", "REAL":
					value = system.ConvertToFloat(value)
				case "BIGSERIAL", "BINARY", "BLOB", "BYTE", "BYTEA", "LONGBLOB", "TINYBLOB", "VARBINARY":
					value = system.ConvertToBytes(value)
				case "BOOL", "BOOLEAN":
					value = system.ConvertToBool(value)
				}
				//----------
				record[Name] = value
				//----------
			}
			//------------------------------------------------------------
			records = append(records, record)
			//------------------------------------------------------------
		}
		//------------------------------------------------------------
	}
	//------------------------------------------------------------
	return records, err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// QueryRecords method
//------------------------------------------------------------

func (conn *MySQLdbStruct) QueryRecords(query string, args ...any) ([]map[string]any, error) {
	//------------------------------------------------------------
	var err error
	var rows *sql.Rows
	//------------------------------------------------------------
	rows, err = conn.DB.Query(strings.TrimSpace(query), args...)
	//----------
	if err != nil {
		return nil, err
	}
	//------------------------------------------------------------
	defer rows.Close()
	//------------------------------------------------------------
	return conn.ScanRows(rows)
	//------------------------------------------------------------
}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------

func EscapeApostrophes(dataString string) string {
	//------------------------------------------------------------
	replacer := strings.NewReplacer(
		`'`, `''`,
	)
	//------------------------------------------------------------
	return replacer.Replace(dataString)
	//------------------------------------------------------------
}

//------------------------------------------------------------

func EscapeDoubleQuotes(dataString string) string {
	//------------------------------------------------------------
	replacer := strings.NewReplacer(
		`"`, `""`,
	)
	//------------------------------------------------------------
	return replacer.Replace(dataString)
	//------------------------------------------------------------
}

//------------------------------------------------------------

func EscapeMySQLString(dataString string) string {
	//------------------------------------------------------------
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\x00", "\\\x00",
		"\r", "\\r",
		"\n", "\\n",
		"\x1A", "\\Z",
		`"`, `\"`,
		`'`, `\'`,
	)
	//------------------------------------------------------------
	return replacer.Replace(dataString)
	//------------------------------------------------------------
}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------

func CheckDatabaseName(databaseName string) bool {
	//------------------------------------------------------------
	return CheckTableName(databaseName)
	//------------------------------------------------------------
}

//------------------------------------------------------------

func CheckTableName(tableName string) bool {
	//------------------------------------------------------------
	var err error
	var match bool
	//------------------------------------------------------------
	// should start with underscore or a letter
	//------------------------------------------------------------
	match, err = regexp.MatchString(`^[_A-Za-z]+`, tableName)
	//------------------------------------------------------------
	if err != nil || !match {
		return false
	}
	//------------------------------------------------------------
	// remaining characters should only contain underscores, letters or numbers
	//------------------------------------------------------------
	match, err = regexp.MatchString(`^[_A-Za-z0-9]*$`, tableName)
	//----------
	return err == nil && match
	//------------------------------------------------------------
}

//------------------------------------------------------------
//############################################################
//------------------------------------------------------------
