package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

//API request storage structs
type OwnerCollection struct {
	Slug            string  `json:"slug"`
	OwnedAssetCount float64 `json:"owned_asset_count"`
}

type CollectionFloorPrice struct {
	Stats struct {
		Floor_price float64 `json:"floor_price"`
	}
}

type EthPrice struct {
	Ethereum struct {
		Usd float64 `json:"usd"`
	}
}

//Input file storage structs
type Collection struct {
	CollectionID string  `json:"collectionId"`
	Count        float64 `json:"count"`
}

type Input struct {
	OwnerAddress string       `json:"ownerAddress"`
	Collections  []Collection `json:"collections"`
}

func readInput(path string) (Input, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return Input{}, err
	}

	var inputFile Input
	json.Unmarshal(file, &inputFile)

	if inputFile.OwnerAddress != "" {
		return getCollectionsByWalletAdress(inputFile.OwnerAddress, err)
	}

	return inputFile, err
}

//get nft collections owned by a specific wallet address :
/*Exemple of API response :
[
	{
		"primary_asset_contracts": [
			{
				"address": _,
				"asset_contract_type": _,
				...
			}
		]
		"traits": {},
		"stats": {},
		...
		"slug": "boredapeyatchclub",
		...
		"owned_asset_count": 1
	},
	{
		"primary_asset_contracts": [
			{
				"address": _,
				"asset_contract_type": _,
				...
			}
		]
		"traits": {},
		"stats": {},
		...
		"slug": "coolcats",
		...
		"owned_asset_count": 2
	},
	...
]*/
func getCollectionsByWalletAdress(adr string, err error) (Input, error) {
	var inputFile Input
	var owner []OwnerCollection

	url := fmt.Sprintf("https://api.opensea.io/api/v1/collections?asset_owner=%s&offset=0&limit=300", adr)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &owner)

	for _, collection := range owner {
		inputFile.Collections = append(inputFile.Collections, Collection{collection.Slug, collection.OwnedAssetCount})
	}

	return inputFile, err
}

// Asking for Opensea's colections Floor Prices
func floorPrice(collection_name string) float64 {
	var colllectionFP CollectionFloorPrice

	url := fmt.Sprintf("https://api.opensea.io/api/v1/collection/%s/stats", collection_name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &colllectionFP)

	return colllectionFP.Stats.Floor_price
}

// Asking for Eth price
func ethPrice() float64 {
	var price EthPrice

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &price)

	return price.Ethereum.Usd
}

// Fetching prices from the results of the API to build the final string
func getFloorPricesAndTotalValue(inputFile Input) string {
	ethSum := 0.0
	result := ""
	usdEstimate := ""
	for _, inputCollection := range inputFile.Collections {
		inputCollectionFloorPrice := floorPrice(inputCollection.CollectionID)
		inputCollectionNumberOfAssets := inputCollection.Count
		ethCollectionTotalEstimate := (inputCollectionFloorPrice * inputCollectionNumberOfAssets)
		ethSum += ethCollectionTotalEstimate
		if inputCollectionFloorPrice == 0 {
			result += ("  x " + inputCollection.CollectionID + " cannot be found on Opensea" + "\n")
		} else {
			result += ("--> " + inputCollection.CollectionID + " Floor price = " + fmt.Sprintf("%f", inputCollectionFloorPrice) + " eth\n")
			result += (" So " + fmt.Sprintf("%f", inputCollectionNumberOfAssets) + "*" + fmt.Sprintf("%f", inputCollectionFloorPrice) + "=" + fmt.Sprintf("%f", ethCollectionTotalEstimate) + " eth\n")
		}
	}
	if ethSum > 0 {
		usdEstimate = fmt.Sprintf("%f", ethSum) + " eth\n Or " + fmt.Sprintf("%f", ethSum*ethPrice()) + " Usd"
	} else {
		usdEstimate = "0 eth\n Or 0 Usd"
	}
	result += ("------------- \n The estimate total value of your portfolio is : " + usdEstimate)

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

	var inputFile Input
	var readErr error

	iexec_out := os.Getenv("IEXEC_OUT")
	iexec_in := os.Getenv("IEXEC_IN")
	iexec_input_file := os.Getenv("IEXEC_INPUT_FILE_NAME_1")
	dataset_file_name := os.Getenv("IEXEC_DATASET_FILENAME")

	if iexec_input_file != "" {
		inputFile, readErr = readInput(iexec_in + "/" + iexec_input_file)
	} else if dataset_file_name != "" {
		inputFile, readErr = readInput(iexec_in + "/" + dataset_file_name)
	} else {
		log.Fatal("Input or Dataset files are missing, exiting")
	}
	if readErr != nil {
		log.Fatal(readErr)
	}

	// Append some results in /iexec_out/
	writeFile(iexec_out+"/result.txt", getFloorPricesAndTotalValue(inputFile))

	// Declare everything is computed
	writeFile(iexec_out+"/computed.json", ("{ \"deterministic-output-path\" : \"" + iexec_out + "/result.txt\" }"))
}
