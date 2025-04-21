package models

type Flight struct {
	FlightNumber     string  `json:"flightNumber"`
	Airline          string  `json:"airline"`
	DepartureAirport Airport `json:"departureAirport"`
	ArrivalAirport   Airport `json:"arrivalAirport"`
	DepartureTime    string  `json:"departureTime"` // You can use time.Time if you want to parse it
	ArrivalTime      string  `json:"arrivalTime"`   // Same here
	Class            string  `json:"class"`
	Status           string  `json:"status"`
	Duration         string  `json:"duration"`
	PriceUSD         float64 `json:"priceUSD"`
}
