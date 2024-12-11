// Package geocoder provides an easy way to use the Google Geocoding API
package geocoder

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"reachesRouting/geocoder/structs"
	u2 "net/url"
	"crypto/tls"

	"fmt"
)

// projectrchrt 

// The user should set the API KEY provided by Google
// var ApiKey string

// Define the Geocode API URL as a constant
const (
	geocodeApiUrl = "https://maps.googleapis.com/maps/api/geocode/json?"
	ApiKey = "<API_KEY>" 
)

// Address structure used in the Geocoding and GeocodingReverse functions
// Note: The FormattedAddress field should be used only for the GeocodingReverse
// to get the formatted address from the Google Geocoding API. It is not used in
// the Geocoding function.
type Address struct {
	Street           string
	Number           int
	Neighborhood     string
	District         string
	City             string
	County           string
	State            string
	Country          string
	PostalCode       string
	FormattedAddress string
	Types            string
}

// Location structure used in the Geocoding and GeocodingReverse functions
type Location struct {
	Latitude  float64
	Longitude float64
}

// Format an address based on the Address structure
// Return the formated address (string)
func (address *Address) FormatAddress() string {

	// Creats a slice with all content from the Address struct
	var content []string
	if address.Number > 0 {
		content = append(content, strconv.Itoa(address.Number))
	}
	content = append(content, address.Street)
	content = append(content, address.Neighborhood)
	content = append(content, address.District)
	content = append(content, address.PostalCode)
	content = append(content, address.City)
	content = append(content, address.County)
	content = append(content, address.State)
	content = append(content, address.Country)

	var formattedAddress string

	// For each value in the content slice check if it is valid
	// and add to the formattedAddress string
	for _, value := range content {
		if value != "" {
			if formattedAddress != "" {
				formattedAddress += ", "
			}
			formattedAddress += value
		}
	}

	return formattedAddress
}

// httpRequest function send the HTTP request, decode the JSON
// and return a Results structure
func httpRequest(url string) (structs.Results, error) {

	fmt.Println(url)

	var results structs.Results

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return results, err
	}

	// For control over HTTP client headers, redirect policy, and other settings, create a Client
	// A Client is an HTTP client
	proxyUrl,_ := u2.Parse("http://10.16.32.150:8081")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy: http.ProxyURL(proxyUrl),
	}
	client := &http.Client{Transport: tr}

	// Send the request via a client
	// Do sends an HTTP request and returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		return results, err
	}

	// Callers should close resp.Body when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Use json.Decode for reading streams of JSON data
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return results, err
	}

	// The "OK" status indicates that no error has occurred, it means
	// the address was analyzed and at least one geographic code was returned
	if strings.ToUpper(results.Status) != "OK" {
		// If the status is not "OK" check what status was returned
		switch strings.ToUpper(results.Status) {
		case "ZERO_RESULTS":
			err = errors.New("No results found.")
			break
		case "OVER_QUERY_LIMIT":
			err = errors.New("You are over your quota.")
			break
		case "REQUEST_DENIED":
			err = errors.New("Your request was denied.")
			break
		case "INVALID_REQUEST":
			err = errors.New("Probably the query is missing.")
			break
		case "UNKNOWN_ERROR":
			err = errors.New("Server error. Please, try again.")
			break
		default:
			break
		}
	}

	return results, err
}

// Geocoding function is used to convert an Address structure
// to a Location structure (latitude and longitude)
func Geocoding(address Address) (Location, error) {

	var location Location

	// Convert whitespaces to +
	formattedAddress := address.FormatAddress()
	formattedAddress = strings.Replace(formattedAddress, " ", "+", -1)

	// Create the URL based on the formated address
	url := geocodeApiUrl + "address=" + formattedAddress

	// Use the API Key if it was set
	if ApiKey != "" {
		url += "&key=" + ApiKey
	}

	// Send the HTTP request and get the results
	results, err := httpRequest(url)
	if err != nil {
		log.Println(err)
		return location, err
	}

	// Get the results (latitude and longitude)
	location.Latitude = results.Results[0].Geometry.Location.Lat
	location.Longitude = results.Results[0].Geometry.Location.Lng

	return location, nil
}

