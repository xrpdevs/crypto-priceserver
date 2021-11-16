package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
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
	go startTickers()
	log.Println("startTickers setup done")
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(config.Serverip+":"+strconv.Itoa(config.ServerPort), nil)
	check(err)

}

func check(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func readConfig(confFile string) {

	log.Println("Using configuration file: " + confFile)

	dat, err := ioutil.ReadFile(confFile)
	check(err)
	//	fmt.Print(string(dat))

	t := generalConfig{}
	err = yaml.Unmarshal([]byte(dat), &t)
	check(err)
	config = t
	//	meh, _ := json.Marshal(t)
	//	log.Println(string(meh))	// dump data structure as JSON for debugging

}

func schedule(a string, b string, c string, d string, interval time.Duration, done <-chan bool, pairInfo string) *time.Ticker {
	log.Println("Checking " + pairInfo + " every " + strconv.FormatInt(int64(interval/1000/1000), 10) + "s")
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				priceTask(a, b, c, d)
			case <-done:
				return
			}
		}
	}()
	return ticker
}

func startTickers() {
	for _, v := range config.Items {
		done := make(chan bool)
		schedule(v.URL, v.JSONKey, v.FallbackURL, v.FallbackKey, time.Duration(v.ScrapeInterval)*time.Second, done, v.PairName)
	}
}

//type bgTasks struct {
//	Funct    func()
//	Interval int
//}

func priceTask(url string, key string, fburl string, fbkey string) {
	//log.Println(url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	response, err := client.Get(url)
	check(err)
	body, err := ioutil.ReadAll(response.Body)
	record := PriceResponse{}
	err = json.Unmarshal(body, &record)
	check(err)
	now := strconv.FormatInt(time.Now().Unix(), 10)

	db, err := sql.Open("sqlite3", dbPath+"/data.db")
	check(err)
	query := "INSERT INTO PRICES (`id`, `coin`, `price`, `ts`) VALUES " +
		"( null, '" + record.Symbol + "', '" + record.Price + "', '" + now + "');"
	log.Println(query)
	_, err = db.Exec(query)
	check(err)

	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "content-type")
	w.Header().Add("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	record := PriceResponse{}
	log.Printf(r.RequestURI)
	db, err := sql.Open("sqlite3", dbPath+"/data.db")
	check(err)

	var output string
	for _, v := range config.Items {
		query := "select * from PRICES where `coin` = '" + v.PairName + "' order by `ts` desc limit 1"
		res, _ := db.Query(query)
		if res.Next() {
			err := res.Scan(&record.Id, &record.Symbol, &record.Price, &record.Ts)
			check(err)

			err = res.Close()
			check(err)

		}
		output += config.PromPrefix + "_price{id=\"" + record.Symbol + "\"} " + record.Price + "\n"
	}

	log.Printf("Output: \n" + output)

	_, err = fmt.Fprintf(w, output)
	check(err)

}

func setupDB(dbVar string) {
	err := os.MkdirAll(dbVar, 0755)
	check(err)

	_, err = os.Stat(dbVar + "/data.db")
	if os.IsNotExist(err) {
		_, err := os.Create(dbVar + "/data.db")
		check(err)

	}

	db, err := sql.Open("sqlite3", dbVar+"/data.db")
	check(err)

	createTable := "CREATE TABLE IF NOT EXISTS PRICES (" +
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
	//	db.Close()

}

type PriceResponse struct {
	Symbol string
	Price  string
	Ts     string
	Id     string
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
