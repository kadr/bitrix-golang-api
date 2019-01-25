package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"./basket"
	"./catalog"
	"./delivery"
	"./element"
	"./section"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/kse/moscow/calc/", delivery.Provide).Methods("POST")
	router.HandleFunc("/kse/spb/calc/", delivery.Provide).Methods("POST")
	router.HandleFunc("/kse/moscow-obl/calc/", delivery.Provide).Methods("POST")
	router.HandleFunc("/kse/spb-obl/calc/", delivery.Provide).Methods("POST")
	router.HandleFunc("/basket/{fuser_id:[0-9]+}/items/", basket.Items).Methods("GET")
	router.HandleFunc("/basket/{fuser_id:[0-9]+}/product/{product_id:[0-9]+}/", basket.Product).Methods("GET")
	router.HandleFunc("/basket/{fuser_id:[0-9]+}/count/", basket.Count).Methods("GET")
	router.HandleFunc("/basket/{fuser_id:[0-9]+}/cost/", basket.Cost).Methods("GET")
	router.HandleFunc("/basket/{fuser_id:[0-9]+}/weight/", basket.Weight).Methods("GET")
	router.HandleFunc("/catalog/{product_id:[0-9]+}/info/", catalog.Info).Methods("GET")
	router.HandleFunc("/catalog/{product_id:[0-9]+}/have-offers/", catalog.HaveOffers).Methods("GET")
	router.HandleFunc("/element/{element_id:[0-9]+}/info/", element.InfoByID).Methods("GET")
	router.HandleFunc("/element/{element_code:[a-zA-Z-_0-9]+}/info/", element.InfoByCode).Methods("GET")
	router.HandleFunc("/element/list/", element.List).Methods("POST")
	router.HandleFunc("/element/{element_id:[0-9]+}/props/", element.GetProperties).Methods("GET")
	router.HandleFunc("/section/{section_id:[0-9]+}/info/", section.InfoByID).Methods("GET")
	router.HandleFunc("/section/{section_code:[a-zA-Z-_0-9]+}/info/", section.InfoByCode).Methods("GET")
	router.HandleFunc("/section/list/", section.List).Methods("POST")

	http.Handle("/", router)
	http.ListenAndServe(":9000", nil)
}
