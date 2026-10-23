package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainRulesSupportMultipleProxyOutboundsAndLookup(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"dns": {
			"final": "dns-direct",
			"rules": [
				{"action":"route","server":"dns-remote","domain_suffix":["example.com"]},
				{"action":"route","server":"dns-direct"}
			],
			"servers": []
		},
		"outbounds": [
			{"type":"selector","tag":"proxy","outbounds":["germany","france"]},
			{"type":"vless","tag":"germany"},
			{"type":"vless","tag":"france"},
			{"type":"vless","tag":"rogue"},
			{"type":"direct","tag":"direct"}
		],
		"route": {
			"final": "direct",
			"rules": [
				{"action":"sniff","timeout":"1s"},
				{"domain_suffix":["example.com"],"outbound":"proxy"},
				{"outbound":"direct"}
			]
		}
	}`), 0o600))
	t.Setenv("CONFIG_PATH", configPath)

	router := http.NewServeMux()
	router.Handle("PUT /dns-rules", AddDNSRuleHandler())
	router.Handle("DELETE /dns-rules", RemoveDNSRuleHandler())
	router.Handle("PUT /ip-rules", AddIPRuleHandler())
	router.Handle("GET /route-rules/assignments", GetRuleAssignmentsHandler())

	addGermany := serveJSONRequest(router, http.MethodPut, "/dns-rules?norestart=true",
		`{"routeMode":"proxy","outbound":"germany","domain":"domain:example.com"}`)
	require.Equal(t, http.StatusCreated, addGermany.Code, addGermany.Body.String())
	addFrance := serveJSONRequest(router, http.MethodPut, "/dns-rules?norestart=true",
		`{"routeMode":"proxy","outbound":"france","domain":"domain:example.com"}`)
	require.Equal(t, http.StatusCreated, addFrance.Code, addFrance.Body.String())

	lookup := serveJSONRequest(router, http.MethodGet,
		"/route-rules/assignments?domain=domain%3Aexample.com", "")
	require.Equal(t, http.StatusOK, lookup.Code, lookup.Body.String())
	assert.JSONEq(t, `{
		"kind":"domain",
		"value":"domain:example.com",
		"assignments":[
			{"routeMode":"proxy","outbound":"germany","order":1},
			{"routeMode":"proxy","outbound":"france","order":2},
			{"routeMode":"proxy","outbound":"proxy","order":3}
		]
	}`, lookup.Body.String())

	assertRouteOutboundOrder(t, configPath, []string{"", "germany", "france", "proxy", "direct"})
	assertDomainDNSDestination(t, configPath, "example.com", "dns-remote")

	removeGermany := serveJSONRequest(router, http.MethodDelete, "/dns-rules?norestart=true",
		`{"routeMode":"proxy","outbound":"germany","domain":"domain:example.com"}`)
	require.Equal(t, http.StatusNoContent, removeGermany.Code, removeGermany.Body.String())
	assertRouteOutboundOrder(t, configPath, []string{"", "france", "proxy", "direct"})
	assertDomainDNSDestination(t, configPath, "example.com", "dns-remote")

	addDirect := serveJSONRequest(router, http.MethodPut, "/dns-rules?norestart=true",
		`{"routeMode":"direct","domain":"domain:example.com"}`)
	require.Equal(t, http.StatusCreated, addDirect.Code, addDirect.Body.String())
	assertRouteOutboundOrder(t, configPath, []string{"", "direct", "france", "proxy", "direct"})
	assertDomainDNSDestination(t, configPath, "example.com", "dns-direct")

	for _, requestBody := range []string{
		`{"routeMode":"proxy","outbound":"","domain":"domain:blank.example"}`,
		`{"routeMode":"proxy","outbound":"   ","domain":"domain:whitespace.example"}`,
	} {
		response := serveJSONRequest(router, http.MethodPut, "/dns-rules?norestart=true", requestBody)
		require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	}
	for _, domain := range []string{"blank.example", "whitespace.example"} {
		lookup := serveJSONRequest(router, http.MethodGet,
			"/route-rules/assignments?domain=domain%3A"+domain, "")
		require.Equal(t, http.StatusOK, lookup.Code, lookup.Body.String())
		assert.Contains(t, lookup.Body.String(), `"outbound": "proxy"`)
	}

	missing := serveJSONRequest(router, http.MethodPut, "/dns-rules?norestart=true",
		`{"routeMode":"proxy","outbound":"missing","domain":"domain:missing.example"}`)
	assert.Equal(t, http.StatusBadRequest, missing.Code)
	assert.Contains(t, missing.Body.String(), "outbound tag 'missing' does not exist")

	nonMember := serveJSONRequest(router, http.MethodPut, "/dns-rules?norestart=true",
		`{"routeMode":"proxy","outbound":"rogue","domain":"domain:rogue.example"}`)
	assert.Equal(t, http.StatusBadRequest, nonMember.Code)
	assert.Contains(t, nonMember.Body.String(), "is not a member of selector 'proxy'")

	invalidBlock := serveJSONRequest(router, http.MethodPut, "/dns-rules?norestart=true",
		`{"routeMode":"block","outbound":"germany","domain":"domain:blocked.example"}`)
	assert.Equal(t, http.StatusBadRequest, invalidBlock.Code)
	assert.Contains(t, invalidBlock.Body.String(), "outbound can only be specified")

	addIP := serveJSONRequest(router, http.MethodPut, "/ip-rules?norestart=true",
		`{"routeMode":"proxy","outbound":"france","ip":"192.0.2.0/24"}`)
	require.Equal(t, http.StatusCreated, addIP.Code, addIP.Body.String())
	ipLookup := serveJSONRequest(router, http.MethodGet,
		"/route-rules/assignments?ip=192.0.2.0%2F24", "")
	require.Equal(t, http.StatusOK, ipLookup.Code, ipLookup.Body.String())
	assert.Contains(t, ipLookup.Body.String(), `"outbound": "france"`)
}

func TestCoveringRuleAssignments(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"dns":{"rules":[],"servers":[]},
		"outbounds":[
			{"type":"selector","tag":"proxy","outbounds":["germany"]},
			{"type":"vless","tag":"germany"},
			{"type":"direct","tag":"direct"}
		],
		"route":{"rules":[
			{"action":"sniff","timeout":"1s"},
			{"action":"reject","domain_keyword":["facebook"]},
			{"outbound":"direct","domain":["facebook.com"],"ip_cidr":["145.67.23.64/26"]},
			{"outbound":"germany","domain_regex":["^.*facebook\\.com$"],"ip_cidr":["145.67.0.0/16"]},
			{"outbound":"proxy","domain_suffix":["facebook.com"],"ip_cidr":["145.0.0.0/8","145.67.23.65"]}
		]}
	}`), 0o600))
	t.Setenv("CONFIG_PATH", configPath)

	router := http.NewServeMux()
	router.Handle("GET /route-rules/assignments", GetRuleAssignmentsHandler())

	domainLookup := serveJSONRequest(router, http.MethodGet,
		"/route-rules/assignments?domain=facebook.com&match=covering", "")
	require.Equal(t, http.StatusOK, domainLookup.Code, domainLookup.Body.String())
	assert.JSONEq(t, `{
		"kind":"domain",
		"value":"facebook.com",
		"match":"covering",
		"matches":[
			{"type":"domain_keyword","pattern":"facebook","routeMode":"block","order":1},
			{"type":"domain","pattern":"facebook.com","routeMode":"direct","outbound":"direct","order":2},
			{"type":"domain_regex","pattern":"^.*facebook\\.com$","routeMode":"proxy","outbound":"germany","order":3},
			{"type":"domain_suffix","pattern":"facebook.com","routeMode":"proxy","outbound":"proxy","order":4}
		]
	}`, domainLookup.Body.String())

	ipLookup := serveJSONRequest(router, http.MethodGet,
		"/route-rules/assignments?ip=145.67.23.65&match=covering", "")
	require.Equal(t, http.StatusOK, ipLookup.Code, ipLookup.Body.String())
	assert.JSONEq(t, `{
		"kind":"ip",
		"value":"145.67.23.65",
		"match":"covering",
		"matches":[
			{"type":"ip_cidr","pattern":"145.67.23.64/26","routeMode":"direct","outbound":"direct","order":2},
			{"type":"ip_cidr","pattern":"145.67.0.0/16","routeMode":"proxy","outbound":"germany","order":3},
			{"type":"ip_cidr","pattern":"145.0.0.0/8","routeMode":"proxy","outbound":"proxy","order":4},
			{"type":"ip_cidr","pattern":"145.67.23.65","routeMode":"proxy","outbound":"proxy","order":4}
		]
	}`, ipLookup.Body.String())

	cidrLookup := serveJSONRequest(router, http.MethodGet,
		"/route-rules/assignments?ip=145.67.0.0%2F16&match=covering", "")
	require.Equal(t, http.StatusOK, cidrLookup.Code, cidrLookup.Body.String())
	assert.JSONEq(t, `{
		"kind":"ip",
		"value":"145.67.0.0/16",
		"match":"covering",
		"matches":[
			{"type":"ip_cidr","pattern":"145.67.0.0/16","routeMode":"proxy","outbound":"germany","order":3}
		]
	}`, cidrLookup.Body.String())

	invalidMode := serveJSONRequest(router, http.MethodGet,
		"/route-rules/assignments?domain=facebook.com&match=partial", "")
	assert.Equal(t, http.StatusBadRequest, invalidMode.Code)
	assert.Contains(t, invalidMode.Body.String(), "match must be 'exact' or 'covering'")

	invalidDomainType := serveJSONRequest(router, http.MethodGet,
		"/route-rules/assignments?domain=keyword%3Afacebook&match=covering", "")
	assert.Equal(t, http.StatusBadRequest, invalidDomainType.Code)
	assert.Contains(t, invalidDomainType.Body.String(), "covering domain lookup requires a hostname")
}

