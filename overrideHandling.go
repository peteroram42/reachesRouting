package reachesRouting

// import "fmt"

func handleOverride(loc *ReachLoc, rd *RoutingData, entList *OrgsData, campaign string, myWorld *World) *RoutingData {
	if campaign != "scn" {
		rd = handleOtherOverride(loc, rd, entList, campaign)
		return rd
	}
	switch loc.Country {
		// more detailed and specific logic
		case "AU":
			switch loc.State {
				case "Western Australia":
					rd = assignOrg_IfNgInRange(entList.GetEntById("ped"), rd)
					return rd
				default:
					rd = assignOrg_DoubleRange(loc, rd, entList)
					return rd
			}
		case "CA":
			switch loc.State {
				case "British Columbia":
					rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("van"), loc, rd, myWorld)
					return rd
				case "Alberta":
					rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("edm"), loc, rd, myWorld)
					return rd
				case "Ontario":
					rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("cam"), loc, rd, myWorld)
					return rd
				case "Quebec", "Québec":
					rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("qbc"), loc, rd, myWorld)
					return rd
				case "Saskatchewan", "Manitoba":
					rd = assignOrg_InCountry(loc, rd, entList, myWorld, "closest")
					return rd
				default:
					rd = assignOrg_InCountry(loc, rd, entList, myWorld, "closest")
					return rd
			}

		// go to specific org if n/g in range
		case "AE", "BH", "DZ", "EG", "IR", "IQ", "JO", "KW", "LY", "MA", "OM", "QA", "SA", "TN", "YE": 
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("ldnd"), loc, rd, myWorld)
			return rd
		case "AO", "CV", "GW", "MZ", "ST":
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("lis"), loc, rd, myWorld)
			return rd
		case "AZ", "BY", "GE", "LV":
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("moscow"), loc, rd, myWorld)
			return rd
		case "BJ", "BF", "BI", "CF", "CG", "CI", "DJ", "GA", "GN", "GQ", "KM", "MG", "ML", "NE", "RE", "SN", "TD", "TG", "YT":
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("par"), loc, rd, myWorld)
			return rd
		case "BW", "CM", "ER", "ET", "GM", "GH", "KE", "LS", "LR", "MW", "MR", "MU", "NA", "NG", "RW", "SC", "SD", "SL", "SO", "SZ", "TZ", "UG", "ZM":
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("jbgd"), loc, rd, myWorld)
			return rd
		case "EE": 
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("stpete"), loc, rd, myWorld)
			return rd
		case "FI":
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("mal"), loc, rd, myWorld)
			return rd
		case "HR", "SK": 
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("budapest"), loc, rd, myWorld)
			return rd
		case "IL":
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("tav"), loc, rd, myWorld)
			return rd
		case "LB", "LT", "SY", "TR":
			rd = assignOrg_IfNgInCountryInRange(entList.GetEntById("brln"), loc, rd, myWorld)
			return rd

		// in country - whatever is closest - prefer ideal, org, established msn in ranges up to 200 miles, then whatever is closest
		case "AT", "GR", "MK", "PT", "RU", "UA":
			rd = assignOrg_InCountry(loc, rd, entList, myWorld, "closest")
			return rd
		case "ZW":
			rd = assignOrg_InCountry(loc, rd, entList, myWorld, "closest")
			return rd

		// in country - default logic
		case "BE", "DK", "ES", "NL":
			rd = assignOrg_InCountry(loc, rd, entList, myWorld, "default")
			return rd

		// multi countries - anybody in these contries, stays in these countries
		case "DE":
			theMultCountries := []string{"DK", "CH", "DE"}
			rd = assignOrg_InMultCountries(loc, rd, entList, myWorld, "default", theMultCountries)
			return rd
		case "KZ":
			theMultCountries := []string{"KZ", "RU"}
			rd = assignOrg_InMultCountries(loc, rd, entList, myWorld, "closest", theMultCountries)
			return rd

		// not to country
		case "PL":
			rd = assignOrg_NotToCountry(loc, rd, entList, myWorld, "default", "SE")
			return rd

		// sort by post-code
		case "FR":
			postCodeBits := []rune(loc.PostCode)
			prefixRunes := postCodeBits[:2]
			prefix := string(prefixRunes)
			
			frenchRouting := map[string][]string{ // for the rest of france outside paris
				"ccp":	{"78", "80", "62", "59", "60", "02", "57", "51", "50", "14", "61", "27", "28", "78", "92"},
				"par": 	{"91", "77", "45", "18", "58", "89", "10", "21", "52", "55", "54", "57", "67", "68", "88", "90", "25", "70"},
				"nice": {"07", "26", "05", "30", "84", "04", "13", "83", "06"},
				"cle":  {"03", "63", "23", "19", "46", "15", "12", "48", "43", "34", "81", "82", "11", "9", "66"},
				"ang":  {"29", "22", "56", "35", "53", "72", "41", "36", "37", "49", "44", "85", "86", "79", "87", "17", "16", "24", "40", "47", "32", "65", "64"},
				"msnbordeaux": {"33"},
			}
			parisRouting := map[string][]string{ // for inside paris itself
				"ccp":  {"75001", "75002", "75006", "75007", "75008", "75009", "75015", "75016", "75017", "75018"},
				"par":  {"75003", "75004", "75005", "75010", "75011", "75012", "75013", "75014", "75019", "75020"},
			}

			if prefix == "75" {
				assignOrg_ByPostCodeRange(rd, entList, loc.PostCode, parisRouting)
			} else {
				assignOrg_ByPostCodeRange(rd, entList, prefix, frenchRouting)
			}

			return rd

		// fail to default rule
		default:
			return rd
	}
	return rd
}

