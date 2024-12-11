package reachesRouting

import "testing"

func TestGetRoutingData(t *testing.T) {
	testCases := map[string][]float64{
		"sead": []float64{82.3, -113.531}, // sample data.
	}

	for org,coords := range testCases {
		reach := ReachLoc{Lat: coords[0], Long: coords[1]}
		routingData, err := reach.GetRoutingData("scn")
		if err != nil {
			t.Error(err)
		}

		v := routingData.Primary.Id
		if v != org {
			t.Error("Expected " + org + ", got ", v)
		}
	}
}