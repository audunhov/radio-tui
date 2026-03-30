package models

type Station struct {
	ID        string `json:"stationuuid"`
	Name      string `json:"name"`
	URL       string `json:"url_resolved"`
	Homepage  string `json:"homepage"`
	Country   string `json:"country"`
	Codec     string `json:"codec"`
	Bitrate   int    `json:"bitrate"`
	Tags       string `json:"tags"`
	IsFav      bool   `json:"-"`
}


func (s Station) Title() string       { return s.Name }
func (s Station) Description() string { return s.Country + " | " + s.Codec }
func (s Station) FilterValue() string { return s.Name }
