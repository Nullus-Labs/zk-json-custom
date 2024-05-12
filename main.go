package main

import (
	"os"
	"github.com/Nullus-Labs/zk-json-custom/utils"
)

func main() {
	order := os.Args[1]
	if order == "generateProof" {
		//will write the proof to the utils/proof folder
		utils.GenertateProof()
	} else if order == "setup" {
		//will write the setup stuff to the utils/proof folder
		utils.Setup()
	} else if order == "verifyProof" {
		//if not panic error, then verify is successful
		utils.VerifyProof()
	}
}
