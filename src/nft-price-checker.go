package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Collection struct {
	Stats struct {
		Floor_price float64 `json:"floor_price"`
	}
}

type Price struct {
	Ethereum struct {
		Usd float64 `json:"usd"`
	}
}

// Reading input.txt file
func readInput(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entries = append(entries, scanner.Text())
	}
	return entries, scanner.Err()
}

// Asking for Opensea's colections Floor Prices
func floorPrice(collection_name string) float64 {
	var col1 Collection

	url := fmt.Sprintf("https://api.opensea.io/api/v1/collection/%s/stats", collection_name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &col1)

	return col1.Stats.Floor_price
}

// Asking for Eth price
func ethPrice() float64 {
	var pr1 Price

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &pr1)

	return pr1.Ethereum.Usd
}

func getFloorPricesAndTotalValue(entries []string) string {
	sum := 0.0
	result := ""
	for _, col := range entries {
		fp := floorPrice(col)
		sum += fp
		if fp == 0 {
			result += ("  x " + col + " cannot be found on Opensea" + "\n")
		} else {
			result += ("--> " + col + " Floor price = " + fmt.Sprintf("%f", fp) + " eth" + "\n")
		}
	}
	result += ("------------- \n The estimate total value of your portfolio is : " + fmt.Sprintf("%f", sum) + " eth\n Or " + fmt.Sprintf("%f", sum*ethPrice()) + " Usd")

	return result
}

func writeFile(file string, str string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(str + "\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	iexec_out := os.Getenv("IEXEC_OUT")
	iexec_in := os.Getenv("IEXEC_IN")
	iexec_input_file := os.Getenv("IEXEC_INPUT_FILE_NAME_1")

	entries, readErr := readInput(iexec_in + "/" + iexec_input_file)
	if readErr != nil {
		log.Fatal(readErr)
	}

	// Append some results in /iexec_out/
	writeFile(iexec_out+"/result.txt", getFloorPricesAndTotalValue(entries))

	// Declare everything is computed
	writeFile(iexec_out+"/computed.json", ("{ \"deterministic-output-path\" : \"" + iexec_out + "/result.txt\" }"))
}
