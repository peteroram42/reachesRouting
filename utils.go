package reachesRouting

import "strings"
// import "fmt"

func OrgLookupById(entSet []Entity, id string) (thisOne Entity) {
	for _, v := range entSet {
		if strings.TrimSpace(strings.ToLower(v.Id)) == id {
			thisOne = v
			return			
		}
	}
	return
}

func (od *OrgsData) GetEntById(org string) (thisOne Entity) {
	for _,v := range od.Entities {
		if v.Id == org {
			thisOne = v
			break
		}
	}
	return
}

func (op *ReachLoc) findClosestAO(entSet []Entity) (thisOne Entity) {
	var AOs []Entity
	for _,v := range entSet {
		if strings.ToLower(strings.TrimSpace(v.EntityType)) == "sosvc" {
			// these are listed as sosvc, but are not AOSHs
			if strings.ToLower(strings.TrimSpace(v.Id)) == "fsso" || strings.ToLower(strings.TrimSpace(v.Id)) == "fh" { continue }
			v.Distance = LatLongDist(op.Lat, op.Long, v.Lat, v.Long)
			AOs = append(AOs, v)
		}
	}
	thisOne.Distance = 999999
	for _,v := range AOs {
		if v.Distance < thisOne.Distance {
			thisOne = v
		}
	}
	return
}

func (op *ReachLoc) findClosest(entSet []Entity, validTypes []string) (thisOne Entity) {
	
	// refactored to be more efficient and more accurate
	thisOne.Distance = 999999
	for _,thisEnt := range entSet {
		if thisEnt.Lat == 0.0 && thisEnt.Long == 0.0 { continue }
		thisEnt.Distance = LatLongDist(op.Lat, op.Long, thisEnt.Lat, thisEnt.Long)
		if thisEnt.Distance < thisOne.Distance {
			thisOne = thisEnt
		}
	}
	return
}

func (op *ReachLoc) findClosestIO(entSet []Entity) (thisOne Entity) {

	thisOne.Distance = 999999
	for _,thisEnt := range entSet {
		if thisEnt.Lat == 0.0 && thisEnt.Long == 0.0 { continue }
		if strings.TrimSpace(thisEnt.CsiOrgStatus) == "ideal" && (strings.TrimSpace(thisEnt.EntityType) == "clvorg" || strings.TrimSpace(thisEnt.EntityType) == "testctr" || strings.TrimSpace(thisEnt.EntityType) == "communityctr") {
			thisEnt.Distance = LatLongDist(op.Lat, op.Long, thisEnt.Lat, thisEnt.Long)
			if thisEnt.Distance < thisOne.Distance {
				thisOne = thisEnt
			}
		}
	}
	return
}

func FindClosestIdealNotInCountry(loc *ReachLoc, entList *OrgsData, myWorld *World, avoid string) (thisOne Entity) {
	avoidThese := myWorld.GetStringsFor(avoid)
	finds := []Entity{}
	for x := 1 ; x<85 ; x++ {
		for _,thisEnt := range entList.Entities {

			if thisEnt.Lat == 0.0 && thisEnt.Long == 0.0 { continue }
			if stringInSlice(thisEnt.Country, avoidThese) { 
				continue 
			}
			
			if !(thisEnt.Lat < loc.Lat - float64(x)) && !(thisEnt.Long < loc.Long - (float64(x)*2)) && !(thisEnt.Lat > loc.Lat + float64(x)) && !(thisEnt.Long > loc.Long + (float64(x)*2)) {
				if strings.TrimSpace(thisEnt.CsiOrgStatus) == "ideal" && (strings.TrimSpace(thisEnt.EntityType) == "clvorg" || strings.TrimSpace(thisEnt.EntityType) == "testctr") {
					thisEnt.Distance = LatLongDist(loc.Lat, loc.Long, thisEnt.Lat, thisEnt.Long)
					finds = append(finds, thisEnt)
				}
			}
		}
		if len(finds) > 0 { break }
	}
	thisOne.Distance = 999999
	for _,v := range finds {
		if v.Distance < thisOne.Distance {
			thisOne = v
		}
	}
	return
}


func (myW *World) GetStringsFor(thisCountry string) (theStrings []string) {
	for _,v := range myW.Countries {
		if thisCountry == v.DuoCode || thisCountry == v.TriCode || thisCountry == v.FullName {
			theStrings = []string{v.DuoCode, v.TriCode, v.FullName}
			break
		}
	}
	return theStrings
}

