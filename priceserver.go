package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
	_ "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var config generalConfig
var dbPath string

func main() {
	var dbvar string
	flag.StringVar(&dbvar, "d", "/var/lib/priceserver/", "Location of sqlite3 database")
	var conf string
	flag.StringVar(&conf, "c", "/etc/priceserver.yml", "Location of configureation file")
	flag.Parse()
	fmt.Println("config file:", conf)
	fmt.Println("db path:", dbvar)

	readConfig(conf) //
	dbPath = dbvar
	setupDB(dbvar)

	log.Println("Db setup done")
	go ticker()
	log.Println("ticker setup done")
	http.HandleFunc("/", handler)
	http.ListenAndServe("192.168.7.101:7071", nil)
}

func check(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func readConfig(confFile string) {

	log.Println("Using configuration file: " + confFile)

	dat, err := os.ReadFile(confFile)
	check(err)
	//	fmt.Print(string(dat))

	t := generalConfig{}
	err = yaml.Unmarshal([]byte(dat), &t)

	config = t
	//	meh, _ := json.Marshal(t)
	//	log.Println(string(meh))	// dump data structure as JSON for debugging

}

type scraperItem struct {
	PairName       string `yaml:"PairName"`
	URL            string `yaml:"Url"`
	ScrapeInterval int    `yaml:"ScrapeInterval"`
	JSONKey        string `yaml:"JsonKey"`
	FallbackURL    string `yaml:"FallbackUrl"`
	FallbackKey    string `yaml:"FallbackKey"`
}

type generalConfig struct {
	Serverip        string `yaml:"serverip"`
	ServerPort      int    `yaml:"ServerPort"`
	DefaultInterval int    `yaml:"DefaultInterval"`
	PromPrefix      string `yaml:"PromPrefix"`
	DbPath          string
	Items           []scraperItem `yaml:"Items"`
}

func ticker() {
	for _, v := range config.Items {
		var interval time.Duration
		if v.ScrapeInterval > 0 {
			interval = time.Duration(v.ScrapeInterval)
		} else {
			interval = time.Duration(config.DefaultInterval)
		}
		// for i in config coins {
		ticker := time.NewTicker(interval * time.Second) // get the scrape frequency from the config file
		quit := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					//	go priceTask(url) //
					priceTask(v.URL, v.JSONKey, v.FallbackURL, v.FallbackKey)

				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}
}

func priceTask(url string, key string, fburl string, fbkey string) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	response, err := client.Get(url)
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

	db, err := sql.Open("sqlite3", dbPath+"/data.db")
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
	db, err := sql.Open("sqlite3", dbPath+"/data.db")
	if err != nil {
		log.Println(err)
	}
	var output string
	for _, v := range config.Items {
		query := "select * from PRICES where `coin` = '" + v.PairName + "' order by `ts` desc limit 1"
		res, _ := db.Query(query)
		if res.Next() {
			err := res.Scan(&record.Id, &record.Symbol, &record.Price, &record.Ts)
			if err != nil {
				log.Println(err)
			}
			res.Close()
		}
		output += config.PromPrefix + "_price{id=\"" + record.Symbol + "\"} " + record.Price + "\n"
	}

	log.Printf("Output: " + output)

	fmt.Fprintf(w, output)
}
func setupDB(dbVar string) {
	os.MkdirAll(dbVar, 0755)
	if _, err := os.Stat(dbVar + "/data.db"); errors.Is(err, os.ErrNotExist) {
		os.Create(dbVar + "/data.db")
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
