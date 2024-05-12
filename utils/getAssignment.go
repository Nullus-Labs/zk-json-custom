package utils

import (
	"bytes"
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"io/ioutil"
	"math/big"
	"os"
	"github.com/Nullus-Labs/zk-json-custom/circuit"
)

const MaxRecLen = 100
const BlkLen = 31

func getRawData(path string) Profile {
	rawData := Profile{}
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create a JSON decoder from the file
	decoder := json.NewDecoder(file)

	// Declare the data structure
	var data []map[string]interface{}

	// Decode the JSON data directly from the file
	if err := decoder.Decode(&data); err != nil {
		panic(err)
	}

	rawData.Profiles = data
	return rawData
}

func ReadJSON(name string) ([]byte, circuit.ProfileJSON) {
	var profile circuit.ProfileJSON
	// Read the JSON file
	data, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = json.Compact(&buf, data)
	if err != nil {
		panic(err)
	}
	// Unmarshal the JSON data into the struct
	err = json.Unmarshal(data, &profile)
	if err != nil {
		panic(err)
	}
	return buf.Bytes(), profile
}

func reverseEndian(input []byte) []byte {
	res := make([]byte, len(input))
	for i := 0; i < len(input); i++ {
		res[i] = input[len(input)-1-i]
	}
	return res
}

func EncryptRec(input []byte, key *fr.Element) []fr.Element {
	var res []fr.Element
	for i := 0; i < len(input); i += BlkLen {
		var end int
		if i+BlkLen > len(input) {
			end = len(input)
		} else {
			end = i + BlkLen
		}
		blk := new(fr.Element).SetBytes(reverseEndian(input[i:end]))
		res = append(res, circuit.EncryptMimcFr(*key, *blk))
	}
	return res
}

func getAssignment() circuit.EditCircuit {
	res := circuit.EditCircuit{}
	rawData := getRawData("files/rawdata_processed.json")
	oldEnc, oldProfile := ReadJSON("files/oldProfile_processed.json")
	newEnc, newProfile := ReadJSON("files/newProfile_processed.json")
	res.OldContent = makeProfile(oldProfile, rawData)
	res.NewContent = makeProfile(newProfile, rawData)
	res.Limit = makeLimit(rawData)

	encryptKey, _ := new(fr.Element).SetString("0x52fdfc072182654f163f5f0f9a621d729566c74d10037c4d")
	res.Key = encryptKey.BigInt(new(big.Int))
	res.CommittedKey = circuit.CommitMiMC(res.Key.(*big.Int).Bytes())
	oldRec := EncryptRec(oldEnc, encryptKey)
	newRec := EncryptRec(newEnc, encryptKey)
	res.OldRecord = make([]frontend.Variable, MaxRecLen)
	res.NewRecord = make([]frontend.Variable, MaxRecLen)

	for i := 0; i < MaxRecLen; i++ {
		if i < len(oldRec) {
			res.OldRecord[i] = oldRec[i]
		} else {
			res.OldRecord[i] = 0 //circuit.DUMMY
		}
		if i < len(newRec) {
			res.NewRecord[i] = newRec[i]
		} else {
			res.NewRecord[i] = 0 //circuit.DUMMY
		}
	}
	return res
}
