# Golang OpenSea's NFT Collections floor price checker
+total portfolio value estimation

### Build
First, let's build the docker image
```
docker build . --tag nft-price-checker
```

### Input
This solution reads a .json input file following this format :
```
{
    "ownerAddress": 0x_owner_address,
    "collections": [
        {
            "collectionId": "collection_id",
            "count":  number_of_asset_owned
        },
        ...
    ]
}
```
If you fill the `ownerAddress`, the dapp will directly ask Opensea for the collections and nft that you own, so you doesn't have to fill the `"collections"` part. It will then look like this :
```
{
    "ownerAddress": "0x01234567891012345678910123456789101234567891",
    "collections": []
}
```
If you want to look for specific collections, you can fill the `collection` part with the Opensea collection id (or slug) and the number of asset that you own from that collection. It will then look like this :
```{
    "ownerAddress": "",
    "collections": [
        {
            "collectionId": "nft-worlds",
            "count": 2
        },
        {
            "collectionId": "psychedelics-anonymous-genesis",
            "count": 1
        },
        {
            "collectionId": "iexec-nft",
            "count": 3
        }
    ]
}
```  
The collection id can be found in the url of the Opensea Collection Page  
ie : for https://opensea.io/collection/boredapeyachtclub, the id is `boredapeyachtclub`

### Run
It is possible to run the application localy to test it out before deploying :
(It is needed to put an input file inside `/tmp/iexec_in/` folder)
```
rm -rf /tmp/iexec_out && \
docker run \
    -v /tmp/iexec_in:/iexec_in \
    -v /tmp/iexec_out:/iexec_out \
    -e IEXEC_IN=/iexec_in \
    -e IEXEC_OUT=/iexec_out \
    -e IEXEC_INPUT_FILE_NAME_1=input.json \
    -e IEXEC_INPUT_FILES_NUMBER=1 \
    nft-price-checker
```
Once the execution ends, the result should be found in the folder
`/tmp/iexec_out`.
```
cat /tmp/iexec_out/result.txt
```

### Deploy
To deploy your app, follow the instructions on the IExec Documentation : https://docs.iex.ec/for-developers/your-first-app

Then, you can run your dApp with the `iexec app run` command (you can add as much parameters and options as you want, follow the SDK and CLI documentation to do so) :  
```
iexec app run --watch
```

### Confidential Computing and TEE
In order to benefit from the computation confidentiality offered by Trusted Execution Environnements, we first need to sconify our dApp.  

To do that, just run the `./sconify.sh` script.  
```
./sconify.sh
```
It will build a sconified docker image of the app, that you can deploy the same way as a Standard dApp (like you did before following the iExec documentation).  
The code will now run inside a private enclave.  

You just have to add the `--tag tee` option in your run command :
```
iexec app run --watch --tag tee
```

But moreover, you can also add layer of confidentiality by protecting your input and output data.

### Datasets
Following this documentation https://docs.iex.ec/for-developers/confidential-computing/sgx-encrypted-dataset, you will be able to encrypt your input file and then give your "secret" (encryption key) to the SMS (Secret Management Service). Like this, no one (except you) will be able to read what your input data was.

### End to End Encryption
Finally, in order to achieve End to End encryption, you can encrypt your result following this documentation https://docs.iex.ec/for-developers/confidential-computing/end-to-end-encryption
