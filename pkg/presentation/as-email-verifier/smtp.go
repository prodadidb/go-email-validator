package asemailverifier

// SMTPPresentation for smtp validator
type SMTPPresentation struct {
	HostExists  bool `json:"host_exists"`
	FullInbox   bool `json:"full_inbox"`
	CatchAll    bool `json:"catch_all"`
	Deliverable bool `json:"deliverable"`
	Disabled    bool `json:"disabled"`
}
