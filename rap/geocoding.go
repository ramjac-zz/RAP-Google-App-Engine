package rap

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/urlfetch"
	"errors"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
	"log"
	"net/http"
)

func geocoding(w http.ResponseWriter, r *http.Request) *appError {
	if r.Method != "POST" {
		return &appError{
			errors.New("Unsupported HTTP method " + r.Method),
			"Unsupported HTTP method",
			http.StatusMethodNotAllowed,
		}
	}

	if !appengine.IsDevAppServer() && r.URL.Scheme != "https" {
		http.Redirect(w, r, "https://"+r.Host, http.StatusMovedPermanently)
	}

	c := appengine.NewContext(r)
	res := make([]*resource, 0)
	q := datastore.NewQuery("Resource").
		Filter("IsActive =", true).
		Filter("Location =", appengine.GeoPoint{})

	keys, err := q.GetAll(c, &res)

	if err != nil {
		return &appError{err, "Error querying database", http.StatusInternalServerError}
	}

	log.Printf("Records returned: %v", len(keys))

	//get updates from the geocoding api

	client := urlfetch.Client(c)
	m, err := maps.NewClient(maps.WithAPIKey("geoKey"), maps.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}

	gr := &maps.GeocodingRequest{
		Address: "9304 Woodcrest Rd Richmond VA 23229",
		Region:  "US",
	}

	resp, err := m.Geocode(context.Background(), gr)

	//resp, err := client.Get(geoCodingUrl + "9304 Woodcrest Rd Richmond VA 23229" + geoRegion + geoKey)

	if err != nil {
		return &appError{err, "Geocode request failed", http.StatusInternalServerError}
	}

	//unmarshall that to something a little easier to work with

	log.Println(resp[0].Geometry)

	/*
				//store the updates

		var res []resource
			var keys []*datastore.Key
				_, err = datastore.PutMulti(c, keys, res)
				if err != nil {
					log.Println(err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return &appError{err, "Error updating database", http.StatusInternalServerError}
				}
	*/

	// clear the cache
	memcache.Flush(c)

	http.Redirect(w, r, "/index.html", http.StatusFound)
	return nil
}

//These structs came from
//https://github.com/googlemaps/google-maps-services-go/blob/master/geocoding.go
//I didn't want to import the whole big "googlemaps.github.io/maps" and it needs a context.Context which doesn't quite fit with appengine.Context

const (
	geoCodingUrl = "https://maps.googleapis.com/maps/api/geocode/json?address="
	geoRegion    = "&region=us&key="
	geoKey       = ""
)

/*
// GeocodingResult is a single geocoded address
type GeocodingResult struct {
	AddressComponents []AddressComponent `json:"address_components"`
	FormattedAddress  string             `json:"formatted_address"`
	Geometry          AddressGeometry    `json:"geometry"`
	Types             []string           `json:"types"`
	PlaceID           string             `json:"place_id"`
}

// AddressComponent is a part of an address
type AddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

// AddressGeometry is the location of a an address
type AddressGeometry struct {
	Location     LatLng       `json:"location"`
	LocationType string       `json:"location_type"`
	Viewport     LatLngBounds `json:"viewport"`
	Types        []string     `json:"types"`
}
*/
