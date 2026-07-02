package provider

type remoteAPI struct {
	Name              string   `json:"name"`
	PackageType       string   `json:"package_type"`
	RepoType          string   `json:"repo_type,omitempty"`
	BaseURL           string   `json:"base_url"`
	Description       string   `json:"description,omitempty"`
	Username          string   `json:"username,omitempty"`
	Password          string   `json:"password,omitempty"`
	ImmutableTTL      int64    `json:"immutable_ttl"`
	MutableTTL        int64    `json:"mutable_ttl"`
	CheckMutable      bool     `json:"check_mutable"`
	Patterns          []string `json:"patterns"`
	Blocklist         []string `json:"blocklist"`
	MutablePatterns   []string `json:"mutable_patterns"`
	ImmutablePatterns []string `json:"immutable_patterns"`
	BanTagsEnabled    bool     `json:"ban_tags_enabled"`
	BanTags           []string `json:"ban_tags"`
	QuarantineEnabled bool     `json:"quarantine_enabled"`
	QuarantineDays    int64    `json:"quarantine_days"`
	StaleOnError      bool     `json:"stale_on_error"`
	ReleasesRemote    string   `json:"releases_remote,omitempty"`
	ManagedBy         string   `json:"managed_by,omitempty"`

	UpstreamDialTimeout           int64 `json:"upstream_dial_timeout"`
	UpstreamTLSTimeout            int64 `json:"upstream_tls_timeout"`
	UpstreamResponseHeaderTimeout int64 `json:"upstream_response_header_timeout"`
}

type virtualAPI struct {
	Name        string   `json:"name"`
	PackageType string   `json:"package_type"`
	Description string   `json:"description,omitempty"`
	Members     []string `json:"members"`
	ManagedBy   string   `json:"managed_by,omitempty"`
}
