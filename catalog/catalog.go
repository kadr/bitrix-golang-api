package catalog

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	yaml "gopkg.in/yaml.v2"
)

const (
	tableWithOffersProp = "b_iblock_element_prop_s4" // Таблица с соответствиями id торговых предложений и id товара
	cml2Link            = 3451                       // Ссылка на родительский объект для торговых предложений
)

// Catalog структура данных для каталога
type Catalog struct {
	ID               uint32      `db:"ID" json:"id"`
	Avaliable        bitrixBool  `db:"AVAILABLE" json:"available"`
	Quantity         int64       `db:"QUANTITY" json:"quantity"`
	QuantityReserved nullInt64   `db:"QUANTITY_RESERVED" json:"quantity_reserved"`
	Weight           nullFloat64 `db:"WEIGHT" json:"weight"`
	Width            nullFloat64 `db:"WIDTH" json:"width"`
	Length           nullFloat64 `db:"LENGTH" json:"length"`
	Height           nullFloat64 `db:"HEIGHT" json:"height"`
	Measure          nullInt64   `db:"MEASURE" json:"measure"`
	Type             string      `db:"TYPE" json:"type"`
	VatIncluded      bitrixBool  `db:"VAT_INCLUDED" json:"vat_included"`
	PriceType        string      `db:"PRICE_TYPE" json:"price_type"`
	WithoutOrder     bitrixBool  `db:"WITHOUT_ORDER" json:"without_order"`
	SelectBestPrice  bitrixBool  `db:"SELECT_BEST_PRICE" json:"select_best_price"`
	Price            nullFloat64 `db:"PRICE" json:"price"`
	Vat              nullFloat64 `db:"VAT_RATE" json:"vat"`
	Offers           []uint32    `json:"offers"`
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
		"p.ID",
		"p.AVAILABLE",
		"p.QUANTITY",
		"p.QUANTITY_RESERVED",
		"p.WEIGHT",
		"p.WIDTH",
		"p.LENGTH",
		"p.HEIGHT",
		"p.MEASURE",
		"p.TYPE",
		"p.VAT_INCLUDED",
		"p.PRICE_TYPE",
		"p.WITHOUT_ORDER",
		"p.SELECT_BEST_PRICE",
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

// Info - достаем информацию по продукту
func Info(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	productID, _ := strconv.Atoi(requestURL[2])

	catalog, err := getData(uint32(productID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(catalog)

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(result))
}

// HaveOffers - проверяем есть ли у продукта торговые предложения
func HaveOffers(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	productID, _ := strconv.Atoi(requestURL[2])

	catalog, err := getData(uint32(productID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	response.WriteHeader(http.StatusOK)
	if catalog.haveOffers() {
		response.Write([]byte("true"))
	} else {
		response.Write([]byte("false"))
	}

}

func getData(productID uint32) (catalog Catalog, errorMessage error) {
	conn, err := sqlx.Connect("mysql", mysqlConnectString)
	defer conn.Close()

	if err != nil {
		errorMessage = err
	}
	selectPrice := ", (SELECT PRICE FROM b_catalog_price WHERE PRODUCT_ID = p.ID) AS PRICE"
	selectVat := ", (SELECT RATE FROM b_catalog_vat WHERE ID = p.VAT_ID) AS VAT_RATE"
	selectStr := "SELECT " +
		strings.Join(fields, ", ") +
		selectPrice +
		selectVat +
		" FROM b_catalog_product p"
	where := " WHERE p.ID = ?"

	query := selectStr + where

	err = conn.Get(&catalog, query, productID)
	if err != nil {
		errorMessage = err
		return
	}

	offers, err := getOffers(productID, conn)
	if err != nil {
		errorMessage = err
		return
	}

	catalog.Offers = offers

	return
}

func getOffers(productID uint32, conn *sqlx.DB) (offers []uint32, errorMessage error) {
	query := "SELECT IBLOCK_ELEMENT_ID" +
		" FROM " + tableWithOffersProp +
		" WHERE PROPERTY_" + strconv.Itoa(cml2Link) + " = ?"

	err := conn.Select(&offers, query, productID)

	if err != nil {
		errorMessage = err
	}

	return
}

func (catalog *Catalog) haveOffers() (have bool) {
	if len(catalog.Offers) > 0 {
		have = true
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
