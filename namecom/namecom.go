package namecom

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type NameCom struct {
	Server string
	User   string
	Token  string
	Client *http.Client
}

func New(user, token string) *NameCom {
	return &NameCom{
		Server: "api.name.com",
		User:   user,
		Token:  token,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func Test(user, token string) *NameCom {
	return &NameCom{
		Server: "api.dev.name.com",
		User:   user,
		Token:  token,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (er ErrorResponse) Error() string {
	return er.Message + ": " + er.Details
}

func (n *NameCom) ErrorResponse(resp *http.Response) error {
	er := &ErrorResponse{}
	err := json.NewDecoder(resp.Body).Decode(er)
	if err != nil {
		return errors.Wrap(err, "api returned unexpected response")
	}

	return errors.WithStack(er)
}

func (n *NameCom) Get(endpoint string, values url.Values) (io.Reader, error) {
	if len(values) == 0 {
		endpoint = endpoint + "?" + values.Encode()
	}
	return n.doRequest("GET", endpoint, nil)
}

func (n *NameCom) Post(endpoint string, post io.Reader) (io.Reader, error) {
	return n.doRequest("POST", endpoint, post)
}

func (n *NameCom) Put(endpoint string, post io.Reader) (io.Reader, error) {
	return n.doRequest("PUT", endpoint, post)
}

func (n *NameCom) Delete(endpoint string, post io.Reader) (io.Reader, error) {
	return n.doRequest("DELETE", endpoint, post)
}

func (n *NameCom) doRequest(method, endpoint string, post io.Reader) (io.Reader, error) {
	url := "https://" + n.Server + endpoint

	req, err := http.NewRequest(method, url, post)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(n.User, n.Token)

	resp, err := n.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, n.ErrorResponse(resp)
	}

	return resp.Body, nil
}

// EmptyResponse is an empty response used for DELETE endpoints.
type EmptyResponse struct {
}

// ErrorResponse is what is returned if the HTTP status code is not 200.
type ErrorResponse struct {
    // Message is the error message.
    Message string `json:"message,omitempty"`
    // Details may have some additional details about the error.
    Details string `json:"details,omitempty"`
}

// Record is an individual DNS resource record.
type Record struct {
    // Unique record id. Value is ignored on Create, and must match the URI on Update.
    Id int32 `json:"id,omitempty"`
    // DomainName is the zone that the record belongs to.
    DomainName string `json:"domainName,omitempty"`
    // Name is the hostname relative to the zone: e.g. for a record for blog.example.org, domain would be "example.org" and host would be "blog".
    // An apex record would be specified by either an empty host "" or "@".
    // A SRV record would be specified by "_{service}._{protocal}.{host}": e.g. "_sip._tcp.phone" for _sip._tcp.phone.example.org.
    Host string `json:"host,omitempty"`
    // FQDN is the Fully Qualified Domain Name. It is the combination of the host and the domain name. It always ends in a ".".
    Fqdn string `json:"fqdn,omitempty"`
    // Type is one of the following: A, AAAA, ANAME, CNAME, MX, NS, SRV, or TXT.
    Type string `json:"type,omitempty"`
    // Answer is either the IP address for A or AAAA records; the target for ANAME, CNAME, MX, or NS records; the text for TXT records.
    // For SRV records, answer has the following format: "{weight} {port} {target}" e.g. "1 5061 sip.example.org".
    Answer string `json:"answer,omitempty"`
    // TTL is the time this record can be cached for in seconds. Name.com allows a minimum TTL of 300, or 5 minutes.
    Ttl uint32 `json:"ttl,omitempty"`
    // Priority is only required for MX and SRV records, it is ignored for all others.
    Priority uint32 `json:"priority,omitempty"`
}

// ListRecordsRequest requests a list of records that exist for the domain
type ListRecordsRequest struct {
    // DomainName is the zone to list the records for.
    DomainName string `json:"domainName,omitempty"`
    // Per Page is the number of records to return per request. Per Page defaults to 1,000.
    PerPage int32 `json:"perPage,omitempty"`
    // Page is which page to return
    Page int32 `json:"page,omitempty"`
}

// ListRecordsResponse is the response from a list request
type ListRecordsResponse struct {
    // Records contains the records in the zone
    Records []*Record `json:"records,omitempty"`
    // NextPage is the identifier for the next page of results. It is only populated if there is another page of results after the current page.
    NextPage int32 `json:"nextPage,omitempty"`
    // LastPage is the identifier for the final page of results. It is only populated if there is another page of results after the current page.
    LastPage int32 `json:"lastPage,omitempty"`
}

// GetRecordRequest requests the record identified by id and domain.
type GetRecordRequest struct {
    // DomainName is the zone the record exists in
    DomainName string `json:"domainName,omitempty"`
    // ID is the server-assigned unique identifier for this record
    Id int32 `json:"id,omitempty"`
}

// DeleteRecordRequest deletes a specific record
type DeleteRecordRequest struct {
    // DomainName is the zone that the record to be deleted exists in.
    DomainName string `json:"domainName,omitempty"`
    // ID is the server-assigned unique identifier for the Record to be deleted. If the Record with that ID does not exist in the specified Domain, an error is returned.
    Id int32 `json:"id,omitempty"`
}

// 
type DNSSEC struct {
    // DomainName is the domain name.
    DomainName string `json:"domainName,omitempty"`
    // KeyTag contains the key tag value of the DNSKEY RR that validates this signature. The algorithm to generate it is here: https://tools.ietf.org/html/rfc4034#appendix-B
    KeyTag int32 `json:"keyTag,omitempty"`
    // Algorithm is an integer identifying the algorithm used for signing. Valid values can be found here: https://www.iana.org/assignments/dns-sec-alg-numbers/dns-sec-alg-numbers.xhtml
    Algorithm int32 `json:"algorithm,omitempty"`
    // DigestType is an integer identifying the algorithm used to create the digest. Valid values can be found here: https://www.iana.org/assignments/ds-rr-types/ds-rr-types.xhtml
    DigestType int32 `json:"digestType,omitempty"`
    // Digest is a digest of the DNSKEY RR that is registered with the registry.
    Digest string `json:"digest,omitempty"`
}

// 
type ListDNSSECsRequest struct {
    // DomainName is the domain name to list keys for.
    DomainName string `json:"domainName,omitempty"`
}

// 
type ListDNSSECsResponse struct {
    // Dnssec is the list of registered DNSSEC keys.
    Dnssec []*DNSSEC `json:"dnssec,omitempty"`
    // NextPage is the identifier for the next page of results. It is only populated if there is another page of results after the current page.
    NextPage int32 `json:"nextPage,omitempty"`
    // LastPage is the identifier for the final page of results. It is only populated if there is another page of results after the current page.
    LastPage int32 `json:"lastPage,omitempty"`
}

// 
type GetDNSSECRequest struct {
    // DomainName is the domain name.
    DomainName string `json:"domainName,omitempty"`
    // Digest is the digest for the DNSKEY RR to retrieve.
    Digest string `json:"digest,omitempty"`
}

// 
type DeleteDNSSECRequest struct {
    // DomainName is the domain name the key is registered for.
    DomainName string `json:"domainName,omitempty"`
    // Digest is the digest for the DNSKEY RR to remove from the registry.
    Digest string `json:"digest,omitempty"`
}

// 
type Contact struct {
    // First name of the contact.
    FirstName string `json:"firstName,omitempty"`
    // Last name of the contact.
    LastName string `json:"lastName,omitempty"`
    // Company name of the contact. Leave blank if the contact is an individual as some registries will assume it is a corporate entity otherwise.
    CompanyName string `json:"companyName,omitempty"`
    // Address1 is the first line of the contact's address.
    Address1 string `json:"address1,omitempty"`
    // Address2 is the second line of the contact's address.
    Address2 string `json:"address2,omitempty"`
    // City of the contact's address.
    City string `json:"city,omitempty"`
    // State or Province for the contact's address.
    State string `json:"state,omitempty"`
    // Zip or Postal Code for the contact's address.
    Zip string `json:"zip,omitempty"`
    // Country code for the contact's address. Required to be a ISO 3166-1 alpha-2 code.
    Country string `json:"country,omitempty"`
    // Phone number of the contact. Should be specified in the following format: "+cc.llllllll" where cc is the country code and llllllll is the local number.
    Phone string `json:"phone,omitempty"`
    // Fax number of the contact. Should be specified in the following format: "+cc.llllllll" where cc is the country code and llllllll is the local number.
    Fax string `json:"fax,omitempty"`
}

// 
type Contacts struct {
    // Registrant is the rightful owner of the account and has the right to use and/or sell the domain name. They are able to make changes to all account, domain, and product settings. This information should be reviewed and updated regularly to ensure accuracy.
    Registrant *Contact `json:"registrant,omitempty"`
    // Registrants often designate an administrative contact to manage their domain name(s). They primarily deal with business information such as the name on record, postal address, and contact information for the official registrant.
    Admin *Contact `json:"admin,omitempty"`
    // The technical contact manages and maintains a domain’s nameservers. If you’re working with a web designer or someone in a similar role, you many want to assign them as a technical contact.
    Tech *Contact `json:"tech,omitempty"`
    // The billing contact is the party responsible for paying bills for the account and taking care of renewals.
    Billing *Contact `json:"billing,omitempty"`
}

// 
type Domain struct {
    // DomainName is the punycode encoded value of the domain name.
    DomainName string `json:"domainName,omitempty"`
    // Nameservers is the list of nameservers for this domain. If unspecified it defaults to your account default nameservers.
    Nameservers []string `json:"nameservers,omitempty"`
    // Contacts for the domain.
    Contacts *Contacts `json:"contacts,omitempty"`
    // PrivacyEnabled reflects if Whois Privacy is enabled for this domain.
    PrivacyEnabled bool `json:"privacyEnabled,omitempty"`
    // Locked indicates that the domain cannot be transfered to another registrar.
    Locked bool `json:"locked,omitempty"`
    // AutorenewEnabled indicates if the domain will attempt to renew automatically before expiration.
    AutorenewEnabled bool `json:"autorenewEnabled,omitempty"`
    // ExpireDate is the date the domain will expire.
    ExpireDate string `json:"expireDate,omitempty"`
    // CreateDate is the date the domain was created at the registry.
    CreateDate string `json:"createDate,omitempty"`
    // RenewalPrice is the price to renew the domain. It may be required for the RenewDomain command.
    RenewalPrice float64 `json:"renewalPrice,omitempty"`
}

// 
type SearchRequest struct {
    // Timeout is a value in milliseconds on how long to perform the search for. Valid timeouts are between 500ms to 5,000ms. If not specified, timeout defaults to 1,000ms.
    // Since some additional processing is performed on the results, a response may take longer then the timeout.
    Timeout int32 `json:"timeout,omitempty"`
    // Keyword is the search term to search for. It can be just a word, or a whole domain name.
    Keyword string `json:"keyword,omitempty"`
    // TLDFilter will limit results to only contain the specified TLDs.
    TldFilter []string `json:"tldFilter,omitempty"`
    // PromoCode is not implemented yet.
    PromoCode string `json:"promoCode,omitempty"`
}

// 
type AvailabilityRequest struct {
    // DomainNames is the list of domains to check if they are available.
    DomainNames []string `json:"domainNames,omitempty"`
    // PromoCode is not implemented yet.
    PromoCode string `json:"promoCode,omitempty"`
}

// 
type SearchResult struct {
    // DomainName is the punycode encoding of the result domain name.
    DomainName string `json:"domainName,omitempty"`
    // SLD is first portion of the domain_name.
    Sld string `json:"sld,omitempty"`
    // TLD is the rest of the domain_name after the SLD.
    Tld string `json:"tld,omitempty"`
    // Purchaseable indicates whether the search result is available for purchase.
    Purchasable bool `json:"purchasable,omitempty"`
    // Premium indicates that this search result is a premium result and the purchase_price needs to be passed to the DomainCreate command.
    Premium bool `json:"premium,omitempty"`
    // PurchasePrice is the price for purchasing this domain for 1 year. Purchase_price is always in USD.
    PurchasePrice float64 `json:"purchasePrice,omitempty"`
    // PurchaseType indicates what kind of purchase this result is for. It should be passed to the DomainCreate command.
    PurchaseType string `json:"purchaseType,omitempty"`
    // RenewalPrice is the annual renewal price for this domain as it may be different then the purchase_price.
    RenewalPrice float64 `json:"renewalPrice,omitempty"`
}

// 
type SearchResponse struct {
    // Results of the search are returned here, the order should not be relied upon.
    Results []*SearchResult `json:"results,omitempty"`
}

// 
type ListDomainsRequest struct {
    // Per Page is the number of records to return per request. Per Page defaults to 1,000.
    PerPage int32 `json:"perPage,omitempty"`
    // Page is which page to return
    Page int32 `json:"page,omitempty"`
}

// ListDomainResponse is the response from a list request
type ListDomainsResponse struct {
    // Domains is the list of domains in your account.
    Domains []*Domain `json:"domains,omitempty"`
    // NextPage is the identifier for the next page of results. It is only populated if there is another page of results after the current page.
    NextPage int32 `json:"nextPage,omitempty"`
    // LastPage is the identifier for the final page of results. It is only populated if there is another page of results after the current page.
    LastPage int32 `json:"lastPage,omitempty"`
}

// 
type GetDomainRequest struct {
    // DomainName is the domain to retrieve.
    DomainName string `json:"domainName,omitempty"`
}

// 
type CreateDomainRequest struct {
    // Domain is the domain object to create. If privacy_enabled is set, Whois Privacy will also be purchased for an additional amount.
    Domain *Domain `json:"domain,omitempty"`
    // PurchasePrice is the amount to pay for the domain. If privacy_enabled is set, the regular price for whois protection will be added automatically. If VAT tax applies, it will also be added automatically.
    // PurchasePrice is required if purchase_type is not "registration" or if it is a premium domain.
    PurchasePrice float64 `json:"purchasePrice,omitempty"`
    // PurchaseType defaults to "registration" but should be copied from the result of a search command otherwise.
    PurchaseType string `json:"purchaseType,omitempty"`
    // Years is for how many years to register the domain for. Years defaults to 1 if not passed and cannot be more than 10.
    // If passing purchase_price make sure to adjust it accordingly.
    Years int32 `json:"years,omitempty"`
    // TLDRequirements is a way to pass additional data that is required by some registries.
    TldRequirements map[string]string `json:"tldRequirements,omitempty"`
    // PromoCode is not yet implemented.
    PromoCode string `json:"promoCode,omitempty"`
}

// 
type CreateDomainResponse struct {
    // Domain is the newly created domain.
    Domain *Domain `json:"domain,omitempty"`
    // Order is an identifier for this purchase.
    Order int32 `json:"order,omitempty"`
    // TotalPaid is the total amount paid, including VAT and whois protection.
    TotalPaid float64 `json:"totalPaid,omitempty"`
}

// 
type RenewDomainRequest struct {
    // DomainName is the domain to renew.
    DomainName string `json:"domainName,omitempty"`
    // PurchasePrice is the amount to pay for the domain renewal. If VAT tax applies, it will also be added automatically.
    // PurchasePrice is required if this is a premium domain.
    PurchasePrice float64 `json:"purchasePrice,omitempty"`
    // Years is for how many years to renew the domain for. Years defaults to 1 if not passed and cannot be more than 10.
    Years int32 `json:"years,omitempty"`
    // PromoCode is not yet implemented.
    PromoCode string `json:"promoCode,omitempty"`
}

// 
type RenewDomainResponse struct {
    // Domain reflects the status of the domain after renewing.
    Domain *Domain `json:"domain,omitempty"`
    // Order is an identifier for this purchase
    Order int32 `json:"order,omitempty"`
    // TotalPaid is the total amount paid, including VAT.
    TotalPaid float64 `json:"totalPaid,omitempty"`
}

// 
type AuthCodeRequest struct {
    // DomainName is the domain name to retrieve the authorization code for.
    DomainName string `json:"domainName,omitempty"`
}

// 
type AuthCodeResponse struct {
    // AuthCode is the authorization code needed to transfer a domain to another registrar. If you are storing auth codes, be sure to store them in a secure manner.
    AuthCode string `json:"authCode,omitempty"`
}

// 
type PrivacyRequest struct {
    // DomainName is the domain to purchase Whois Privacy for.
    DomainName string `json:"domainName,omitempty"`
    // PurchasePrice is the amount you expect to pay.
    PurchasePrice float64 `json:"purchasePrice,omitempty"`
    // Years is the number of years you wish to purchase Whois Privacy for. Years defaults to 1 and cannot be more then the domain expiration date.
    Years int32 `json:"years,omitempty"`
    // PromoCode is not yet implemented
    PromoCode string `json:"promoCode,omitempty"`
}

// 
type PrivacyResponse struct {
    // Domain is the status of the domain after the purchase of Whois Privacy.
    Domain *Domain `json:"domain,omitempty"`
    // Order is an identifier for this purchase.
    Order int32 `json:"order,omitempty"`
    // TotalPaid is the total amount paid, including VAT.
    TotalPaid float64 `json:"totalPaid,omitempty"`
}

// 
type SetNameserversRequest struct {
    // DomainName is the domain name to set the nameservers for.
    DomainName string `json:"domainName,omitempty"`
    // Namesevers is a list of the nameservers to set. Nameservers should already be set up and hosting the zone properly as some registries will verify before allowing the change.
    Nameservers []string `json:"nameservers,omitempty"`
}

// 
type SetContactsRequest struct {
    // DomainName is the domain name to set the contacts for.
    DomainName string `json:"domainName,omitempty"`
    // Contacts is the list of contacts to set.
    Contacts *Contacts `json:"contacts,omitempty"`
}

// 
type EnableAutorenewForDomainRequest struct {
    // DomainName is the domain name to enable autorenew for.
    DomainName string `json:"domainName,omitempty"`
}

// 
type DisableAutorenewForDomainRequest struct {
    // DomainName is the domain name to disable autorenew for.
    DomainName string `json:"domainName,omitempty"`
}

// 
type LockDomainRequest struct {
    // DomainName is the domain name to lock.
    DomainName string `json:"domainName,omitempty"`
}

// 
type UnlockDomainRequest struct {
    // DomainName is the domain name to unlock.
    DomainName string `json:"domainName,omitempty"`
}

// 
type EmailForwarding struct {
    // DomainName is the domain part of the email address to forward.
    DomainName string `json:"domainName,omitempty"`
    // EmailBox is the user portion of the email address to forward.
    EmailBox string `json:"emailBox,omitempty"`
    // EmailTo is the entire email address to forward email to.
    EmailTo string `json:"emailTo,omitempty"`
}

// 
type ListEmailForwardingsRequest struct {
    // DomainName is the domain to list email forwarded boxes for.
    DomainName string `json:"domainName,omitempty"`
    // Per Page is the number of records to return per request. Per Page defaults to 1,000.
    PerPage int32 `json:"perPage,omitempty"`
    // Page is which page to return.
    Page int32 `json:"page,omitempty"`
}

// 
type ListEmailForwardingsResponse struct {
    // EmailForwarding is the list of forwarded email boxes.
    EmailForwarding []*EmailForwarding `json:"emailForwarding,omitempty"`
    // NextPage is the identifier for the next page of results. It is only populated if there is another page of results after the current page.
    NextPage int32 `json:"nextPage,omitempty"`
    // LastPage is the identifier for the final page of results. It is only populated if there is another page of results after the current page.
    LastPage int32 `json:"lastPage,omitempty"`
}

// 
type GetEmailForwardingRequest struct {
    // DomainName is the domain to list email forwarded box for.
    DomainName string `json:"domainName,omitempty"`
    // EmailBox is which email box to retrieve.
    EmailBox string `json:"emailBox,omitempty"`
}

// 
type DeleteEmailForwardingRequest struct {
    // DomainName is the domain to delete the email forwarded box from.
    DomainName string `json:"domainName,omitempty"`
    // EmailBox is which email box to delete.
    EmailBox string `json:"emailBox,omitempty"`
}

// HelloRequest doesn't take any parameters.
type HelloRequest struct {
}

// HelloResponse is the response from the HelloFunc command
type HelloResponse struct {
    // ServerName is an identfier for which server is being accessed.
    ServerName string `json:"serverName,omitempty"`
    // Motd is a message of the day. It might provide some useful information.
    Motd string `json:"motd,omitempty"`
    // Username is the account name you are currently logged into.
    Username string `json:"username,omitempty"`
    // ServerTime is the current date/time at the server.
    ServerTime string `json:"serverTime,omitempty"`
}

// 
type Transfer struct {
    // DomainName is the domain to be transfered to Name.com.
    DomainName string `json:"domainName,omitempty"`
    // Email is the email address that the approval email was sent to. Not every TLD requries an approval email. This is usaully pulled from Whois.
    Email string `json:"email,omitempty"`
    // Status is the current status of the transfer. Details about statuses can be found in the following Knowledge Base article: <https://www.name.com/support/articles/115012519688-Transfer-status-FAQ>.
    Status string `json:"status,omitempty"`
}

// 
type ListTransfersRequest struct {
    // Per Page is the number of records to return per request. Per Page defaults to 1,000.
    PerPage int32 `json:"perPage,omitempty"`
    // Page is which page to return
    Page int32 `json:"page,omitempty"`
}

// 
type ListTransfersResponse struct {
    // Transfers is a list of pending transfers
    Transfers []*Transfer `json:"transfers,omitempty"`
    // NextPage is the identifier for the next page of results. It is only populated if there is another page of results after the current page.
    NextPage int32 `json:"nextPage,omitempty"`
    // LastPage is the identifier for the final page of results. It is only populated if there is another page of results after the current page.
    LastPage int32 `json:"lastPage,omitempty"`
}

// 
type GetTransferRequest struct {
    // DomainName is the domain you want to get the transfer information for.
    DomainName string `json:"domainName,omitempty"`
}

// 
type CreateTransferRequest struct {
    // DomainName is the domain you want to transfer to Name.com.
    DomainName string `json:"domainName,omitempty"`
    // AuthCode is the authorization code for the transfer. Not all TLDs require authorization codes, but most do.
    AuthCode string `json:"authCode,omitempty"`
    // PrivacyEnabled is a flag on whether to purchase Whois Privacy with the transfer.
    PrivacyEnabled bool `json:"privacyEnabled,omitempty"`
    // PurchasePrice is the amount to pay for the transfer of the domain. If privacy_enabled is set, the regular price for Whois Privacy will be added automatically. If VAT tax applies, it will also be added automatically.
    // PurchasePrice is required if the domain to transfer is a premium domain.
    PurchasePrice float64 `json:"purchasePrice,omitempty"`
    // PromoCode is not implemented yet
    PromoCode string `json:"promoCode,omitempty"`
}

// 
type CreateTransferResponse struct {
    // Transfer is the transfer resource created.
    Transfer *Transfer `json:"transfer,omitempty"`
    // Order is an identifier for this purchase.
    Order int32 `json:"order,omitempty"`
    // TotalPaid is the total amount paid, including VAT and Whois Privacy.
    TotalPaid float64 `json:"totalPaid,omitempty"`
}

// 
type CancelTransferRequest struct {
    // DomainName is the domain to cancel the transfer for.
    DomainName string `json:"domainName,omitempty"`
}

// 
type VanityNameserver struct {
    // DomainName is the domain the nameserver is a subdomain of.
    DomainName string `json:"domainName,omitempty"`
    // Hostname is the hostname of the nameserver.
    Hostname string `json:"hostname,omitempty"`
    // IPs is a list of IP addresses that are used for glue records for this nameserver.
    Ips []string `json:"ips,omitempty"`
}

// 
type ListVanityNameserversRequest struct {
    // DomainName is the domain to list the vanity nameservers for.
    DomainName string `json:"domainName,omitempty"`
    // Per Page is the number of records to return per request. Per Page defaults to 1,000.
    PerPage int32 `json:"perPage,omitempty"`
    // Page is which page to return
    Page int32 `json:"page,omitempty"`
}

// 
type ListVanityNameserversResponse struct {
    // VanityNameservers is the list of vanity nameservers.
    VanityNameservers []*VanityNameserver `json:"vanityNameservers,omitempty"`
    // NextPage is the identifier for the next page of results. It is only populated if there is another page of results after the current page.
    NextPage int32 `json:"nextPage,omitempty"`
    // LastPage is the identifier for the final page of results. It is only populated if there is another page of results after the current page.
    LastPage int32 `json:"lastPage,omitempty"`
}

// 
type GetVanityNameserverRequest struct {
    // DomainName is the domain to for the vanity nameserver.
    DomainName string `json:"domainName,omitempty"`
    // Hostname is the hostname for the vanity nameserver.
    Hostname string `json:"hostname,omitempty"`
}

// 
type DeleteVanityNameserverRequest struct {
    // DomainName is the domain of the vanity nameserver to delete.
    DomainName string `json:"domainName,omitempty"`
    // Hostname is the hostname of the vanity nameserver to delete.
    Hostname string `json:"hostname,omitempty"`
}
