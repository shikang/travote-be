package main

// Country - Caps for field names, because of json.Marshal requirements
type Country struct {
	Abbr  string `json:"abbr"`
	Name  string `json:"name"`
	Xaxis int64  `json:"xaxis,string"`
	Yaxis int64  `json:"yaxis,string"`
}

// Place - Caps for field names, because of json.Marshal requirements
type Place struct {
	ID       string  `json:"id"`
	Abbr     string  `json:"abbr"`
	Name     string  `json:"name"`
	Master   string  `json:"master"`
	Category string  `json:"category"`
	Desc     string  `json:"desc"`
	Lat      float64 `json:"lat,string"`
	Long     float64 `json:"long,string"`
	Address  string  `json:"address"`
	Postal   string  `json:"postal"`
	Contact  string  `json:"contact"`
	Hours    string  `json:"hours"`
	Website  string  `json:"website"`
	Email    string  `json:"email"`
	Zone     string  `json:"zone"`
	Ext1     string  `json:"ext_1"`
}
