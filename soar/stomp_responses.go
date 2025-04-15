package soar

type FunctionCall struct {
	Function         Function         `json:"function"`
	Groups           []any            `json:"groups"`
	Inputs           any              `json:"inputs"`
	PlaybookInstance PlaybookInstance `json:"playbook_instance"`
	Principal        Principal        `json:"principal"`
	Workflow         Workflow         `json:"workflow"`
	WorkflowInstance WorkflowInstance `json:"workflow_instance"`
}
type TagHandle struct {
	DisplayName string `json:"display_name"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
}
type Tags struct {
	TagHandle TagHandle `json:"tag_handle"`
	Value     string    `json:"value"`
}
type Function struct {
	Creator           any    `json:"creator"`
	Description       any    `json:"description"`
	DisplayName       string `json:"display_name"`
	ID                int    `json:"id"`
	Name              string `json:"name"`
	OutputDescription any    `json:"output_description"`
	Tags              []Tags `json:"tags"`
	UUID              any    `json:"uuid"`
	Version           any    `json:"version"`
	ViewItems         []any  `json:"view_items"`
	Workflows         []any  `json:"workflows"`
}
type RestAPIBody struct {
	Format  string `json:"format"`
	Content any    `json:"content"`
}
type RestAPICookies struct {
	Format  string `json:"format"`
	Content any    `json:"content"`
}
type RestAPIHeaders struct {
	Format  string `json:"format"`
	Content any    `json:"content"`
}
type RestAPIMethod struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//	type Inputs struct {
//		JwtHeaders                string         `json:"jwt_headers"`
//		ClientAuthCert            string         `json:"client_auth_cert"`
//		JwtToken                  string         `json:"jwt_token"`
//		RestAPIAllowedStatusCodes string         `json:"rest_api_allowed_status_codes"`
//		RestAPIBody               RestAPIBody    `json:"rest_api_body"`
//		JwtAlgorithm              string         `json:"jwt_algorithm"`
//		OauthRedirectURI          string         `json:"oauth_redirect_uri"`
//		JwtKey                    string         `json:"jwt_key"`
//		JwtPayload                string         `json:"jwt_payload"`
//		RestAPIVerify             bool           `json:"rest_api_verify"`
//		OauthScope                string         `json:"oauth_scope"`
//		OauthAccessToken          string         `json:"oauth_access_token"`
//		OauthRefreshToken         string         `json:"oauth_refresh_token"`
//		OauthTokenType            string         `json:"oauth_token_type"`
//		RestAPITimeout            any            `json:"rest_api_timeout"`
//		OauthClientID             string         `json:"oauth_client_id"`
//		RestAPICookies            RestAPICookies `json:"rest_api_cookies"`
//		ClientAuthPem             string         `json:"client_auth_pem"`
//		RestAPIURL                string         `json:"rest_api_url"`
//		OauthTokenURL             string         `json:"oauth_token_url"`
//		OauthClientSecret         string         `json:"oauth_client_secret"`
//		ClientAuthKey             string         `json:"client_auth_key"`
//		RestAPIHeaders            RestAPIHeaders `json:"rest_api_headers"`
//		RestAPIMethod             RestAPIMethod  `json:"rest_api_method"`
//		OauthCode                 string         `json:"oauth_code"`
//	}
type PlaybookInstance struct {
	IsPlaybookDeleted      bool   `json:"is_playbook_deleted"`
	PlaybookActivationType string `json:"playbook_activation_type"`
	PlaybookDisplayName    string `json:"playbook_display_name"`
	PlaybookID             int    `json:"playbook_id"`
	PlaybookInstanceID     int    `json:"playbook_instance_id"`
}

type Principal struct {
	DisplayName string `json:"display_name"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type ObjectType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Workflow struct {
	Actions          []any      `json:"actions"`
	Description      any        `json:"description"`
	Name             string     `json:"name"`
	ObjectType       ObjectType `json:"object_type"`
	ProgrammaticName string     `json:"programmatic_name"`
	Tags             []any      `json:"tags"`
	UUID             any        `json:"uuid"`
	WorkflowID       int        `json:"workflow_id"`
}

type WorkflowInstance struct {
	Workflow           Workflow `json:"workflow"`
	WorkflowInstanceID int      `json:"workflow_instance_id"`
}

//Response upon function completion
type FuncResponse struct {
	MessageType int      `json:"message_type"`
	Message     string   `json:"message"`
	Complete    bool     `json:"complete"`
	Results     *Results `json:"results,omitempty"`
}

type WorkflowStatus struct {
	InstanceID   int    `json:"instance_id"`
	Status       string `json:"status"`
	StartDate    int64  `json:"start_date"`
	EndDate      any    `json:"end_date"`
	Reason       any    `json:"reason"`
	IsTerminated bool   `json:"is_terminated"`
}

type Content struct {
	WorkflowStatus WorkflowStatus `json:"Workflow Status"`
}

type Inputs struct {
	TimerTime  string `json:"timer_time"`
	TimerEpoch any    `json:"timer_epoch"`
}

type Metrics struct {
	Version         string `json:"version"`
	Package         string `json:"package"`
	PackageVersion  string `json:"package_version"`
	Host            string `json:"host"`
	ExecutionTimeMs int    `json:"execution_time_ms"`
	Timestamp       string `json:"timestamp"`
}

type Results struct {
	Version float64 `json:"version"`
	Success bool    `json:"success"`
	Reason  any     `json:"reason"`
	Content any     `json:"content"`
	Raw     any     `json:"raw"`
	Inputs  Inputs  `json:"inputs"`
	Metrics Metrics `json:"metrics"`
}
