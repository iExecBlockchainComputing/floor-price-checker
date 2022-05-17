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

func get(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return body
}

func readInput(path string) ([]Collection, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return []Collection{}, err
	}

	var inputFile Input
	json.Unmarshal(file, &inputFile)

	if inputFile.OwnerAddress != "" {
		return getCollectionsByWalletAdress(inputFile.OwnerAddress, err)
	}

	return inputFile.Collections, err
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
func getCollectionsByWalletAdress(ownerAddress string, err error) ([]Collection, error) {
	var inputFile Input
	var owner []OwnerCollection

	body := get(fmt.Sprintf("https://api.opensea.io/api/v1/collections?asset_owner=%s&offset=0&limit=300", ownerAddress))

	json.Unmarshal(body, &owner)

	for _, collection := range owner {
		inputFile.Collections = append(inputFile.Collections, Collection{collection.Slug, collection.OwnedAssetCount})
	}

	return inputFile.Collections, err
}

// Asking for Opensea's colections Floor Prices
func floorPrice(collection_name string) float64 {
	var colllectionFP CollectionFloorPrice

	body := get(fmt.Sprintf("https://api.opensea.io/api/v1/collection/%s/stats", collection_name))

	json.Unmarshal(body, &colllectionFP)

	return colllectionFP.Stats.Floor_price
}

// Asking for Eth price
func ethPrice() float64 {
	var price EthPrice

	body := get(fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd"))

	json.Unmarshal(body, &price)

	return price.Ethereum.Usd
}

// Fetching prices from the results of the API to build the final string
func getFloorPricesAndTotalValue(inputCollections []Collection) string {
	ethSum := 0.0
	result := ""
	usdEstimate := ""
	for _, inputCollection := range inputCollections {
		inputCollectionFloorPrice := floorPrice(inputCollection.CollectionID)
		inputCollectionNumberOfAssets := inputCollection.Count
		ethCollectionTotalEstimate := (inputCollectionFloorPrice * inputCollectionNumberOfAssets)
		ethSum += ethCollectionTotalEstimate
		if inputCollectionFloorPrice == 0 {
			result += fmt.Sprintf("  x %s cannot be found on Opensea \n", inputCollection.CollectionID)
		} else {
			result += fmt.Sprintf("--> %s Floor price = %f eth\n", inputCollection.CollectionID, inputCollectionFloorPrice)
			result += fmt.Sprintf(" So %f*%f=%f eth\n", inputCollectionNumberOfAssets, inputCollectionFloorPrice, ethCollectionTotalEstimate)
		}
	}
	if ethSum > 0 {
		usdEstimate = fmt.Sprintf("%f eth\n Or %f Usd", ethSum, ethSum*ethPrice())
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
		log.Fatalln(err)
	}
	if _, err := f.Write([]byte(str + "\n")); err != nil {
		log.Fatalln(err)
	}
	if err := f.Close(); err != nil {
		log.Fatalln(err)
	}
}

func main() {

	var inputCollections []Collection
	var readErr error

	iexec_out := os.Getenv("IEXEC_OUT")
	iexec_in := os.Getenv("IEXEC_IN")
	iexec_input_file := os.Getenv("IEXEC_INPUT_FILE_NAME_1")
	dataset_file_name := os.Getenv("IEXEC_DATASET_FILENAME")

	if iexec_input_file != "" {
		inputCollections, readErr = readInput(iexec_in + "/" + iexec_input_file)
	} else if dataset_file_name != "" {
		inputCollections, readErr = readInput(iexec_in + "/" + dataset_file_name)
	} else {
		log.Fatalln("Input or Dataset files are missing, exiting")
	}
	if readErr != nil {
		log.Fatalln(readErr)
	}

	// Append some results in /iexec_out/
	writeFile(iexec_out+"/result.txt", getFloorPricesAndTotalValue(inputCollections))

	// Declare everything is computed
	writeFile(iexec_out+"/computed.json", ("{ \"deterministic-output-path\" : \"" + iexec_out + "/result.txt\" }"))
}
