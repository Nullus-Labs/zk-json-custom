package utils

import (
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/iancoleman/orderedmap"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	outputDir = "utils/proof"
)

func capitalizeKeys(data []byte) ([]byte, error) {
	omap := orderedmap.New()
	if err := json.Unmarshal(data, omap); err != nil {
		return nil, err
	}

	capitalizedMap := orderedmap.New()
	for _, k := range omap.Keys() {
		v, _ := omap.Get(k)
		capitalizedKey := strings.Title(k)
		if subMap, ok := v.(*orderedmap.OrderedMap); ok {
			subData, err := json.Marshal(subMap)
			if err != nil {
				return nil, err
			}
			modifiedSubData, err := capitalizeKeys(subData)
			if err != nil {
				return nil, err
			}
			newSubMap := orderedmap.New()
			if err := json.Unmarshal(modifiedSubData, newSubMap); err != nil {
				return nil, err
			}
			capitalizedMap.Set(capitalizedKey, newSubMap)
		} else {
			capitalizedMap.Set(capitalizedKey, v)
		}
	}

	return json.MarshalIndent(capitalizedMap, "", "  ")
}

// modifyJSON reads a JSON file, capitalizes the keys preserving the order, and writes the modified JSON back to a new file.
func modifyJSON(originalFilename, newFilename string) error {
	content, err := ioutil.ReadFile(originalFilename)
	if err != nil {
		return err
	}

	modifiedContent, err := capitalizeKeys(content)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(newFilename, modifiedContent, 0644); err != nil {
		return err
	}

	return nil
}

func Setup() {
	err := modifyJSON("files/newProfile.json", "files/newProfile_processed.json")
	if err != nil {
		panic(err)
	}
	err = modifyJSON("files/oldProfile.json", "files/oldProfile_processed.json")
	if err != nil {
		panic(err)
	}

	circ := getAssignment()
	cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circ)
	if err != nil {
		panic(err)
	}

	pk, vk, err := groth16.Setup(cs)
	f, err := os.Create(path.Join(outputDir, "profile.vk"))
	if err != nil {
		panic(err)
	}
	_, err = vk.WriteRawTo(f)
	if err != nil {
		panic(err)
	}

	fpk, err := os.Create(path.Join(outputDir, "profile.pk"))
	if err != nil {
		panic(err)
	}
	_, err = pk.WriteRawTo(fpk)
	if err != nil {
		panic(err)
	}

	fr, err := os.Create(path.Join(outputDir, "profile.r1cs"))
	if err != nil {
		panic(err)
	}
	_, err = cs.WriteTo(fr)
	if err != nil {
		panic(err)
	}
}
