# Веб сервис для работы с базой битрикса
Структура:
- basket - работа с корзиной пользователя
- catalog - работа с каталогом
- element - работа с элементами инфоблока
- section - работа с разделами инфоблока
- env-example.yml - файл для хранения переменных окружения (переименовать в env.yml)
- main.go - рулит запросами через gorilla mux сервер
## Описание методов
### Basket
- Items - получаем все записи по определенному пользователю (GET /basket/{fuser_id:[0-9]+}/items/)
- Product - получаем запись по определенному продукту (GET /basket/{fuser_id:[0-9]+}/product/{product_id:[0-9]+}/)
- Count - получаем количество товаров в корзине пользователя (GET /basket/{fuser_id:[0-9]+}/count/)
- Cost - получаем стоимость всех товаров в корзине (GET /basket/{fuser_id:[0-9]+}/cost/)
- Weight - получаем общий вес всех товаров в корзине (GET /basket/{fuser_id:[0-9]+}/weight/)
### Catalog
- Info - достаем информацию по продукту (GET /catalog/{product_id:[0-9]+}/info/)
- HaveOffers - проверяем есть ли у продукта торговые предложения (GET /catalog/{product_id:[0-9]+}/have-offers/)
### Element
- InfoByID - получение одной записи по ID (GET /element/{element_id:[0-9]+}/info/)
- InfoByCode - получение одной записи по Code (GET /element/{element_code:[a-zA-Z-_0-9]+}/info/)
- List - достаем елементы по фильтру (POST /element/list/). По умолчанию 100 записей, если limit не задан
    
    Тело запроса:
    ```
    {
	    "filter": {
		    "IBLOCK_ID": "1",
		    "ACTIVE": "N",
		    "IBLOCK_SECTION_ID": "10"
	    }
        "params": {
            "LIMIT": 100,
            "ORDER": "SORT ASC",
            "GROUP": "ID"
        }
    }
    ```
- GetProperties - получение свойств елемента (GET /element/{element_id:[0-9]+}/props/)
### Section
- InfoByID - получение одной записи по ID (GET /section/{section_id:[0-9]+}/info/)
- InfoByCode - получение одной записи по Code (GET /section/{section_code:[a-zA-Z-_0-9]+}/info/)
- List - достаем елементы по фильтру (POST /section/list/). По умолчанию 100 записей, если limit не задан
    
    Тело запроса:
    ```
    {
	    "filter": {
		    "IBLOCK_ID": "1",
		    "ACTIVE": "N",
		    "IBLOCK_SECTION_ID": "10"
	    }
        "params": {
            "LIMIT": 100,
            "ORDER": "SORT ASC",
            "GROUP": "ID"
        }
    }
    ```