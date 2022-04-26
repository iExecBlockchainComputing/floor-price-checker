# Golang OpenSea's NFT Collections floor price checker
+total portfolio value estimation

### Build
First, let's build the docker image
```
docker build . --tag floor-price-checker
```

### Input
This aplication is reading an input file following this format :
(If you own 2 Nfts of the collection 1 and only 1 Nft from collection 2 and 3)
```
collection_id_1
collection_id_1
collection_id_2
collection_id_3
```
The collection id can be found in the url of the Opensea Collection Page  
ie : for https://opensea.io/collection/boredapeyachtclub, the id is ```boredapeyachtclub```

### Run
It's possible to run the application localy to test it out before deploying :
(It's need to put an input file inside ```/tmp/iexec_in/``` folder)
```
rm -rf /tmp/iexec_out && \
docker run \
    -v /tmp/iexec_in:/iexec_in \
    -v /tmp/iexec_out:/iexec_out \
    -e IEXEC_IN=/iexec_in \
    -e IEXEC_OUT=/iexec_out \
    -e IEXEC_INPUT_FILE_NAME_1=input \
    -e IEXEC_INPUT_FILES_NUMBER=1 \
    floor-price-checker
```
Once the execution ends, the result should be found in the folder
`/tmp/iexec_out`.
```
cat /tmp/iexec_out/result.txt
```

### Deploy
To deploy your app, follow the instructions on the IExec Documentation : https://docs.iex.ec/for-developers/your-first-app