package service

import (
	"dbclient"
	// "encoding/json"
	"database/sql"
	// "fmt"
	// cfg "cas-dicts/config"
	// "github.com/gorilla/mux"
	// "io"
	"baseinfo"
	// "io/ioutil"
	"log"
	// "net/http"
	// "os"
	"strconv"
	// "html/template"
	// "github.com/satori/go.uuid"
	// return uuid.NewV4().String()
	"reflect"
	"strings"
	"time"
	// "net"
)

var currentTable string = "dict"

type IDictsClient interface {
	dbclient.IMysqlClient

	GetDictById(id int64) (ret interface{})
	GetAllDicts(page, items int, dtype string, schoolId int64) (ret []interface{})
	InsertDict(username, name, desc, dtype string) (sql.Result, bool)
	UpdateDict(id int64, username, name, desc, dtype string) (sql.Result, bool)
	DelDictById(id int64) (sql.Result, bool)
	DelDictByIdReal(id int64) (sql.Result, bool)
	DelDicts(ids []int64) (sql.Result, bool)

	GetBaseInfo(username string) (int64, string, string)
}

type DictsClient struct {
	dbclient.MysqlClient
}

// Reflect all fields to map
func GetFieldMap(obj interface{}) (ret map[string]string) {
	val := reflect.ValueOf(obj).Elem()
	ret = make(map[string]string)
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		key := strings.ToLower(typeField.Name)
		if typeField.PkgPath != "" {
			// Private method
			continue
		} else {
			ret[key] = typeField.Name
		}
	}
	return
}

func (client *DictsClient) GetBaseInfo(username string) (int64, string, string) {
	return baseinfo.GetSchoolInfoFromUser(client.Db, username)
}

func formatResultSet(m map[string]string) interface{} {
	ret := map[string]interface{}{}
	log.Println("query dict return...", m)

	ret["id"], _ = strconv.ParseInt(m["id"], 10, 64)
	ret["name"] = m["key"]
	ret["description"] = m["desc"]
	ret["type"] = m["type"]
	ret["value"] = m["value"]
	ret["status"] = m["status"]
	if m["school_id"] == "0" {
		ret["school_id"] = ""
	} else {
		ret["school_id"] = m["school_id"]
	}

	return ret
}

func (client *DictsClient) GetDictById(id int64) (ret interface{}) {
	dbret := dbclient.Query(client.Db, "select * from "+currentTable+" s where s.id = ? and is_deleted = false", formatResultSet, id)
	if len(dbret) >= 1 {
		// ret = ret[:1]
		ret = dbret[0]
	} else {
		ret = map[string]string{}
	}
	return
}

func (client *DictsClient) GetAllDicts(page, items int, dtype string, schoolId int64) (ret []interface{}) {

	// sid, sname, utype := baseinfo.GetSchoolInfoFromUser(client.Db, "admin002")
	// log.Println("school info...", sid, sname, utype)

	clauses := " is_deleted = false "
	conds := []interface{}{}

	if dtype != "" {
		clauses = clauses + "and type = ? "
		conds = append(conds, dtype)
	}

	clauses = clauses + "and school_id = ? "
	conds = append(conds, schoolId)

	if page == 0 {
		ret = dbclient.Query(client.Db, "select * from "+currentTable+" where "+clauses, formatResultSet, conds...)
	} else {
		offset := (page - 1) * items
		conds = append(conds, strconv.Itoa(items), strconv.Itoa(offset))
		ret = dbclient.Query(client.Db, "select * from "+currentTable+" where "+clauses+" limit ? offset ? ", formatResultSet, conds...)
	}
	return
}

func (client *DictsClient) InsertDict(username, name, desc, dtype string) (sql.Result, bool) {
	sid, sname, stype := baseinfo.GetSchoolInfoFromUser(client.Db, username)

	if stype == "" {
		log.Println("there is no user or invalid user with no type")
		return nil, false
	}

	log.Println("baseinfo...", sid, sname, stype)

	sql, vals := dbclient.BuildInsert(currentTable, dbclient.ParamsPairs(
		"key", name,
		"desc", desc,
		"type", dtype,
		"school_id", sid,
		"is_deleted", false,
		"create_time", time.Now(),
	),
	)

	tx, err := client.Db.Begin()
	if err != nil {
		panic(err)
	}
	ret := dbclient.Exec(tx, sql, vals...)
	log.Println(ret)
	tx.Commit()
	return ret, true
}

// TODO...
func (client *DictsClient) InsertDicts(username, name, desc, dtype string) (sql.Result, bool) {
	sid, sname, stype := baseinfo.GetSchoolInfoFromUser(client.Db, username)

	if stype == "" {
		log.Println("there is no user or invalid user with no type")
		return nil, false
	}

	log.Println("baseinfo...", sid, sname, stype)

	sql, vals := dbclient.BuildInsert(currentTable, dbclient.ParamsPairs(
		"key", name,
		"desc", desc,
		"type", dtype,
		"school_id", sid,
		"is_deleted", false,
		"create_time", time.Now(),
	),
	)

	tx, err := client.Db.Begin()
	if err != nil {
		panic(err)
	}
	ret := dbclient.Exec(tx, sql, vals...)
	log.Println(ret)
	tx.Commit()
	return ret, true
}

func (client *DictsClient) UpdateDict(id int64, username, name, desc, dtype string) (sql.Result, bool) {
	sid, _, _ := baseinfo.GetSchoolInfoFromUser(client.Db, username)

	tx, err := client.Db.Begin()
	if err != nil {
		panic(err)
	}

	sql, vals := dbclient.BuildUpdate(currentTable, dbclient.ParamsPairs(
		"key", name,
		"desc", desc,
		"type", dtype,
		"school_id", sid,
	), dbclient.ParamsPairs(
		"id", id,
	),
	)

	ret := dbclient.Exec(tx, sql, vals...)
	log.Println(ret)
	tx.Commit()
	return ret, true
}

// TODO add schoolId check
func (client *DictsClient) DelDictById(id int64) (sql.Result, bool) {
	tx, err := client.Db.Begin()
	if err != nil {
		panic(err)
	}

	sql, vals := dbclient.BuildUpdate(currentTable, dbclient.ParamsPairs(
		"is_deleted", true,
	), dbclient.ParamsPairs(
		"id", id,
	),
	)

	ret := dbclient.Exec(tx, sql, vals...)
	log.Println(ret)
	tx.Commit()
	return ret, true
}

func (client *DictsClient) DelDictByIdReal(id int64) (sql.Result, bool) {
	tx, err := client.Db.Begin()
	if err != nil {
		panic(err)
	}

	sql, vals := dbclient.BuildDelete(currentTable, dbclient.ParamsPairs(
		"id", id,
	),
	)

	ret := dbclient.Exec(tx, sql, vals...)
	log.Println(ret)
	tx.Commit()
	return ret, true
}

func (client *DictsClient) DelDicts(ids []int64) (sql.Result, bool) {
	tx, err := client.Db.Begin()
	if err != nil {
		panic(err)
	}

	ids2str := make([]string, len(ids))
	for i, v := range ids {
		ids2str[i] = strconv.FormatInt(v, 10)
	}

	sql, vals := dbclient.BuildUpdateWithOpts(currentTable, dbclient.ParamsPairs(
		"is_deleted", true,
	), nil, nil,
		"id in "+"("+strings.Join(ids2str, ",")+")",
	)

	ret := dbclient.Exec(tx, sql, vals...)
	log.Println(ret)
	tx.Commit()
	return ret, true
}
