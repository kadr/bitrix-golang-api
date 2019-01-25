package basket

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

// Basket данные которые будем доставать из базы
type Basket struct {
	BasePrice       float64    `db:"BASE_PRICE" json:"base_price"`
	CanBuy          bitrixBool `db:"CAN_BUY" json:"can_buy"`
	Currency        string     `db:"CURRENCY" json:"currency"`
	CustomPrice     bitrixBool `db:"CUSTOM_PRICE" json:"custom_price"`
	DateInsert      string     `db:"DATE_INSERT" json:"date_insert"`
	Delay           bitrixBool `db:"DELAY" json:"delay"`
	DetailPageURL   string     `db:"DETAIL_PAGE_URL" json:"detail_page_url"`
	Discount        float64    `db:"DISCOUNT_PRICE" json:"discont"`
	FuserID         int        `db:"FUSER_ID" json:"fuser_id"`
	ID              int        `db:"ID" json:"id"`
	SectionName     nullString `db:"SECTION_NAME" json:"section_name"`
	Name            string     `db:"NAME" json:"name"`
	Notes           string     `db:"NOTES" json:"notes"`
	OrderID         nullInt64  `db:"ORDER_ID" json:"order_id"`
	Price           float64    `db:"PRICE" json:"price"`
	PriceTypeID     int        `db:"PRICE_TYPE_ID" json:"price_type_id"`
	ProductID       int        `db:"PRODUCT_ID" json:"product_id"`
	Quantity        int        `db:"QUANTITY" json:"quantity"`
	Reserved        bitrixBool `db:"RESERVED" json:"reserved"`
	ReserveQuantity nullInt64  `db:"RESERVE_QUANTITY" json:"reserved_quantity"`
	Sort            int        `db:"SORT" json:"sort"`
	VatIncluded     bitrixBool `db:"VAT_INCLUDED" json:"vat_include"`
	VatRate         float64    `db:"VAT_RATE" json:"vat_rate"`
	Weight          float64    `db:"WEIGHT" json:"weight"`
}

type nullInt64 struct {
	sql.NullInt64
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
		"b.BASE_PRICE",
		"b.CAN_BUY",
		"b.CURRENCY",
		"b.CUSTOM_PRICE",
		"b.DATE_INSERT",
		"b.DELAY",
		"b.DETAIL_PAGE_URL",
		"b.DISCOUNT_PRICE",
		"b.FUSER_ID",
		"b.ID",
		"b.NAME",
		"b.NOTES",
		"b.ORDER_ID",
		"b.PRICE",
		"b.PRICE_TYPE_ID",
		"b.PRODUCT_ID",
		"b.QUANTITY",
		"b.RESERVED",
		"b.RESERVE_QUANTITY",
		"b.SORT",
		"b.VAT_INCLUDED",
		"b.VAT_RATE",
		"b.WEIGHT",
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

// Items - получаем все записи по определенному пользователю
func Items(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	fuserID, _ := strconv.Atoi(requestURL[2])
	basket, err := getData(uint32(fuserID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	result, _ := json.Marshal(basket)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)

}

// Product - получаем запись по определенному продукту
func Product(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	fuserID, _ := strconv.Atoi(requestURL[2])
	productID, _ := strconv.Atoi(requestURL[4])
	basket, err := getData(uint32(fuserID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}
	var result []byte
	for index, item := range basket {
		if item.ProductID == productID {
			result, _ = json.Marshal(basket[index])
			break
		}
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(result)

}

// Count - получаем количество товаров в корзине пользователя
func Count(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	fuserID, _ := strconv.Atoi(requestURL[2])

	basket, err := getData(uint32(fuserID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}
	count := strconv.Itoa(len(basket))
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(count))

}

// Cost - получаем стоимость всех товаров в корзине
func Cost(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	fuserID, _ := strconv.Atoi(requestURL[2])

	basket, err := getData(uint32(fuserID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}
	var summ float64
	for _, item := range basket {
		summ += item.Price
	}
	cost := strconv.FormatFloat(summ, 'f', 2, 64)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(cost))

}

// Weight - получаем общий вес всех товаров в корзине
func Weight(response http.ResponseWriter, request *http.Request) {
	requestURL := strings.Split(request.RequestURI, "/")
	fuserID, _ := strconv.Atoi(requestURL[2])

	basket, err := getData(uint32(fuserID))
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}
	var summ float64
	for _, item := range basket {
		summ += item.Weight / 1000
	}
	weight := strconv.FormatFloat(summ, 'f', 2, 64)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(weight))
}

func getData(fuserID uint32) (basket []Basket, errorMessage error) {
	conn, err := sqlx.Connect("mysql", mysqlConnectString)
	defer conn.Close()

	if err != nil {
		errorMessage = err
	}
	sectionNameSelect := " (SELECT s.NAME FROM b_iblock_section s WHERE ID = " +
		"(SELECT IBLOCK_SECTION_ID FROM b_iblock_element e WHERE e.ID = b.PRODUCT_ID)" +
		")"
	selectStr := "select " +
		strings.Join(fields, ", ") +
		", " + sectionNameSelect +
		"AS SECTION_NAME from b_sale_basket b"
	where := " where b.FUSER_ID = ?"

	query := selectStr + where
	err = conn.Select(&basket, query, fuserID)

	if err != nil {
		errorMessage = err
	}

	return
}

// MarshalJSON MarshalJSON interface redefinition
func (r nullInt64) MarshalJSON() ([]byte, error) {
	if r.Valid {
		return json.Marshal(r.Int64)
	} else {
		return json.Marshal(0)
	}
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
