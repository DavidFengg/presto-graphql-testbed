package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/jmoiron/sqlx"
	_ "github.com/prestodb/presto-go-client/presto"
)

// Column ...
type Column struct {
	Column  string
	Type    string
	Extra   string
	Comment string
}

// Variant ...
type Variant struct {
	Chrom string
	Start int
	End   int
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

var db *sqlx.DB
var err error

func main() {
	// presto-go-client
	dsn := "http://user@docker.for.mac.localhost:8080?catalog=default&schema=test"
	db, err = sqlx.Open("presto", dsn)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Server is running")

	router := mux.NewRouter()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET"})
	origins := handlers.AllowedOrigins([]string{"*"})

	// routes
	router.HandleFunc("/", home).Methods("GET")
	router.HandleFunc("/count", getCount).Methods("GET")
	router.HandleFunc("/column", getColumn).Methods("GET")

	// endpoints to test presto connector with joins across different endpoints
	router.HandleFunc("/variant_start", getVariantStart).Methods("GET")
	router.HandleFunc("/variant_end", getVariantEnd).Methods("GET")

	router.HandleFunc("/variants", getVariants).Methods("GET")
	router.HandleFunc("/samples", getSamples).Methods("GET")
	router.HandleFunc("/samples/{id}", getSample).Methods("GET")

	router.HandleFunc("/test", test).Methods("GET")
	router.HandleFunc("/test2", test2).Methods("GET")
	router.HandleFunc("/test3", testJoin).Methods("GET")
	router.HandleFunc("/test3v2", testJoin2).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(headers, methods, origins)(router)))
}

// root url exposes the api schema of the rest api as a json response
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	b, _ := ioutil.ReadFile("./schema.json")
	rawIn := json.RawMessage(string(b))

	var objmap map[string]*json.RawMessage

	err := json.Unmarshal(rawIn, &objmap)
	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(objmap)
}

func getCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	result, err := db.Query("WITH a AS (SELECT variant_id, patient_id FROM mysql.var_db.calls), b AS (SELECT code.text, subject.referenceid FROM mongodb.fhir.conditions), c AS (SELECT variant_id, chrom, start, ref, alt, gene, aa_change FROM mysql.var_db.variants), d AS (SELECT a.*, b.* FROM a JOIN b ON a.patient_id = b.referenceid) SELECT count(*) FROM d JOIN c ON d.variant_id = c.variant_id")

	if err != nil {
		panic(err)
	}

	result.Next()

	var count int
	err = result.Scan(&count)

	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(count)
}

func getColumn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	result, err := db.Queryx("SHOW COLUMNS FROM mysql.var_db.calls")

	if err != nil {
		panic(err)
	}

	var columns = make([]string, 0)

	var results = make(map[string]interface{})

	for result.Next() {
		err = result.MapScan(results) // MapScan puts rows as a map (col name, value)

		if err != nil {
			panic(err)
		}

		columns = append(columns, results["Column"].(string)) // assertion to get the value of a certain key in the interface
		fmt.Println(results)
	}

	json.NewEncoder(w).Encode(columns)
}

func test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	result, err := db.Queryx("SELECT * FROM mysql.var_db.calls")

	if err != nil {
		panic(err)
	}

	defer result.Close()

	var data = make([]int64, 0)

	var results = make(map[string]interface{})

	for result.Next() {
		// cols, err := result.SliceScan()
		err = result.MapScan(results)

		if err != nil {
			panic(err)
		}

		// fmt.Println(cols)
		data = append(data, results["variant_id"].(int64))
	}

	// json.NewEncoder(w).Encode(data)
}

func test2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	id := make(map[string]interface{})

	// fetch a slice of a result
	err := db.Select(&id, "SELECT variant_id FROM mysql.var_db.calls")
	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(id)
}

// test join v2 utilizes MapScan from sqlx
func testJoin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	rows, err := db.Query("WITH a AS (SELECT variant_id, patient_id FROM mysql.var_db.calls), b AS (SELECT code.text, subject.referenceid FROM mongodb.fhir.conditions), c AS (SELECT variant_id, chrom, start, ref, alt, gene, aa_change FROM mysql.var_db.variants), d AS (SELECT a.*, b.* FROM a JOIN b ON a.patient_id = b.referenceid) SELECT c.*, d.* FROM d JOIN c ON d.variant_id = c.variant_id LIMIT 10")
	if err != nil {
		panic(err)
	}

	cols, _ := rows.Columns()
	result := make([]interface{}, 0)

	// loop through each row from query and add it to the slice
	for rows.Next() {
		columnPointers := make([]interface{}, len(cols))
		columns := make([]interface{}, len(cols))

		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		err := rows.Scan(columnPointers...)
		if err != nil {
			panic(err)
		}

		m := make(map[string]interface{})

		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		result = append(result, m)
	}

	// return slice as a json response
	json.NewEncoder(w).Encode(result)
}

