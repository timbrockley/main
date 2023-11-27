//--------------------------------------------------------------------------------

package sqlitedb

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/timbrockley/golang-main/file"
)

//--------------------------------------------------------------------------------

type SQLiteDBStruct struct {
	//----------
	FilePath    string
	Database    string
	DatabaseExt string
	//----------
	AutoCreate bool
	//----------
	DB *sql.DB
	//----------
}

//--------------------------------------------------------------------------------
//################################################################################
//--------------------------------------------------------------------------------

//--------------------------------------------------------------------------------
// Connect method
//--------------------------------------------------------------------------------

func (conn *SQLiteDBStruct) Connect() error {

	//--------------------------------------------------------------------------------
	var err error
	var filePath, database, databaseExt string
	//--------------------------------------------------------------------------------
	filePath = conn.FilePath
	database = conn.Database
	databaseExt = conn.DatabaseExt
	//--------------------------------------------------------------------------------
	if database != "" {

		//----------
		if databaseExt == "" {
			databaseExt = file.FilenameExt(database)
		}
		//----------
		if databaseExt != "" {
			database = file.FilenameBase(database)
		} else {
			databaseExt = "db"
		}
		//----------
		if filePath == "" {
			filePath = file.Path()
		}
		//----------
		filePath = file.FilePathJoin(filePath, database+"."+databaseExt)
		//----------

	} else {

		//----------
		if filePath == "" {
			//----------
			// runtime.Caller(0) = this script / runtime.Caller(1) = calling script
			_, filePath, _, _ = runtime.Caller(1)
			//----------
			filePath = file.FilePathBase(filePath) + ".db"
			//----------
		}
		//----------
		database = file.Filename(filePath)
		//----------
		if database != "" {
			if databaseExt == "" {
				databaseExt = file.FilenameExt(database)
			}
			if databaseExt != "" {
				database = file.FilenameBase(database)
			} else {
				databaseExt = "db"
			}
		}
		//----------
	}
	//--------------------------------------------------------------------------------
	conn.FilePath = filePath
	conn.Database = database
	conn.DatabaseExt = databaseExt
	//--------------------------------------------------------------------------------
	if !conn.AutoCreate && !file.FilePathExists(filePath) {

		return fmt.Errorf("database file %q does not exists ", filePath)
	}
	//--------------------------------------------------------------------------------
	conn.DB, err = sql.Open("sqlite3", filePath)
	//--------------------------------------------------------------------------------
	return err
	//--------------------------------------------------------------------------------
}

//--------------------------------------------------------------------------------
// Connect - interface to connect method
//--------------------------------------------------------------------------------
// conn, err = Connect(SQLiteDBStruct{ }, checkENV)
//--------------------------------------------------------------------------------

func Connect(conn SQLiteDBStruct) (SQLiteDBStruct, error) {

	//--------------------------------------------------------------------------------
	return conn, conn.Connect()
	//--------------------------------------------------------------------------------
}

//------------------------------------------------------------
// Exec method
//------------------------------------------------------------

func (conn *SQLiteDBStruct) Exec(query string, args ...any) (sql.Result, error) {
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

func (conn *SQLiteDBStruct) Query(query string, args ...any) (*sql.Rows, error) {
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

func (conn *SQLiteDBStruct) QueryRow(query string, args ...any) *sql.Row {
	//------------------------------------------------------------
	return conn.DB.QueryRow(strings.TrimSpace(query), args...)
	//------------------------------------------------------------
}

//--------------------------------------------------------------------------------
// Close method
//--------------------------------------------------------------------------------

func (conn *SQLiteDBStruct) Close() error {

	//--------------------------------------------------------------------------------
	var err error
	//--------------------------------------------------------------------------------
	if conn.DB != nil {
		err = conn.DB.Close()
	}
	//--------------------------------------------------------------------------------
	conn.DB = nil
	//--------------------------------------------------------------------------------
	return err
	//--------------------------------------------------------------------------------
}

//--------------------------------------------------------------------------------
//################################################################################
//--------------------------------------------------------------------------------

//------------------------------------------------------------
// GetSQLTableInfo method
//------------------------------------------------------------

func (conn *SQLiteDBStruct) GetSQLTableInfo(table string) (
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
	if table == "" {
		return nil, nil, errors.New("table cannot be blank")
	}
	//------------------------------------------------------------
	var err error
	var rows *sql.Rows
	//------------------------------------------------------------
	rows, err = conn.DB.Query("SELECT IFNULL(cid, 0)+1, IFNULL(name, ''), IFNULL(type, '') FROM PRAGMA_TABLE_INFO(?);", table)
	//------------------------------------------------------------
	var columInfoRows []struct {
		Sequence int
		Name     string
		Type     string
	}
	columnInfoMap := map[string]string{}
	//------------------------------------------------------------
	if err == nil {
		//------------------------------------------------------------
		defer rows.Close()
		//----------
		for rows.Next() {
			//----------
			var columInfoRow struct {
				Sequence int
				Name     string
				Type     string
			}
			//----------
			if err = rows.Scan(&columInfoRow.Sequence, &columInfoRow.Name, &columInfoRow.Type); err != nil {
				break
			}
			//----------
			columInfoRows = append(columInfoRows, columInfoRow)
			//----------
			columnInfoMap[columInfoRow.Name] = columInfoRow.Type
			//------------------------------------------------------------
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

func (conn *SQLiteDBStruct) GetTableInfo(table string) (
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
	if table == "" {
		return nil, nil, errors.New("table cannot be blank")
	}
	//------------------------------------------------------------
	var err error
	var rows *sql.Rows
	//------------------------------------------------------------
	rows, err = conn.DB.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 1;", table))
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

func (conn *SQLiteDBStruct) GetRowsInfo(rows *sql.Rows) (
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
		//------------------------------------------------------------
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
		//------------------------------------------------------------
	}
	//------------------------------------------------------------
	return columInfoRows, columnInfoMap, err
	//------------------------------------------------------------
}

//------------------------------------------------------------
// ScanRows method
//------------------------------------------------------------

func (conn *SQLiteDBStruct) ScanRows(sqlRows *sql.Rows) ([]map[string]any, error) {
	//------------------------------------------------------------
	var records []map[string]any
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
				//------------------------------------------------------------
				switch value.(type) {
				case int8:
					value = int(value.(int8))
				case int16:
					value = int(value.(int16))
				case int32:
					value = int(value.(int32))
				case int64:
					value = int(value.(int64))
				case uint:
					value = int(value.(uint))
				case uint8:
					value = int(value.(uint8))
				case uint16:
					value = int(value.(uint16))
				case uint32:
					value = int(value.(uint32))
				case uint64:
					value = int(value.(uint64))
				case float32:
					value = float64(value.(float32))
				}
				//----------
				record[columns[index]] = value
				//------------------------------------------------------------
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

func (conn *SQLiteDBStruct) QueryRecords(query string, args ...any) ([]map[string]any, error) {
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