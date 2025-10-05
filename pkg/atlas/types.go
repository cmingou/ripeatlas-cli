package atlas

// Probe represents a RIPE Atlas probe
type Probe struct {
	ID          int    `json:"id"`
	AddressV4   string `json:"address_v4"`
	AddressV6   string `json:"address_v6"`
	ASNV4       int    `json:"asn_v4"`
	ASNV6       int    `json:"asn_v6"`
	CountryCode string `json:"country_code"`
	Description string `json:"description"`
	Status      Status `json:"status"`
	IsPublic    bool   `json:"is_public"`
}

// Status represents probe status
type Status struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Since string `json:"since"`
}

// ProbeResponse represents the API response for probe queries
type ProbeResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []Probe `json:"results"`
}

// MeasurementDefinition defines a traceroute measurement
type MeasurementDefinition struct {
	Type            string `json:"type"`
	AF              int    `json:"af"`
	Target          string `json:"target"`
	Description     string `json:"description"`
	Protocol        string `json:"protocol"`
	Packets         int    `json:"packets"`
	Size            int    `json:"size"`
	MaxHops         int    `json:"max_hops"`
	Paris           int    `json:"paris"`
	ResponseTimeout int    `json:"response_timeout"`
}

// ProbeSet defines which probes to use
type ProbeSet struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Requested int    `json:"requested"`
}

// MeasurementRequest represents a measurement creation request
type MeasurementRequest struct {
	Definitions []MeasurementDefinition `json:"definitions"`
	Probes      []ProbeSet              `json:"probes"`
	IsOneoff    bool                    `json:"is_oneoff"`
}

// MeasurementResponse represents the response after creating a measurement
type MeasurementResponse struct {
	Measurements []int `json:"measurements"`
}

// MeasurementStatus represents the status of a measurement
type MeasurementStatus struct {
	ID               int        `json:"id"`
	Status           StatusInfo `json:"status"`
	ProbesScheduled  int        `json:"probes_scheduled"`
	ProbesRequested  int        `json:"probes_requested"`
	ParticipantCount int        `json:"participant_count"`
	StartTime        int64      `json:"start_time"`
	StopTime         int64      `json:"stop_time"`
}

// StatusInfo represents detailed status information
type StatusInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TracerouteResult represents a single traceroute measurement result
type TracerouteResult struct {
	ProbeID   int         `json:"prb_id"`
	MsmID     int         `json:"msm_id"`
	Timestamp int64       `json:"timestamp"`
	From      string      `json:"from"`
	Type      string      `json:"type"`
	Result    []HopResult `json:"result"`
	DstAddr   string      `json:"dst_addr"`
	SrcAddr   string      `json:"src_addr"`
}

// HopResult represents a single hop in a traceroute
type HopResult struct {
	Hop    int        `json:"hop"`
	Result []HopReply `json:"result"`
}

// HopReply represents a reply from a hop
type HopReply struct {
	From string  `json:"from,omitempty"`
	RTT  float64 `json:"rtt,omitempty"`
	Size int     `json:"size,omitempty"`
	TTL  int     `json:"ttl,omitempty"`
	Err  string  `json:"err,omitempty"`
	X    string  `json:"x,omitempty"` // Timeout indicator
}

// ASNInfo represents ASN information extracted from traceroute
type ASNInfo struct {
	ASN         int
	Name        string
	Occurrences int
	Percentage  float64
	AvgHopStart int
	AvgHopEnd   int
}

// ProbeAllocation tracks probe distribution per ASN
type ProbeAllocation struct {
	ASN       int
	Available int
	Allocated int
	ProbeIDs  []int
}
