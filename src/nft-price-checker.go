package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
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

//Estimates structs
type Estimates struct {
	EthTotalEstimate    float64
	CollectionEstimates map[string]CollectionEstimate
}

type CollectionEstimate struct {
	FloorPrice  float64
	Count       float64
	EthEstimate float64
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

func readInput(path string) []Collection {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
		return []Collection{}
	}

	var jsonInput Input
	json.Unmarshal(file, &jsonInput)

	if jsonInput.OwnerAddress != "" {
		return getCollectionsByWalletAdress(jsonInput.OwnerAddress)
	}

	return jsonInput.Collections
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
func getCollectionsByWalletAdress(ownerAddress string) []Collection {
	var jsonInput Input
	var ownerCollection []OwnerCollection

	body := get(fmt.Sprintf("https://api.opensea.io/api/v1/collections?asset_owner=%s&offset=0&limit=300", ownerAddress))

	json.Unmarshal(body, &ownerCollection)

	for _, collection := range ownerCollection {
		jsonInput.Collections = append(jsonInput.Collections, Collection{collection.Slug, collection.OwnedAssetCount})
	}

	return jsonInput.Collections
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

	body := get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")

	json.Unmarshal(body, &price)

	return price.Ethereum.Usd
}

// Fetching prices from the results of the API to build the final string
func computeEstimates(inputCollections []Collection) Estimates {
	var estimates Estimates
	estimates.CollectionEstimates = make(map[string]CollectionEstimate)
	for _, inputCollection := range inputCollections {
		floorPrice := floorPrice(inputCollection.CollectionID)
		ethEstimate := floorPrice * inputCollection.Count

		estimates.CollectionEstimates[inputCollection.CollectionID] = CollectionEstimate{floorPrice, inputCollection.Count, ethEstimate}
		estimates.EthTotalEstimate += ethEstimate
	}

	return estimates
}

func estimatesToString(estimates Estimates, outputType string) string {
	usdEstimate := ethPrice() * estimates.EthTotalEstimate
	if outputType == "web2" {
		result := ""
		for collectionID, collectionEstimates := range estimates.CollectionEstimates {
			if collectionEstimates.FloorPrice == 0 {
				result += fmt.Sprintf("  x %s cannot be found on Opensea, or its floor price is equal to 0\n", collectionID)
			} else {
				result += fmt.Sprintf("--> %s Floor price = %f eth\n\t So %f*%f=%f eth\n",
					collectionID, collectionEstimates.FloorPrice,
					collectionEstimates.Count, collectionEstimates.FloorPrice, collectionEstimates.EthEstimate)
			}
		}
		result += fmt.Sprintf("------------- \n The estimate total value of your portfolio is : %f eth\n Or %f Usd", estimates.EthTotalEstimate, usdEstimate)

		return result
	} else {
		return hexutil.Encode([]byte(fmt.Sprintf("%f", usdEstimate)))
	}
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

	if len(os.Args) < 2 {
		log.Fatalln("Expecting one argument : \"web2\" or \"web3\"")
	}
	outputType := os.Args[1]
	if !(outputType == "web2" || outputType == "web3") {
		log.Fatalln("Args[1] needs to be either equal to \"web2\" or \"web3\"")
	}

	var inputCollections []Collection

	iexec_out := os.Getenv("IEXEC_OUT")
	iexec_in := os.Getenv("IEXEC_IN")
	iexec_input_file := os.Getenv("IEXEC_INPUT_FILE_NAME_1")
	dataset_file_name := os.Getenv("IEXEC_DATASET_FILENAME")

	if iexec_input_file != "" {
		inputCollections = readInput(iexec_in + "/" + iexec_input_file)
	} else if dataset_file_name != "" {
		inputCollections = readInput(iexec_in + "/" + dataset_file_name)
	} else {
		log.Fatalln("Input or Dataset files are missing, exiting")
	}
	result := estimatesToString(computeEstimates(inputCollections), outputType)
	if outputType == "web2" {
		// Append some results in /iexec_out/
		writeFile(iexec_out+"/result.txt", result)
		// Declare everything is computed
		writeFile(iexec_out+"/computed.json", ("{ \"deterministic-output-path\" : \"" + iexec_out + "/result.txt\" }"))
	} else {
		writeFile(iexec_out+"/computed.json", ("{ \"callback-data\" : \"" + result + "\" }"))
	}
}
