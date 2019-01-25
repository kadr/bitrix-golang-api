package section

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	yaml "gopkg.in/yaml.v2"
)

// section - структура элемента
type Section struct {
	ID                uint64            `db:"ID" json:"id"`
	Code              nullString        `db:"CODE" json:"code"`
	Name              string            `db:"NAME" json:"name"`
	Picture           nullString        `db:"PICTURE" json:"picture"`
	Description       nullString        `db:"DESCRIPTION" json:"description"`
	XMLID             nullString        `db:"XML_ID" json:"xml_id"`
	IblockID          uint64            `db:"IBLOCK_ID" json:"iblock_id"`
	IblockSectionID   nullInt64         `db:"IBLOCK_SECTION_ID" json:"iblock_section_id"`
	Active            bitrixBool        `db:"ACTIVE" json:"active"`
	Sort              uint64            `db:"SORT" json:"sort"`
	DepthLevel        uint64            `db:"DEPTH_LEVEL" json:"depth_level"`
	SearchableContent nullString        `db:"SEARCHABLE_CONTENT" json:"searchable_content"`
	DateCreate        nullString        `db:"DATE_CREATE" json:"date_create"`
	CreatedBy         uint64            `db:"CREATED_BY" json:"created_by"`
	TimestampX        nullString        `db:"TIMESTAMP_X" json:"timestamp_x"`
	ModifiedBy        nullInt64         `db:"MODIFIED_BY" json:"modified_by"`
	Elements          []uint64          `json:"elements"`
	Meta              map[string]string `json:"meta"`
	Props             map[string]string `json:"props"`
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
		"SORT",
		"PICTURE",
		"DESCRIPTION",
		"SEARCHABLE_CONTENT",
		"DATE_CREATE",
		"CREATED_BY",
		"TIMESTAMP_X",
		"MODIFIED_BY",
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
	sectionID := requestURL[2]

	filter := map[string]map[string]string{
		"filter": map[string]string{
			"ID": sectionID,
		},
	}

	sections, err := getData(filter)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(sections[0])
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)
}

// InfoByCode - получение одной записи по Code
func InfoByCode(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	sectionCode := requestURL[2]

	filter := map[string]map[string]string{
		"filter": map[string]string{
			"CODE": sectionCode,
		},
	}

	sections, err := getData(filter)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(sections[0])
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

	sections, err := getData(filter)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(sections)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)
}

// GetProperties - получение свойств елемента
// func GetProperties(response http.ResponseWriter, request *http.Request) {
// 	requestURL := strings.Split(request.RequestURI, "/")
// 	sectionID, _ := strconv.ParseUint(requestURL[2], 10, 32)

// 	props, err := getProperties(uint(sectionID))
// 	if err != nil {
// 		response.WriteHeader(http.StatusBadRequest)
// 		response.Write([]byte(err.Error()))
// 		return
// 	}

// 	result, _ := json.Marshal(props)
// 	response.Header().Set("Content-Type", "application/json")
// 	response.WriteHeader(http.StatusOK)
// 	response.Write(result)

// }

func getData(filter map[string]map[string]string) (sections []*Section, errorMessage error) {
	var where, param string
	var args []interface{}

	conn, err := sqlx.Connect("mysql", mysqlConnectString)
	defer conn.Close()

	if err != nil {
		errorMessage = err
	}

	selectedFields := prepareSelect()
	table := " FROM `b_iblock_section` t "
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

	err = conn.Select(&sections, query, args...)
	if err != nil {
		errorMessage = err
		return
	}

	for _, section := range sections {
		section.Elements, err = getChildElements(conn, section.ID)
		if err != nil {
			errorMessage = err
			break
		}
		section.Meta, err = getMeta(conn, section.ID)
		if err != nil {
			errorMessage = err
			break
		}
		section.Props, err = getProperties(conn, section.ID, section.IblockID)
		if err != nil {
			errorMessage = err
			break
		}
	}

	return
}

func getChildElements(conn *sqlx.DB, sectionID uint64) (elements []uint64, errorMessage error) {
	query := "SELECT IBLOCK_ELEMENT_ID FROM b_iblock_section_element" +
		" WHERE IBLOCK_SECTION_ID = ?"
	err := conn.Select(&elements, query, sectionID)
	if err != nil {
		errorMessage = err
		return
	}

	return
}

func getMeta(conn *sqlx.DB, sectionID uint64) (meta map[string]string, errorMessage error) {
	query := "SELECT (SELECT CODE FROM b_iblock_iproperty ip WHERE ip.ID = si.IPROP_ID) AS meta_name," +
		" si.VALUE AS value" +
		" FROM b_iblock_section_iprop si" +
		" WHERE si.SECTION_ID = ?"

	rows, err := conn.Queryx(query, sectionID)
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
func getProperties(conn *sqlx.DB, sectionID uint64, iblockID uint64) (props map[string]string, errorMessage error) {
	fields := []string{
		"UF_ALT_LINK",
		"UF_BANNER",
		"UF_CODE",
		"UF_COMBUSTION",
		"UF_DESCRIPTION_ARCH",
		"UF_DESCRIPTION_B2B",
		"UF_EMAIL_TO",
		"UF_H1_ARCH",
		"UF_H1_B2B",
		"UF_KEYWORDS_ARCH",
		"UF_KEYWORDS_B2B",
		"UF_MAKE_OFFER_BETTER",
		"UF_PROMO_END",
		"UF_PROMO_START",
		"UF_PROPERTY_CODE",
		"UF_RULES_FILE",
		"UF_TITLE_ARCH",
		"UF_TITLE_B2B",
	}

	query := "SELECT " + strings.Join(fields, ", ") +
		" FROM `b_uts_iblock_" + strconv.FormatUint(iblockID, 10) + "_section` " +
		" WHERE VALUE_ID = ?"

	rows, err := conn.Queryx(query, sectionID)
	if err != nil {
		errorMessage = err
		return
	}
	props = make(map[string]string, 0)
	for rows.Next() {
		var altLink, banner, code, combustion, descriptionArch, descriptionB2B string
		var emailTO, h1Arch, h1B2B, keywordsArch, keywordsB2B, makeOfferBetter string
		var promoEnd, promoStart, propertyCode, rulesFile, titleArch, titleB2B string

		rows.Scan(&altLink, &banner, &code, &combustion, &descriptionArch, &descriptionB2B, &emailTO, &h1Arch, &h1B2B, &keywordsArch, &keywordsB2B, &makeOfferBetter, &promoEnd, &promoStart, &propertyCode, &rulesFile, &titleArch, &titleB2B)

		fmt.Println(sectionID, code, descriptionArch, descriptionB2B)
		props["alt_link"] = altLink
		props["banner"] = banner
		props["code"] = code
		props["combustion"] = combustion
		props["description_arch"] = descriptionArch
		props["description_b2b"] = descriptionB2B
		props["email_to"] = emailTO
		props["h1_arch"] = h1Arch
		props["h1_b2b"] = h1B2B
		props["keywords_arch"] = keywordsArch
		props["keywords_b2b"] = keywordsB2B
		props["make_offer_better"] = makeOfferBetter
		props["promo_end"] = promoEnd
		props["promo_start"] = promoStart
		props["property_code"] = propertyCode
		props["rules_file"] = rulesFile
		props["title_arch"] = titleArch
		props["title_b2b"] = titleB2B

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
