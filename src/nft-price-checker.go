package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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

type Account_collection struct {
	Slug              string  `json:"slug"`
	Owned_asset_count float64 `json:"owned_asset_count"`
}

// Reading input.txt file
func readInput(path string) ([]string, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var entries []string
	var nb []string
	line := 1
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if line == 1 {
			if scanner.Text() != "Collections:" {
				return askAdress(scanner.Text(), scanner.Err())
			}
		} else if line != 1 {
			entries = append(entries, scanner.Text()[:len(scanner.Text())-2])
			nb = append(nb, scanner.Text()[len(scanner.Text())-1:])
		}
		line++
	}

	return entries, nb, scanner.Err()
}

//Asking API for an account (address)
func askAdress(adr string, err error) ([]string, []string, error) {
	var entries []string
	var nb []string
	var cols []Account_collection

	url := fmt.Sprintf("https://api.opensea.io/api/v1/collections?asset_owner=%s&offset=0&limit=300", adr)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &cols)

	for _, result := range cols {
		entries = append(entries, result.Slug)
		nb = append(nb, fmt.Sprintf("%f", result.Owned_asset_count))
	}

	return entries, nb, err
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

// Fetching prices from the results of the API to build the final string
func getFloorPricesAndTotalValue(entries []string, nb []string) string {
	sum := 0.0
	result := ""
	estimate := ""
	for i, col := range entries {
		fp := floorPrice(col)
		fnb, _ := strconv.ParseFloat(nb[i], 64)
		prod := (fnb * fp)
		sum += prod
		if fp == 0 {
			result += ("  x " + col + " cannot be found on Opensea" + "\n")
		} else {
			result += ("--> " + col + " Floor price = " + fmt.Sprintf("%f", fp) + " eth\n So " + nb[i] + "*" + fmt.Sprintf("%f", fp) + "=" + fmt.Sprintf("%f", prod) + " eth\n")
		}
	}
	if sum > 0 {
		estimate = fmt.Sprintf("%f", sum) + " eth\n Or " + fmt.Sprintf("%f", sum*ethPrice()) + " Usd"
	} else {
		estimate = "0 eth\n Or 0 Usd"
	}
	result += ("------------- \n The estimate total value of your portfolio is : " + estimate)

	return result
}

// Writing into the result file
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

	var entries []string
	var nb []string
	var readErr error

	iexec_out := os.Getenv("IEXEC_OUT")
	iexec_in := os.Getenv("IEXEC_IN")
	iexec_input_file := os.Getenv("IEXEC_INPUT_FILE_NAME_1")
	dataset_file_name := os.Getenv("IEXEC_DATASET_FILENAME")

	if iexec_input_file != "" {
		entries, nb, readErr = readInput(iexec_in + "/" + iexec_input_file)
	} else if dataset_file_name != "" {
		entries, nb, readErr = readInput(iexec_in + "/" + dataset_file_name)
	} else {
		log.Fatal("Input or Dataset files are missing, exiting")
	}
	if readErr != nil {
		log.Fatal(readErr)
	}

	// Append some results in /iexec_out/
	writeFile(iexec_out+"/result.txt", getFloorPricesAndTotalValue(entries, nb))

	// Declare everything is computed
	writeFile(iexec_out+"/computed.json", ("{ \"deterministic-output-path\" : \"" + iexec_out + "/result.txt\" }"))
}