func assertRouteOutboundOrder(t *testing.T, configPath string, expected []string) {
	t.Helper()
	contents, err := os.ReadFile(configPath)
	require.NoError(t, err)
	var saved struct {
		Route struct {
			Rules []struct {
				Outbound string `json:"outbound"`
			} `json:"rules"`
		} `json:"route"`
	}
	require.NoError(t, json.Unmarshal(contents, &saved))
	actual := make([]string, len(saved.Route.Rules))
	for index, rule := range saved.Route.Rules {
		actual[index] = rule.Outbound
	}
	assert.Equal(t, expected, actual)
}

func assertDomainDNSDestination(t *testing.T, configPath, domain, expectedServer string) {
	t.Helper()
	contents, err := os.ReadFile(configPath)
	require.NoError(t, err)
	var saved struct {
		DNS struct {
			Rules []struct {
				Server       string   `json:"server"`
				DomainSuffix []string `json:"domain_suffix"`
			} `json:"rules"`
		} `json:"dns"`
	}
	require.NoError(t, json.Unmarshal(contents, &saved))
	found := ""
	for _, rule := range saved.DNS.Rules {
		for _, value := range rule.DomainSuffix {
			if value == domain {
				if found != "" {
					t.Fatalf("domain %s is present in multiple DNS rules", domain)
				}
				found = rule.Server
			}
		}
	}
	assert.Equal(t, expectedServer, found)
}