func GetPrimaryFromEnts(entList []Entity, logic string) Entity {
	i := Entity{IncommId: -1, Distance: 999999.9} // placeholder for ideal org
	o := Entity{IncommId: -1, Distance: 999999.9}   // placeholder for org
	m := Entity{IncommId: -1, Distance: 999999.9}   // placeholder for msn
	p := Entity{IncommId: -1, Distance: 999999.9}   // placeholder for primary
	c := Entity{IncommId: -1, Distance: 999999.9}   // placeholder for closest thing

	for _, v := range entList {
		if strings.TrimSpace(v.CsiOrgStatus) == "ideal" && (strings.TrimSpace(v.EntityType) == "clvorg" || strings.TrimSpace(v.EntityType) == "testctr") {
			if i.Distance > v.Distance {
				i = v
			}
		} 
		if strings.TrimSpace(v.EntityType) == "clvorg" || strings.TrimSpace(v.EntityType) == "testctr" {
			if o.Distance > v.Distance {
				o = v
			} 
		}
		if strings.TrimSpace(v.EntityType) == "msn" {
			if m.Distance > v.Distance {
				m = v
			}
		}
		if c.Distance > v.Distance {
			c = v
		}
	}
	
	switch logic {
		case "ideal": // if there is an ideal org in the set, use it. (if mult use closest one)
			if i.IncommId > 0 { p = i } else { p = c }
		case "org": // if there is an org in the set, use it. (if mult use closest one)
			if o.IncommId > 0 { p = o } else { p = c }
			p = o
		case "closest": // really look for ideal org or clv or estab msn nearby - up to 200 mines - if not, whatever is closest
			if i.Distance < 30 {
				p = i
			} else if c.Distance < 30 && (c.EntityType == "clvorg" || (c.EntityType == "msn" && c.CsiOrgStatus == "established")) {
				p = c
			} else if i.Distance < 60 {
				p = i
			} else if c.Distance < 60 && (c.EntityType == "clvorg" || (c.EntityType == "msn" && c.CsiOrgStatus == "established")) {
				p = c
			} else if i.Distance < 90 {
				p = i
			} else if c.Distance < 90 && (c.EntityType == "clvorg" || (c.EntityType == "msn" && c.CsiOrgStatus == "established")) {
				p = c
			} else if i.Distance < 120 {
				p = i
			} else if c.Distance < 120 && (c.EntityType == "clvorg" || (c.EntityType == "msn" && c.CsiOrgStatus == "established")) {
				p = c
			} else if i.Distance < 200 {
				p = i
			} else if c.Distance < 200 && (c.EntityType == "clvorg" || (c.EntityType == "msn" && c.CsiOrgStatus == "established")) {
				p = c
			} else {
				p = c	
			}
		default: // look for ideal or cl v or estab msn up to 60 miles out. if not, use closest ideal regardless of distance. 
				 // fail to org, fail to msn, fail to closest
			if i.Distance < 30 {
				p = i
			} else if c.Distance < 30 && (c.EntityType == "clvorg" || (c.EntityType == "msn" && c.CsiOrgStatus == "established")) {
				p = c
			} else if i.Distance < 60 {
				p = i
			} else if c.Distance < 60 && (c.EntityType == "clvorg" || (c.EntityType == "msn" && c.CsiOrgStatus == "established")) {
				p = c
			} else if i.IncommId > 0 {
				p = i
			} else if o.IncommId > 0 {
				p = o
			} else if m.IncommId > 0 {
				p = m
			}

			if p.IncommId < 0 {
				p = c
			}
	}

	return p
}

func GetCountryList(thisCountry string, entList []Entity, myW *World) []Entity {
	inCountryList := []Entity{}
	countryStrings := myW.GetStringsFor(thisCountry)
	for _,v := range entList {
		for _,x := range countryStrings {
			if v.Country == x {
				inCountryList = append(inCountryList, v)
			}
		}
	}
	return inCountryList
}
		
func PopulateDistance(loc *ReachLoc, entList []Entity) []Entity {
	newlist := []Entity{}
	for _,v := range entList {
		v.Distance = LatLongDist(loc.Lat, loc.Long, v.Lat, v.Long)
		newlist = append(newlist, v)
	}
	return newlist
}		

func stringInSlice(myString string, mySlice []string) bool {
	for _,v := range mySlice {
		if myString == v {
			return true
		}
	}
	return false
}