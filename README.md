# ZK-JSON

Welcome to ZK-JSON!

## Setup
1. Visit [zkjson.com](https://zkjson.com) to upload your JSON file.
2. Follow the instructions on the website to provide information about your desired editing rules.
3. You will receive two files: `circuitTypes.go` and `rawdata_processed.json`.
  - `circuitTypes.go`: Contains the main body of your circuit.
  - `rawdata_processed.json`: Contains all information about the JSON type and corresponding editing rules.

## Complete Sample
Once you have `circuitTypes.go`, move this file to the `circuit` directory.

Move the `rawdata_processed.json` file to the `files` directory.

Additionally, place your pre-edited JSON file `oldProfile.json` and post-edited JSON file `newProfile.json` in the `files` directory.

## Run Program
1. Open a terminal and navigate to this repo.
2. **Build Project** First, build the Go project using the following commands:
    ```sh
    go build -o ZK-JSON
    chmod +x ZK-JSON 
    ```
4. **Setup Circuit:** Set up the circuit and generate the Rank-1 Constraint System, proving, and verifying keys by running the following command:
    ```sh
    ./ZK-JSON setup
    ```
5. **Generate Proof:** To generate the proof, run:
    ```sh
    ./ZK-JSON generateProof
    ```
   If you want to see the proof content, navigate to `utils/proof/proof.txt`.
6. **Verify Proof:** To verify the proof, run:
    ```sh
    ./ZK-JSON verifyProof
    ```
   If no errors occur during the verification process, the verification is successful.

