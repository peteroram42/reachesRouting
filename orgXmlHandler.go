package reachesRouting

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
)

type OrgsData struct {
	XMLName   xml.Name 	`xml:"orgs"`
	Entities  []Entity  `xml:"entity"`
}

type Entity struct {
	XMLName 		xml.Name 	`xml:"entity"`			
	Id    			string   	`xml:"id,attr"`			// ldnd
	FullName 		string 		`xml:"full_name"`		// Organization in London
	IncommId 		int 		`xml:"incomm_id"`		// 67
	EntityType 		string 		`xml:"type"`			// dissemorg, clvorg, msn, sosvc, testctr, cont, pubs, communityctr
	Address1 		string 		`xml:"address1"`		// 146 Queen Victoria Street
	City 			string 		`xml:"city"`			// London
	State 			string 		`xml:"state_province"`	// England
	PostCode 		string 		`xml:"postal_code"`		// EC4V 4BY
	Country     	string 		`xml:"country_code"`	// GB
	GeoAccuracy 	string 		`xml:"geo_accuracy"`	// ADDRESS
	Lat 			float64 	`xml:"latitude"`		// +051.512100
	Long 			float64 	`xml:"longitude"`		// -000.100816
	Lang 			string		`xml:"locale"`			// en_GB
	Cont 			string 		`xml:"cont"`			// af anzo can cis eu EU eus latam uk wus int  
	CsiOrgStatus 	string 		`xml:"csiorgstatus"`	// ideal, established, other
	Distance		float64 							// distance in miles, will be calculated elsewhere
	Determined		string								// marking how entity selected in, assigned elsewhere
}

func LoadOrgsXml(od *OrgsData) *OrgsData {

	xmlFile, err := os.Open("orgs.xml")
	if err != nil {
		fmt.Println(err)
	} 
	// close the file when done
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// assign and catch loading/parsing errors
	var err2 = xml.Unmarshal(byteValue, od)
	if err2 != nil {
		fmt.Println(err2)
	}

	return od

}