// Convert a structs.Results to a slice of Address structures
func convertResultsToAddress(results structs.Results) (addresses []Address) {

	for index := 0; index < len(results.Results); index++ {
		var address Address

		// Put each component from the AddressComponents slice in the correct field in the Address structure
		for _, component := range results.Results[index].AddressComponents {
			// Check all types of each component
			for _, types := range component.Types {
				switch types {
				case "route":
					address.Street = component.LongName
					break
				case "street_number":
					address.Number, _ = strconv.Atoi(component.LongName)
					break
				case "neighborhood":
					address.Neighborhood = component.LongName
					break
				case "sublocality":
					address.District = component.LongName
					break
				case "sublocality_level_1":
					address.District = component.LongName
					break
				case "locality":
					address.City = component.LongName
					break
				case "administrative_area_level_3":
					address.City = component.LongName
					break
				case "administrative_area_level_2":
					address.County = component.LongName
					break
				case "administrative_area_level_1":
					address.State = component.LongName
					break
				case "country":
					address.Country = component.LongName
					break
				case "postal_code":
					address.PostalCode = component.LongName
					break
				default:
					break
				}
			}
		}

		address.FormattedAddress = results.Results[index].FormattedAddress
		address.Types = results.Results[index].Types[0]

		addresses = append(addresses, address)
	}
	return
}

// GeocodingReverse function is used to convert a Location structure
// to an Address structure
func GeocodingReverse(location Location) ([]Address, error) {

	var addresses []Address

	// Convert the latitude and longitude from double to string
	latitude := strconv.FormatFloat(location.Latitude, 'f', 8, 64)
	longitude := strconv.FormatFloat(location.Longitude, 'f', 8, 64)

	// Create the URL based on latitude and longitude
	url := geocodeApiUrl + "latlng=" + latitude + "," + longitude

	// Use the API key if it was set
	if ApiKey != "" {
		url += "&key=" + ApiKey
	}

	// Send the HTTP request and get the results
	results, err := httpRequest(url)
	if err != nil {
		log.Println(err)
		return addresses, err
	}

	// Convert the results to an Address slice called addresses
	addresses = convertResultsToAddress(results)

	fmt.Println("Street: " + addresses[0].Street)
	fmt.Println("Number: ", addresses[0].Number)
	fmt.Println("Neighborhood: " + addresses[0].Neighborhood)
	fmt.Println("District: " + addresses[0].District)
	fmt.Println("City: " + addresses[0].City)
	fmt.Println("County: " + addresses[0].County)
	fmt.Println("State: " + addresses[0].State)
	fmt.Println("Country: " + addresses[0].Country)
	fmt.Println("PostalCode: " + addresses[0].PostalCode)
	fmt.Println("FormattedAddress: " + addresses[0].FormattedAddress)
	fmt.Println("Types: " + addresses[0].Types)

	// EXAMPLE RESULTS: 
	// Lat: 51.10892100, Long: -0.02894900
	// Street: Saint Hill Road
	// Number: 2
	// Neighborhood: 
	// District: 
	// City: 
	// County: West Sussex
	// State: England
	// Country: United Kingdom
	// PostalCode: RH19 4JX
	// FormattedAddress: 2 Saint Hill Rd, East Grinstead RH19 4JX, UK

	// Lat: 42.40480000, Long: -82.19100000
	// Street: King Street West
	// Number: 315
	// Neighborhood: Chatham
	// District:
	// City: Chatham-Kent
	// County: Chatham-Kent Division
	// State: Ontario
	// Country: Canada
	// PostalCode: N7M 5K8
	// FormattedAddress: 315 King St W, Chatham, ON N7M 5K8, Canada

	// Lat: 34.10197800,-18.32797600
	// Street: Hollywood Blvd
	// Number: 6331
	// Neighborhood: Hollywood
	// District: 
	// City: Los Angeles
	// County: Los Angeles County
	// State: California
	// Country: United States
	// PostalCode: 90028
	// FormattedAddress: 6331 Hollywood Blvd, Los Angeles, CA 90028, USA

	// Lat: 43.6047, Long: 1.4442 - south of France
	// Street: Rue Lafayette
	// Number: 5
	// Neightborhood: 
	// District: 
	// City: Toulouse
	// County: Haute-Garonne
	// State: Occitanie
	// Country: France
	// PostalCode: 31000
	// FormattedAddress: 5 Rue Lafayette, 31000 Toulouse, France

	// Lat: 50.84925, 142.6652
	// Street: Sovetskaya Ulitsa
	// Number: 0
	// Neightborhood: 
	// District: 
	// City: Tymovskoye
	// County: Tymovskiy Rayon
	// State: Sakhalinskaya oblast'
	// Country: Russia
	// PostalCode: 694400
	// FormattedAddress: Sovetskaya Ulitsa, Tymovskoye, Sakhalinskaya oblast', Russia, 694400

	// Lat: 66.5039, 25.7294
	// Street: Poromiehentie
	// Number: 8
	// Neightborhood: 
	// District: 
	// City: Rovaniemi
	// County: 
	// State: 
	// Country: Finland
	// PostalCode: 96200
	// FormattedAddress: Poromiehentie 8, 96200 Rovaniemi, Finland

	return addresses, nil
}
