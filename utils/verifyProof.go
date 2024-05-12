package utils

import (
	"bytes"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"io/ioutil"
	"os"
	"path"
)

func VerifyProof() bool {
	proofData, _ := ioutil.ReadFile("utils/proof/proof.txt")
	witnessData, _ := ioutil.ReadFile("utils/proof/witness.txt")

	proofReader := bytes.NewReader(proofData)
	proof := groth16.NewProof(ecc.BN254)
	_, err := proof.ReadFrom(proofReader)
	if err != nil {
		panic(err)
	}

	pubWitness, _ := witness.New(ecc.BN254.ScalarField())
	err = pubWitness.UnmarshalBinary(witnessData)
	if err != nil {
		panic(err)
	}

	vk := groth16.NewVerifyingKey(ecc.BN254)

	fvk, err2 := os.Open(path.Join(outputDir, "profile.vk"))
	if err2 != nil {
		panic(err2)
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(fvk)
	_, err3 := vk.ReadFrom(buf)
	if err3 != nil {
		panic(err3)
	}
	fvk.Close()

	err = groth16.Verify(proof, vk, pubWitness)
	if err != nil {
		panic(err)
	}
	return true

}
