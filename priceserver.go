package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	_ "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	setupDB()

	//readconfig() // get config file from /etc/cryptodatasource

	log.Println("Db setup done")
	go ticker()
	log.Println("ticker setup done")
	http.HandleFunc("/", handler)
	http.ListenAndServe("192.168.7.101:7071", nil)
}

func readConfig() {
	defconfig := "/etc/priceserver.yml"
	log.Println("Using configuration file: " + defconfig)

}

type scraperItem struct {
	Url       string
	Fallback  string
	FieldName string
	Frequency string
}

type generalConfig struct {
	Serverip         string
	Serverport       string
	PrometheusPrefix string
}

func ticker() {
	// for i in config coins {
	ticker := time.NewTicker(15 * time.Second) // get the scrape frequency from the config file
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				go priceTask() //pricetask(url, key)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	// } end for i
}

func priceTask() {
	// run every 15 seconds
	//	timer := 15000
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// array of coins to fetch here, then loop

	client := &http.Client{Transport: tr}
	response, err := client.Get("https://www.bitrue.com/api/v1/ticker/price?symbol=SGBUSDT")
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	record := PriceResponse{}
	err = json.Unmarshal(body, &record)
	if err != nil {
		fmt.Println(err)
	}
	now := strconv.FormatInt(time.Now().Unix(), 10)

	db, err := sql.Open("sqlite3", "./db/data.db")
	if err != nil {
		log.Println(err)
	}
	query := "INSERT INTO PRICES (`id`, `coin`, `price`, `ts`) VALUES " +
		"( null, '" + record.Symbol + "', '" + record.Price + "', '" + now + "');"
	log.Println(query)
	_, err = db.Exec(query)
	if err != nil {
		log.Println(err)
	}

}

type PriceResponse struct {
	Symbol string
	Price  string
	Ts     string
	Id     string
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "content-type")
	w.Header().Add("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	record := PriceResponse{}
	log.Printf(r.RequestURI)
	db, err := sql.Open("sqlite3", "./db/data.db")
	if err != nil {
		log.Println(err)
	}
	query := "select * from PRICES where `coin` = 'SGBUSDT' order by `ts` desc limit 1"
	res, _ := db.Query(query)
	if res.Next() {
		err := res.Scan(&record.Id, &record.Symbol, &record.Price, &record.Ts)
		if err != nil {
			log.Println(err)
		}
		res.Close()
	}
	//	output := "[{\"target\":\""+record.Symbol+"\",\"datapoints\":[[\""+record.Price+"\",\""+now+"\"]]}]"
	output := "priceserver_price{id=\"" + record.Symbol + "\"} " + record.Price + "\n"
	//	io.Copy(buf, response.Body)

	log.Printf("Output: " + output)

	fmt.Fprintf(w, output)
}
func setupDB() {
	os.MkdirAll("./db", 0755)
	if _, err := os.Stat("./db/data.db"); errors.Is(err, os.ErrNotExist) {
		os.Create("./db/data.db")
	}

	db, err := sql.Open("sqlite3", "./db/data.db")
	if err != nil {
		log.Println(err)
	}

	createTable := "CREATE TABLE IF NOT EXISTS PRICES(" +
		"`id` INTEGER PRIMARY KEY AUTOINCREMENT, " +
		"`coin` STRING NOT NULL, " +
		"`price` REAL NOT NULL, " +
		"`ts` BIGINT)"
	log.Println(createTable)
	_, err = db.Exec(createTable)

	if err != nil {
		log.Println(err)
	}

	err = db.Close()
	if err != nil {
		log.Println(err)
	}
	db.Close()

}