func assignOrg_ByPostCodeRange(rd *RoutingData, entList *OrgsData, prefix string, pcRange map[string][]string) *RoutingData {
	for k, v := range pcRange {
		if stringInSlice(prefix, v) {
			rd.Primary = entList.GetEntById(k)
			rd.Primary.Determined = "ByOverride"
			return rd
		}
	}
	return rd
}

func assignOrg_IfNgInRange(org Entity, rd *RoutingData) *RoutingData {
	if len(rd.ValidEntities) == 0 {
		rd.Primary = org
		rd.Primary.Determined = "ByOverride"
		return rd
	}
	return rd
}

func assignOrg_DoubleRange(loc *ReachLoc, rd *RoutingData, entList *OrgsData) *RoutingData {

	myEntList := []Entity{}
	for _,thisEnt := range entList.Entities {
		if thisEnt.Lat == 0.0 && thisEnt.Long == 0.0 { continue }
		thisEnt.Distance = LatLongDist(loc.Lat, loc.Long, thisEnt.Lat, thisEnt.Long)
		if thisEnt.Distance < 200 && stringInSlice(thisEnt.EntityType, []string{"clvorg", "testctr", "msn", "communityctr"}) {
			thisEnt.Determined = "ByOverride"
			myEntList = append(myEntList, thisEnt)
		}
	}
	// going to add closest and closest ideal org in case even w/ 2 degrees there's nothing
	myEntList = append(myEntList, loc.findClosest(entList.Entities, []string{"clvorg", "testctr", "msn", "communityctr"}))
	myEntList = append(myEntList, loc.findClosestIO(entList.Entities))
	rd.Primary = GetPrimaryFromEnts(myEntList, "default")
	rd.Primary.Determined = "ByOverride"
	return rd
}

func assignOrg_IfNgInCountryInRange(org Entity, loc *ReachLoc, rd *RoutingData, myWorld *World) *RoutingData {
	// if n/g is in range at all assign and be done with it
	if len(rd.ValidEntities) == 0 {
		rd.Primary = org
		rd.Primary.Determined = "ByOverride"
		return rd
	} else { // stuff in range, lets find out if it's in this country - if so, build a list of it 
		inCountryList := GetCountryList(loc.Country, rd.ValidEntities, myWorld)
		// check if anything that was in range was in country. If not, assign.
		if len(inCountryList) == 0 {
			rd.Primary = org
			rd.Primary.Determined = "ByOverride"
			return rd
		} else if len(inCountryList) == 1 { // only 1 thing in range in country, go to it
			rd.Primary = inCountryList[0]
			rd.Primary.Determined = "ByOverride"
			return rd
		} else { // multiple things in range in country - do routing logic on THAT list to get primary
			rd.Primary = GetPrimaryFromEnts(inCountryList, "default")
			rd.Primary.Determined = "ByOverride"
			return rd
		}
	}
	rd.Primary.Determined = "ByOverride"
	return rd
}

func assignOrg_InCountry(loc *ReachLoc, rd *RoutingData, entList *OrgsData, myWorld *World, pref string) *RoutingData {
	inCountryList := GetCountryList(loc.Country, entList.Entities, myWorld)
	myEntList := PopulateDistance(loc, inCountryList)
	rd.Primary = GetPrimaryFromEnts(myEntList, pref)
	rd.Primary.Determined = "ByOverride"
	return rd
}

func assignOrg_InMultCountries(loc *ReachLoc, rd *RoutingData, entList *OrgsData, myWorld *World, pref string, countries []string) *RoutingData {
	inCountriesList := []Entity{}
	theseCountries := []Entity{}
	for _,v := range countries {
		theseCountries = GetCountryList(v, entList.Entities, myWorld)
		for _,z := range theseCountries {
			inCountriesList = append(inCountriesList, z)	
		}
	}
	myEntList := PopulateDistance(loc, inCountriesList)
	rd.Primary = GetPrimaryFromEnts(myEntList, pref)
	rd.Primary.Determined = "ByOverride"
	return rd
}

func assignOrg_NotToCountry(loc *ReachLoc, rd *RoutingData, entList *OrgsData, myWorld *World, pref string, avoid string) *RoutingData {
	avoidThese := myWorld.GetStringsFor("SE")
	allowedList := []Entity{}
	for _,v := range rd.ValidEntities {
		if !(stringInSlice(v.Country, avoidThese)) {
			allowedList = append(allowedList, v)
		}
	}
	if len(allowedList) > 1 { 
		rd.Primary = GetPrimaryFromEnts(allowedList, pref)
		rd.Primary.Determined = "ByOverride"
		return rd
	} else if len(allowedList) == 1 {
		rd.Primary = allowedList[0]
		rd.Primary.Determined = "ByOverride"
		return rd
	} else if !(stringInSlice(rd.ClosestIdeal.Country, avoidThese))  {
		rd.Primary = rd.ClosestIdeal
		rd.Primary.Determined = "ByOverride"
		return rd
	} else if !(stringInSlice(rd.ClosestThing.Country, avoidThese))  {
		rd.Primary = rd.ClosestThing
		rd.Primary.Determined = "ByOverride"
		return rd
	} else {
		rd.Primary = FindClosestIdealNotInCountry(loc, entList, myWorld, "SE")
		rd.Primary.Determined = "ByOverride"
		return rd
	}
	return rd
}




// this is here for possible future extended functionality
func handleOtherOverride(loc *ReachLoc, rd *RoutingData, entList *OrgsData, campaign string) *RoutingData {
	// if this gets used for dfw, or landmarks or whatever else then that specific logic can go here or in another file with this function
	return rd
}