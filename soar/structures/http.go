package structures

type SessionResponseJson struct {
	Orgs                   []Org `json:"orgs"`
	PasswordExpirationDate int64  `json:"password_expiration_date"`
	APIKeyHandle           int    `json:"api_key_handle"`
	ClientID               string `json:"client_id"`
	DisplayName            string `json:"display_name"`
}
type LastModifiedBy struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}
type Org struct {
	ID                          int            `json:"id"`
	Name                        string         `json:"name"`
	Addr                        any            `json:"addr"`
	Addr2                       any            `json:"addr2"`
	City                        any            `json:"city"`
	State                       any            `json:"state"`
	Zip                         any            `json:"zip"`
	AttachmentsEnabled          bool           `json:"attachments_enabled"`
	FinalPhaseRequired          bool           `json:"final_phase_required"`
	TasksPrivate                bool           `json:"tasks_private"`
	HasSaml                     bool           `json:"has_saml"`
	RequireSaml                 bool           `json:"require_saml"`
	TwofactorAuthDomain         any            `json:"twofactor_auth_domain"`
	HasAvailableTwofactor       bool           `json:"has_available_twofactor"`
	AuthorizedLdapGroup         any            `json:"authorized_ldap_group"`
	SupportsLdap                bool           `json:"supports_ldap"`
	IncidentDeletionAllowed     bool           `json:"incident_deletion_allowed"`
	ConfigurationType           string         `json:"configuration_type"`
	ParentOrg                   any            `json:"parent_org"`
	SessionTimeout              int            `json:"session_timeout"`
	LastModifiedBy              LastModifiedBy `json:"last_modified_by"`
	LastModifiedTime            int64          `json:"last_modified_time"`
	UUID                        string         `json:"uuid"`
	Timezone                    any            `json:"timezone"`
	CloudAccount                any            `json:"cloud_account"`
	Perms                       any            `json:"perms"`
	EffectivePermissions        []any          `json:"effective_permissions"`
	RoleHandles                 []any          `json:"role_handles"`
	Enabled                     bool           `json:"enabled"`
	TwofactorCookieLifetimeSecs int            `json:"twofactor_cookie_lifetime_secs"`
}


type MessageDestination struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	ProgrammaticName string `json:"programmatic_name"`
	DestinationType  int    `json:"destination_type"`
	ExpectAck        bool   `json:"expect_ack"`
	//Users            []any  `json:"users"`
	UUID             string `json:"uuid"`
	ExportKey        string `json:"export_key"`
	//Tags skipped
	APIKeys          []int  `json:"api_keys"`
}

type InboundDestination struct {
	ID              int    `json:"id"`
	DisplayName     string `json:"display_name"`
	Name            string `json:"name"`
	WritePrincipals []int  `json:"write_principals"`
	ReadPrincipals  []int  `json:"read_principals"`
	UUID            string `json:"uuid"`
	Tags            []any  `json:"tags"`
	Version         int    `json:"version"`
	ExportKey       string `json:"export_key"`
}
