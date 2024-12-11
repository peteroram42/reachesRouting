package reachesRouting

import ( 
	"errors"
	"strings"
	"fmt"
	"reachesRouting/geocoder"
)

const (
	offLineTesting = 1
)

// This is the incoming reach
type ReachLoc struct {  
	Lat 		float64  		// 34.101978 or 42.4048
	Long 		float64			// -18.327976 or -82.1910
	Street      string 			// Hollywood Blvd or King Street West
	Number		int 			// 6331 or 315
	City        string 			// Los Angeles or Chatham-Kent
 	State 		string			// California or Ontario - note that some places don't have a state
 	PostCode 	string			// 90028 or N7M 5K8
 	Country 	string			// United States or Canada
	FormattedAddress string 	// 6331 Hollywood Blvd, Los Angeles, CA 90028, USA or 
}								// 315 King St W, Chatham, ON N7M 5K8, Canada

// This is the final return structure
type RoutingData struct {
	Primary 		Entity 			// the one we're telling him to go to / routing to
	ClosestThing	Entity 			// the closest thing regardless of range, but in validTypes
	ClosestIdeal	Entity 			// the closest Ideal Org regardless of range, but in validTypes
	Aosh 			Entity			// closest ao regardless of how far
	Iorg 			Entity			// closest ideal org in range
	Org 			Entity 			// closest org in range
	Msn 			Entity			// closest msn in range
	AllEntities		[]Entity 		// all ents w/in range (+/-1 lat/long (~80-90 miles))
	ValidEntities	[]Entity 		// all ents matching validTypes in range
	OriginPoint		[]float64		// the lat/long we started from
}

// This is the full method to start from at least a lat/long and return a fully populated RoutingData struct
func (thisLocation *ReachLoc) GetRoutingData(campaign string) (*RoutingData, error) {
	// if don't have lat/long then die with error here as can't do anything with this 
	if thisLocation.Lat == 0.0 && thisLocation.Long == 0.0 {
		// can fix this so that if don't have lat long but have address then it'll go get it - last to do 
		// as not part of what's needed for internet/tv launch - feature request
		return new(RoutingData), errors.New("reachesRouting: Have to have at least a lat/long to work with")
	} 

	// load orgs.xml - do this here so that we can resuse popRoutingData with entSet in overrides
	od := OrgsData{}
	orgsXml := LoadOrgsXml(&od)

	// get the default data
	thisRoutingData := thisLocation.popRoutingData(orgsXml, campaign)

	// if we don't have state, postcode and country, go get it - we need this to evaluate override rules
	if offLineTesting == 0 {
		if thisLocation.State == "" || thisLocation.PostCode == "" || thisLocation.Country == "" {
			thisLocation.popByLatLong()
		}
	} else {
		// spoofed data if we are doing offline testing:
		thisLocation.State = "Queensland"
		thisLocation.PostCode = "75015"
		thisLocation.Country = "United States"
	}

	// now apply the override rules
	thisFinalRoutingData := checkOverrides(thisLocation, thisRoutingData, orgsXml, campaign)

	return thisFinalRoutingData, nil
}

