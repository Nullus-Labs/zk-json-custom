package circuit

import (
	"github.com/consensys/gnark/frontend"
)

func EditCheck(api frontend.API, OldRecord []frontend.Variable, NewRecord []frontend.Variable, limit Limit, commitedKey frontend.Variable, oldContent Profile, newContent Profile, Key frontend.Variable) {
	contentCheck(api, commitedKey, Key, oldContent, newContent, OldRecord, NewRecord, limit)
}

func contentCheck(api frontend.API, commitedKey frontend.Variable, Key frontend.Variable, oldContent Profile, newContent Profile, oldRecord []frontend.Variable, newRecord []frontend.Variable, limit Limit) {
	compareContent(api, oldContent, newContent, limit)
	api.AssertIsEqual(commitedKey, commit(api, Key))

	encodedOldContent := encodeProfile(api, oldContent)
	assertArrayEqualWithUnequalLength(api, oldRecord, encrypt(api, Key, encodedOldContent))

	encodedNewContent := encodeProfile(api, newContent)
	assertArrayEqualWithUnequalLength(api, newRecord, encrypt(api, Key, encodedNewContent))
}

func encodeProfile(api frontend.API, profile Profile) []frontend.Variable {
	var mergeList [][]frontend.Variable
	mergeList = encodeDict(api, toDict(api, profile, MaxKeyLen), mergeList)
	return batchMerge(api, mergeList)
}

func assertArrayEqualWithUnequalLength(api frontend.API, a []frontend.Variable, b []frontend.Variable) {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	padA := make([]frontend.Variable, maxLen-len(a))
	padB := make([]frontend.Variable, maxLen-len(b))
	for i := range padA {
		padA[i] = frontend.Variable(0)
	}
	for i := range padB {
		padB[i] = frontend.Variable(0)
	}
	a = append(a, padA...)
	b = append(b, padB...)

	numEqual := frontend.Variable(0)
	//api.Println(a...)
	//api.Println(b...)
	for i := 0; i < maxLen; i++ {
		numEqual = api.Add(numEqual, isEqual(api, a[i], b[i]))
	}

	api.AssertIsEqual(numEqual, maxLen)
}
