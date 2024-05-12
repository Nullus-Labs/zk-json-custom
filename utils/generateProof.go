package utils

import (
	"bytes"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"io"
	"os"
	"path"
)

func GenertateProof() (groth16.Proof, witness.Witness) {
	cs := groth16.NewCS(ecc.BN254)

	f, _ := os.Open(path.Join(outputDir, "profile.r1cs"))
	_, err1 := cs.ReadFrom(f)
	if err1 != nil {
		panic(err1)
	}
	f.Close()

	pk := groth16.NewProvingKey(ecc.BN254)

	fpk, err2 := os.Open(path.Join(outputDir, "profile.pk"))
	if err2 != nil {
		panic(err2)
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(fpk)
	_, err3 := pk.ReadFrom(buf)
	if err3 != nil {
		panic(err3)
	}
	fpk.Close()

	assignment := getAssignment()
	witness, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		panic(err)

	}
	witnessPub, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	proof, err := groth16.Prove(cs, pk, witness)
	if err != nil {
		panic(err)
	}

	var proofBuffer bytes.Buffer
	var writer io.Writer = &proofBuffer
	proof.WriteTo(writer)
	proofBytes := proofBuffer.Bytes()

	//convert witness
	witnessBytes, _ := witnessPub.MarshalBinary()

	fileProof, err := os.Create(path.Join(outputDir, "proof.txt"))
	if err != nil {
		panic(err)
	}
	defer fileProof.Close()

	fileWit, err := os.Create(path.Join(outputDir, "witness.txt"))
	if err != nil {
		panic(err)
	}
	defer fileWit.Close()

	// Write bytes to file
	_, err = fileProof.Write(proofBytes)
	if err != nil {
		panic(err)
	}
	_, err = fileWit.Write(witnessBytes)
	if err != nil {
		panic(err)
	}
	//return proofBytes, witnessBytes

	return proof, witnessPub
}