// Initialize a RoutingData struct and populate with default data
func (thisLocation *ReachLoc) popRoutingData(entList *OrgsData, campaign string) *RoutingData {

	// define valid EntityTypes. Currently orgs.xml contains: dissemorg, clvorg, msn, sosvc, testctr, cont, pubs
	// in future could be used for dfw, cchr, twth, hr - or include landmarks if houses get added to orgs.xml, etc. 
	var validTypes []string
	switch campaign {
		case "scn":
			validTypes = append(validTypes, "clvorg", "testctr", "msn", "communityctr")
	}

	// builds out localEnts w/in 100 miles
	// builds out validEnts = subset of localEnts that contains types in validTypes
	localEnts := []Entity{}
	validEnts := []Entity{}
	for _,thisEnt := range entList.Entities {

		if thisEnt.Lat == 0.0 && thisEnt.Long == 0.0 { continue }

		thisEnt.Distance = LatLongDist(thisLocation.Lat, thisLocation.Long, thisEnt.Lat, thisEnt.Long)
		if thisEnt.Distance < 100 {
			thisEnt.Determined = "ByDefaultRules"
			localEnts = append(localEnts, thisEnt)
		
			if stringInSlice(thisEnt.EntityType, validTypes) {
				validEnts = append(validEnts, thisEnt)
			}
		}
	}

	// set closest Ideal org, closest Org, closest Msn
	placeHolderI := Entity{IncommId: -1, Distance: 999999.9}
	placeHolderO := Entity{IncommId: -1, Distance: 999999.9}
	placeHolderM := Entity{IncommId: -1, Distance: 999999.9}
	for _, v := range localEnts {
		if strings.TrimSpace(v.CsiOrgStatus) == "ideal" && (strings.TrimSpace(v.EntityType) == "clvorg" || strings.TrimSpace(v.EntityType) == "testctr" || strings.TrimSpace(v.EntityType) == "communityctr") {
			if placeHolderI.Distance > v.Distance {
				placeHolderI = v
			}
		} 
		if strings.TrimSpace(v.EntityType) == "clvorg" || strings.TrimSpace(v.EntityType) == "testctr" || strings.TrimSpace(v.EntityType) == "communityctr" {
			if placeHolderO.Distance > v.Distance {
				placeHolderO = v
			} 
		}
		if strings.TrimSpace(v.EntityType) == "msn" {
			if placeHolderM.Distance > v.Distance {
				placeHolderM = v
			}
		}
	}

	// determine closest ao
	placeHolderA := thisLocation.findClosestAO(entList.Entities)
	
	// get the closest thing
	// if we've got entities in range, then check through those
	placeHolderC := Entity{IncommId: -1, Distance: 999999.9}
	if len(validEnts) > 0 {
		for _, v := range validEnts {
			if v.Distance < placeHolderC.Distance {
				placeHolderC = v
			}
		}
	} else { // otherwise find wout what IS closest
		placeHolderC = thisLocation.findClosest(entList.Entities, validTypes)	
	}
	placeHolderC.Determined = "ByDefaultRules"

	// get the Closest Ideal Org
	// if it's in range, get that, otherwise find it wherever it is
	placeHolderCIO := Entity{IncommId: -1, Distance: 999999.9}
	if placeHolderI.IncommId != -1 { 
		placeHolderCIO = placeHolderI 
	} else {
		placeHolderCIO = thisLocation.findClosestIO(entList.Entities)
	}
	placeHolderCIO.Determined = "ByDefaultRules"

	// Assign Primary: If nothing in range
	// Go to closest Ideal Org regardless of distance
	// unless closest Org/Msn is < 120 miles and Ideal is > 250
	placeHolderP := Entity{IncommId: -1, Distance: 999999.9}
	if len(validEnts) == 0 {
		if placeHolderC.Distance < 120 && placeHolderCIO.Distance > 250 { 
			placeHolderP = placeHolderC
		} else {
			placeHolderP = placeHolderCIO
		}
	}

	// Assign Primary: If only 1 thing in range, go to it.
	// Go to it unless it's > 60 and closest Ideal < 100
	if len(validEnts) == 1 {
		if placeHolderC.Distance > 60 && placeHolderCIO.Distance < 100 {
			placeHolderP = placeHolderCIO
		} else {
			placeHolderP = placeHolderC
		}
	}
	
	// Assign Primary: Multiple things in range
	// If closest Ideal < 30, go to it
	// else if closest org or established msn < 30 go to it
	// else if closest ideal < 60, go to it
	// else if closest org or established msn < 60 go to it
	// else if ideal in range go to it
	// else if org in range go to it
	// else if msn in range go to it
	// else - if all else fails, go to the closest ideal org
	if len(validEnts) > 1 {
		if placeHolderCIO.Distance < 30 {
			placeHolderP = placeHolderCIO
		} else if placeHolderC.Distance < 30 && (placeHolderC.EntityType == "clvorg" || (placeHolderC.EntityType == "msn" && placeHolderC.CsiOrgStatus == "established")) {
			placeHolderP = placeHolderC
		} else if placeHolderCIO.Distance < 60 {
			placeHolderP = placeHolderCIO
		} else if placeHolderC.Distance < 60 && (placeHolderC.EntityType == "clvorg" || (placeHolderC.EntityType == "msn" && placeHolderC.CsiOrgStatus == "established")) {
			placeHolderP = placeHolderC
		} else if placeHolderI.IncommId > 0 {
			placeHolderP = placeHolderI
		} else if placeHolderO.IncommId > 0 {
			placeHolderP = placeHolderO
		} else if placeHolderM.IncommId > 0 {
			placeHolderP = placeHolderM
		}
	}

	if placeHolderP.IncommId == -1 {
		placeHolderP = placeHolderCIO
	}

	oPoint := []float64{thisLocation.Lat, thisLocation.Long}
	thisRoutingData := RoutingData{placeHolderP, placeHolderC, placeHolderCIO, placeHolderA,placeHolderI,placeHolderO,placeHolderM,localEnts, validEnts, oPoint}
	return &thisRoutingData
}

