package element

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	yaml "gopkg.in/yaml.v2"
)

// Element - структура элемента
type Element struct {
	ID                uint64            `db:"ID" json:"id"`
	Code              nullString        `db:"CODE" json:"code"`
	Name              string            `db:"NAME" json:"name"`
	PreviewPicture    nullString        `db:"PREVIEW_PICTURE" json:"preview_picture"`
	DetailPicture     nullString        `db:"DETAIL_PICTURE" json:"detail_picture"`
	PreviewText       nullString        `db:"PREVIEW_TEXT" json:"preview_text"`
	DetailText        nullString        `db:"DETAIL_TEXT" json:"detail_text"`
	XMLID             nullString        `db:"XML_ID" json:"xml_id"`
	IblockID          uint64            `db:"IBLOCK_ID" json:"iblock_id"`
	IblockSectionID   nullInt64         `db:"IBLOCK_SECTION_ID" json:"iblock_section_id"`
	Active            bitrixBool        `db:"ACTIVE" json:"active"`
	ActiveFrom        nullString        `db:"ACTIVE_FROM" json:"active_from"`
	ActiveTo          nullString        `db:"ACTIVE_TO" json:"active_to"`
	Sort              uint64            `db:"SORT" json:"sort"`
	SearchableContent nullString        `db:"SEARCHABLE_CONTENT" json:"searchable_content"`
	DateCreate        nullString        `db:"DATE_CREATE" json:"date_create"`
	CreatedBy         uint64            `db:"CREATED_BY" json:"created_by"`
	TimestampX        nullString        `db:"TIMESTAMP_X" json:"timestamp_x"`
	ModifiedBy        nullInt64         `db:"MODIFIED_BY" json:"modified_by"`
	ShowCounter       nullInt64         `db:"SHOW_COUNTER" json:"show_counter"`
	Meta              map[string]string `json:"meta"`
}

// Properties - структура свойств
type Properties struct {
	Name  nullString `db:"Name" json:"name"`
	Code  nullString `db:"Code" json:"code"`
	Value nullString `db:"Value" json:"value"`
}

type nullInt64 struct {
	sql.NullInt64
}

type nullFloat64 struct {
	sql.NullFloat64
}

type nullString struct {
	sql.NullString
}

type bitrixBool struct {
	sql.NullString
}

// EnvStruct - структура для данных из env.yml файла
type EnvStruct struct {
	DBName     string `yaml:"db_name"`
	DBLogin    string `yaml:"db_login"`
	DBPassword string `yaml:"db_password"`
	DBHost     string `yaml:"db_host"`
	DBPort     int    `yaml:"db_port"`
}

var env EnvStruct
var fields []string
var mysqlConnectString string

func init() {
	fields = []string{
		"ID",
		"CODE",
		"XML_ID ",
		"NAME",
		"IBLOCK_ID",
		"IBLOCK_SECTION_ID",
		"ACTIVE",
		"ACTIVE_FROM",
		"ACTIVE_TO",
		"SORT",
		"PREVIEW_PICTURE",
		"PREVIEW_TEXT",
		"DETAIL_PICTURE",
		"DETAIL_TEXT",
		"SEARCHABLE_CONTENT",
		"DATE_CREATE",
		"CREATED_BY",
		"TIMESTAMP_X",
		"MODIFIED_BY",
		"SHOW_COUNTER",
	}
	fileEnv, err := ioutil.ReadFile("env.yml")
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(fileEnv, &env)
	if err != nil {
		log.Fatal(err)
	}

	mysqlConnectString = env.DBLogin + ":" + env.DBPassword +
		"@tcp(" + env.DBHost + ":" + strconv.Itoa(env.DBPort) + ")/" + env.DBName

}

// InfoByID - получение одной записи по ID
func InfoByID(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	elementID := requestURL[2]

	filter := map[string]map[string]string{
		"filter": map[string]string{
			"ID": elementID,
		},
	}

	elements, err := getData(filter)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(elements[0])
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)
}