func testJoin2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	field := r.URL.Query().Get("limit")

	queryString := fmt.Sprintf("WITH a AS (SELECT variant_id, patient_id FROM mysql.var_db.calls), b AS (SELECT code.text, subject.referenceid FROM mongodb.fhir.conditions), c AS (SELECT variant_id, chrom, start, ref, alt, gene, aa_change FROM mysql.var_db.variants), d AS (SELECT a.*, b.* FROM a JOIN b ON a.patient_id = b.referenceid) SELECT c.aa_change, c.alt, c.chrom, c.start, c.variant_id FROM d JOIN c ON d.variant_id = c.variant_id %s", selectLimit(field))

	rows, err := db.Queryx(queryString)
	if err != nil {
		panic(err)
	}

	var data = make([]interface{}, 0)
	var result = make(map[string]interface{})

	for rows.Next() {
		err = rows.MapScan(result)

		data = append(data, result)

		result = make(map[string]interface{})
	}

	fmt.Println("Sending Response Body...")

	json.NewEncoder(w).Encode(data)

	fmt.Println("Finished")
	data = nil
}

func getVariants(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	// requires struct definition before complilation
	data := []Variant{}

	// presto has reserved keywords which need to be in double quotes if used as an identifier
	err := db.Select(&data, "SELECT chrom, start, \"end\" FROM mysql.var_db.variants")

	if err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(data)
}

// return json response of variant_id and end
func getVariantEnd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	queryString := fmt.Sprintf("SELECT variant_id, \"end\" FROM mysql.var_db.variants")

	rows, err := db.Queryx(queryString)
	if err != nil {
		panic(err)
	}

	var data = make([]interface{}, 0)
	var result = make(map[string]interface{})

	for rows.Next() {
		// map row to results dict/map
		err = rows.MapScan(result)
		// append results map to data array
		data = append(data, result)
		// clear result
		result = make(map[string]interface{})
	}

	// send data array as response
	json.NewEncoder(w).Encode(data)

	// clear data
	data = nil
}

// return json response of variant_id and start
func getVariantStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	queryString := fmt.Sprintf("SELECT variant_id, start FROM mysql.var_db.variants")

	rows, err := db.Queryx(queryString)
	if err != nil {
		panic(err)
	}

	var data = make([]interface{}, 0)
	var result = make(map[string]interface{})

	for rows.Next() {
		err = rows.MapScan(result)
		data = append(data, result)
		result = make(map[string]interface{})
	}

	json.NewEncoder(w).Encode(data)

	data = nil
}

func getSamples(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	fields, _ := r.URL.Query()["fields"]

	// build query string
	query := fmt.Sprintf("SELECT %s from mysql.var_db.samples", selectCols(fields...))

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	cols, _ := rows.Columns()
	result := make([]interface{}, 0)

	// loop through each row from query and add it to the slice
	for rows.Next() {
		columnPointers := make([]interface{}, len(cols))
		columns := make([]interface{}, len(cols))

		// slice of pointers to each column
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		err := rows.Scan(columnPointers...)
		if err != nil {
			panic(err)
		}

		obj := make(map[string]interface{})

		// map each value to object
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			obj[colName] = *val
		}

		result = append(result, obj)
	}

	// return slice as a json response
	json.NewEncoder(w).Encode(result)
}

func getSample(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enableCORS(&w)

	params := mux.Vars(r)
	fields, _ := r.URL.Query()["fields"]

	query := fmt.Sprintf("SELECT %s from mysql.var_db.samples WHERE sample_id=%s", selectCols(fields...), params["id"])

	// QueryRowx gets the first row of a query
	row := db.QueryRowx(query)

	res := make(map[string]interface{})
	err = row.MapScan(res)

	json.NewEncoder(w).Encode(res)
}

/* Build select Query
- db, table, fields
- static (perbuilt) join logic b/t db and tables
*/
func selectCols(fields ...string) string {
	// default to select all columns if fields query params is not specified
	if fields == nil {
		return "*"
	}

	return fields[0]
}

func selectLimit(val string) (limit string) {
	if val != "" {
		limit = "LIMIT " + val
	} else {
		limit = "LIMIT 1000"
	}

	return
}

/*
Data:
db joins (w/ tables) are static (comparing the same two id columns)

Queryx + MapScan: better if further processing is needed since rows are loaded upon each iteration
Select: loads all the data in memory (at once), must be mapped to a predefined struct
*/