// Get state, postcode and country from lat/log 
func (thisLocation *ReachLoc) popByLatLong() *ReachLoc {
	newLoc := geocoder.Location{
		Latitude: thisLocation.Lat,
		Longitude: thisLocation.Long,
	}
	addresses, err := geocoder.GeocodingReverse(newLoc)
	if err != nil {
		fmt.Println("Could not get the addresses: ", err)
	} else {
		// Usually, the first address returned from the API
		// is more detailed, so let's work with it
		address := addresses[0]
		thisLocation.Street = address.Street
		thisLocation.Number = address.Number
		thisLocation.City = address.City
		thisLocation.State = address.State
		thisLocation.PostCode = address.PostalCode
		thisLocation.Country = address.Country
		thisLocation.FormattedAddress = address.FormattedAddress
	}
	return thisLocation
}

func checkOverrides(loc *ReachLoc, rd *RoutingData, entList *OrgsData, campaign string) (finalRouting *RoutingData) {

	myWorld := createWorld()
	overridesFor := []string{
		"AE", "AO", "AT", "AU", "AZ", "BE", "BF", "BH", "BI", "BJ", "BY", "BW", "CA", "CF", "CG", "CI", "CM", "CV", "DE", "DJ", "DK", "DZ", 
		"EE", "EG", "ER", "ES", "ET", "FR", "GA", "GE", "GH", "GM", "GN", "GR", "GQ", "GW", "HR", "IL", "IR", "IQ", "JO", "KE", 
		"LR", "LS", "KM", "KW", "KZ", "LB", "LT", "LV", "LY", "MA", "MG", "MK", "ML", "MR", "MU", "MW", "MZ", 
		"NA", "NE", "NG", "NL", "OM", "PL", "PT", "QA", "RE", "RU", "RW", "SA", "SC", "SD", "SK", "SL", "SN", "SO", "ST", "SY", "SZ", 
		"TD", "TG", "TN", "TR", "TZ", "UA", "UG", "YE", "YT", "ZM", "ZW", 
	}

	for _,v := range myWorld.Countries {
		if loc.Country == v.FullName || loc.Country == v.DuoCode || loc.Country == v.TriCode {
			loc.Country = v.DuoCode
			break
		}
	}

	finalRouting = rd
	// fmt.Println("Original Primary is: ",finalRouting.Primary)

	for z := range overridesFor {
		if loc.Country == overridesFor[z] {
			// fmt.Println("We're going to do an override!")
			finalRouting = handleOverride(loc, rd, entList, campaign, myWorld)
		}
	}

	return finalRouting
}