// InfoByCode - получение одной записи по Code
func InfoByCode(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	elementCode := requestURL[2]

	filter := map[string]map[string]string{
		"filter": map[string]string{
			"CODE": elementCode,
		},
	}

	elements, err := getData(filter)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(elements[0])
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)
}

// List - достаем елементы по фильтру
func List(response http.ResponseWriter, request *http.Request) {
	// var context []interface{}
	var filter map[string]map[string]string

	body, _ := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	err := json.Unmarshal(body, &filter)

	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
	}

	elements, err := getData(filter)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(elements)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)
}

// GetProperties - получение свойств елемента
func GetProperties(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	elementID, _ := strconv.ParseUint(requestURL[2], 10, 32)

	props, err := getProperties(uint(elementID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(props)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)

}

func getData(filter map[string]map[string]string) (elements []*Element, errorMessage error) {
	var where, param string
	var args []interface{}

	conn, err := sqlx.Connect("mysql", mysqlConnectString)
	defer conn.Close()

	if err != nil {
		errorMessage = err
	}

	selectedFields := prepareSelect()
	table := " FROM `b_iblock_element` t "
	if filter, found := filter["filter"]; found {
		where, args = prepareFilter(filter)
	}
	if params, found := filter["params"]; found {
		param = prepareParams(params)
	}

	query := "SELECT " + selectedFields +
		table +
		where +
		param

	err = conn.Select(&elements, query, args...)
	if err != nil {
		errorMessage = err
		return
	}

	for _, element := range elements {
		element.Meta, err = getMeta(conn, element.ID)
		if err != nil {
			errorMessage = err
			break
		}
	}

	return
}

func getMeta(conn *sqlx.DB, elementID uint64) (meta map[string]string, errorMessage error) {
	query := "SELECT (SELECT CODE FROM b_iblock_iproperty ip WHERE ip.ID = ei.IPROP_ID) AS meta_name," +
		" ei.VALUE AS value" +
		" FROM b_iblock_iblock_iprop ei" +
		" WHERE ei.IBLOCK_ID = ?"

	rows, err := conn.Queryx(query, elementID)
	meta = make(map[string]string, 0)
	for rows.Next() {
		var metaName string
		var metaValue string
		err = rows.Scan(&metaName, &metaValue)
		meta[metaName] = metaValue
		if err != nil {
			errorMessage = err
		}
	}
	return
}

func prepareSelect() (query string) {
	var fieldsArray []string
	for _, field := range fields {
		f := strings.ToUpper(field)
		fieldsArray = append(fieldsArray, "`t`."+f)
	}
	query = strings.Join(fieldsArray, ", ")

	return
}

func prepareFilter(filter map[string]string) (where string, values []interface{}) {
	var whereTemp []string

	if id, ok := filter["ID"]; ok {
		whereTemp = append(whereTemp, "t.ID = ?")
		values = append(values, id)
	}
	if code, ok := filter["CODE"]; ok {
		whereTemp = append(whereTemp, "t.CODE = ?")
		values = append(values, code)
	}
	if active, ok := filter["ACTIVE"]; ok {
		whereTemp = append(whereTemp, "t.ACTIVE = ?")
		values = append(values, active)
	}
	if activeFrom, ok := filter[">ACTIVE_FROM"]; ok {
		whereTemp = append(whereTemp, "t.ACTIVE_FROM > ?")
		values = append(values, activeFrom)
	}
	if activeFrom, ok := filter["<ACTIVE_FROM"]; ok {
		whereTemp = append(whereTemp, "t.ACTIVE_FROM < ?")
		values = append(values, activeFrom)
	}
	if activeTo, ok := filter[">ACTIVE_TO"]; ok {
		whereTemp = append(whereTemp, "t.ACTIVE_TO > ?")
		values = append(values, activeTo)
	}
	if activeTo, ok := filter["<ACTIVE_TO"]; ok {
		whereTemp = append(whereTemp, "t.ACTIVE_TO < ?")
		values = append(values, activeTo)
	}
	if name, ok := filter["NAME"]; ok {
		whereTemp = append(whereTemp, "t.NAME = ?")
		values = append(values, name)
	}
	if name, ok := filter["NAME%"]; ok {
		whereTemp = append(whereTemp, "t.NAME LIKE ?%")
		values = append(values, name)
	}
	if name, ok := filter["%NAME"]; ok {
		whereTemp = append(whereTemp, "t.NAME LIKE %?")
		values = append(values, name)
	}
	if name, ok := filter["%NAME%"]; ok {
		whereTemp = append(whereTemp, "t.NAME LIKE %?%")
		values = append(values, name)
	}
	if xmlID, ok := filter["XML_ID"]; ok {
		whereTemp = append(whereTemp, "t.XML_ID = ?")
		values = append(values, xmlID)
	}
	if iblockID, ok := filter["IBLOCK_ID"]; ok {
		whereTemp = append(whereTemp, "t.IBLOCK_ID = ?")
		values = append(values, iblockID)
	}
	if iblockSectionID, ok := filter["IBLOCK_SECTION_ID"]; ok {
		whereTemp = append(whereTemp, "t.IBLOCK_SECTION_ID = ?")
		values = append(values, iblockSectionID)
	}

	where = " WHERE " + strings.Join(whereTemp, " AND ")

	return
}

func prepareParams(filter map[string]string) (params string) {
	if limit, ok := filter["LIMIT"]; ok {
		params += " LIMIT " + limit
	} else {
		params += " LIMIT 100"
	}
	if order, ok := filter["ORDER"]; ok {
		params += " ORDER BY " + order
	} else {
		params += " ORDER BY SORT ASC"
	}
	if group, ok := filter["GROUP"]; ok {
		params += " GROUP BY " + group
	}

	return
}

// getProperties - получение свойств элемента
func getProperties(elementID uint) (props []Properties, errorMessage error) {

	conn, err := sqlx.Connect("mysql", mysqlConnectString)
	defer conn.Close()

	query := "SELECT " +
		"`prop`.NAME AS Name, " +
		"`prop`.CODE AS Code, " +
		"IF (`prop`.PROPERTY_TYPE = 'F', IFNULL (CONCAT(`file`.SUBDIR, '/', `file`.FILE_NAME), ''), IFNULL(`prop_enum`.VALUE, `el_prop`.VALUE)) AS Value  " +
		"FROM `b_iblock_element` el " +
		"LEFT JOIN `b_iblock_element_property` el_prop ON `el_prop`.IBLOCK_ELEMENT_ID = `el`.ID " +
		"LEFT JOIN `b_iblock_property` prop ON `prop`.ID = `el_prop`.IBLOCK_PROPERTY_ID " +
		"LEFT JOIN `b_iblock_property_enum` prop_enum ON `prop_enum`.ID = `el_prop`.VALUE " +
		"LEFT JOIN `b_file` file ON `file`.ID = `el_prop`.VALUE " +
		"WHERE `el`.ID = ?"

	if err != nil {
		errorMessage = err
	}

	err = conn.Select(&props, query, elementID)
	if err != nil {
		errorMessage = err
		return
	}

	return
}

// MarshalJSON MarshalJSON interface redefinition
func (r nullInt64) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.Int64)
	}

	return json.Marshal(0)

}

// MarshalJSON MarshalJSON interface redefinition
func (r nullFloat64) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.Float64)
	}

	return json.Marshal(0)

}

// MarshalJSON MarshalJSON interface redefinition
func (r nullString) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.String)
	}

	return json.Marshal("")

}

// MarshalJSON MarshalJSON interface redefinition
func (r bitrixBool) MarshalJSON() ([]byte, error) {
	if r.String == "N" {
		return json.Marshal(false)
	}

	return json.Marshal(true)

}
