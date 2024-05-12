
package circuit

import (
	"github.com/consensys/gnark/frontend"
)





type Profile struct {
	Name String
	Age Integer
	Publications []struct{
		Title String
		Year Integer
	}
	Transcript struct{
	Details []struct{
		Course String
		Grade Integer
	}
    
	}
	
}










type ProfileJSON struct {
		Name string`json:"Name"`
	

		Age int64`json:"Age"`
	

	Publications []struct{
		Title string`json:"Title"`
	

		Year int64`json:"Year"`
	

	}
	Transcript struct{
	Details []struct{
		Course string`json:"Course"`
	

		Grade int64`json:"Grade"`
	

	}
    
	}
	
}





type Limit struct{
		
			Name []String
		
		
			Age []frontend.Variable
		
		
		
			Publications___Year []frontend.Variable
		
    
		
			Transcript___Details___Course []frontend.Variable
		
		
			Transcript___Details___Grade []frontend.Variable
		
    
	
}


type EditCircuit struct {
	OldRecord    []frontend.Variable `gnark:",public"`
	NewRecord    []frontend.Variable `gnark:",public"`
	Limit        Limit            `gnark:",public"`
	CommittedKey frontend.Variable   `gnark:",public"`
	OldContent   Profile
	NewContent   Profile
	Key          frontend.Variable
}


func (c *EditCircuit) Define(api frontend.API) error {
	EditCheck(api, c.OldRecord[:], c.NewRecord[:], c.Limit, c.CommittedKey, c.OldContent, c.NewContent, c.Key)
	return nil
}


















func compareContent (api frontend.API, oldContent Profile, newContent Profile, limit Limit){
		api.AssertIsEqual(1, checkOneOfSet(api, len(limit.Name), limit.Name[:], newContent.Name))



		api.AssertIsEqual(1, checkWithinRange(api,  limit.Age[0], limit.Age[1], newContent.Age))



		api.AssertIsEqual(1, checkAppendOnly(api, oldContent.Publications[:], newContent.Publications[:]))



		for _, iPublications := range newContent.Publications {



		api.AssertIsEqual(1, checkWithinRange(api,  limit.Publications___Year[0], limit.Publications___Year[1], iPublications.Year))



		}
    
		api.AssertIsEqual(1, checkAppendOnly(api, oldContent.Transcript.Details[:], newContent.Transcript.Details[:]))



		for _, iTranscript___Details := range newContent.Transcript.Details {
		api.AssertIsEqual(1, checkFormat(api, len(limit.Transcript___Details___Course), limit.Transcript___Details___Course, iTranscript___Details.Course))



		api.AssertIsEqual(1, checkWithinRange(api,  limit.Transcript___Details___Grade[0], limit.Transcript___Details___Grade[1], iTranscript___Details.Grade))



		}
    
		
	
}





