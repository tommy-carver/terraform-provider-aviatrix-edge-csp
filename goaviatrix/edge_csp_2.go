package goaviatrix

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"log"
	"reflect"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type EdgePlatform struct {
	Action         string                   `json:"action,omitempty"`
	CID            string                   `json:"CID,omitempty"`
	GwName         string                   `json:"name,omitempty"`
	SiteId         string                   `json:"site_id,omitempty"`
	ProjectID      string                   `json:"project_id,omitempty"`
	DeviceID       string                   `json:"device_id,omitempty"`
	Dhcp           bool                     `json:"dhcp,omitempty"`
	LocalAsNumber  string                   `json:"local_as_number"`
	WanPublicIp    string                   `json:"wan_discovery_ip"`
	NoProgressBar  bool                     `json:"no_progress_bar,omitempty"`
	WanInterfaces  string                   `json:"wan_ifnames"`
	LanInterfaces  string                   `json:"lan_ifnames"`
	MgmtInterfaces string                   `json:"mgmt_ifnames"`
	LANIP          string                   `json:"lan_ip"`
	InterfaceList  []*EdgePlatformInterface `json:"interfaces"`
}

type EdgePlatformInterface struct {
	IfName       string `json:"ifname"`
	Type         string `json:"type"`
	PublicIp     string `json:"public_ip"`
	Tag          string `json:"tag"`
	Dhcp         bool   `json:"dhcp"`
	IpAddr       string `json:"ipaddr"`
	GatewayIp    string `json:"gateway_ip"`
	DnsPrimary   string `json:"dns_primary"`
	DnsSecondary string `json:"dns_secondary"`
	AdminState   string `json:"admin_state"`
}

type EdgePlatformResponse struct {
	GwName    string `json:"name"`
	SiteId    string `json:"site_id"`
	ProjectID string `json:"project_id"`
	DeviceID  string `json:"device_id"`
	VpcState  string `json:"vpc_state"`
	VpcID     string `json:"vpc_id"`
	Status    string `json:"status"`
}

type EdgePlatformListResp struct {
	Return  bool                 `json:"return"`
	Results EdgePlatformResponse `json:"results"`
	Reason  string               `json:"reason"`
}

func (c *Client) CreateAvxEdgeCSP(ctx context.Context, edgePlatform *EdgePlatform) error {
	edgePlatform.Action = "create_edge_csp_gateway"
	edgePlatform.CID = c.CID
	edgePlatform.NoProgressBar = true

	if edgePlatform.Dhcp {
		edgePlatform.Dhcp = true
	}

	err := c.PostAPIContext2(ctx, nil, edgePlatform.Action, edgePlatform, BasicCheck)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetEdgeCSPAvx(ctx context.Context, gwName string) (*EdgePlatformListResp, error) {
	form := map[string]string{
		"action":       "get_gateway_info",
		"CID":          c.CID,
		"gateway_name": gwName,
	}

	var data EdgePlatformListResp
	timeout := time.After(time.Minute * 20)
	tick := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-timeout:
			return nil, ErrNotFound
		case <-tick.C:
			if err := c.GetAPI(&data, form["action"], form, BasicCheck); err == nil {
				return &data, nil
			} else {
				log.Printf("[DEBUG] Didn't get results")
			}
		}
	}
}

func (c *Client) DeleteAvxEdgeCSPAvx(ctx context.Context, project_id, device_id, name string) error {
	form := map[string]string{
		"action":     "delete_edge_csp_gateway",
		"CID":        c.CID,
		"project_id": project_id,
		"device_id":  device_id,
		"name":       name,
	}

	return c.PostAPIContext2(ctx, nil, form["action"], form, BasicCheck)
}

func (c *Client) UpdateAvxEdgeCSP(ctx context.Context, edgePlatform *EdgePlatform) error {
	form := map[string]string{
		"action": "update_edge_gateway",
		"CID":    c.CID,
		"name":   edgePlatform.GwName,
	}

	if edgePlatform.InterfaceList != nil && len(edgePlatform.InterfaceList) != 0 {
		interfaces, err := json.Marshal(edgePlatform.InterfaceList)
		if err != nil {
			return err
		}

		form["interfaces"] = b64.StdEncoding.EncodeToString(interfaces)
	}

	return c.PostAPIContext2(ctx, nil, form["action"], form, BasicCheck)
}

func DiffSuppressFuncInterfacesEdgePlatform(k, old, new string, d *schema.ResourceData) bool {
	ifOld, ifNew := d.GetChange("interfaces")
	var interfacesOld []map[string]interface{}

	for _, if0 := range ifOld.([]interface{}) {
		if1 := if0.(map[string]interface{})
		interfacesOld = append(interfacesOld, if1)
	}

	var interfacesNew []map[string]interface{}

	for _, if0 := range ifNew.([]interface{}) {
		if1 := if0.(map[string]interface{})
		interfacesNew = append(interfacesNew, if1)
	}

	sort.Slice(interfacesOld, func(i, j int) bool {
		return interfacesOld[i]["ifname"].(string) < interfacesOld[j]["ifname"].(string)
	})

	sort.Slice(interfacesNew, func(i, j int) bool {
		return interfacesNew[i]["ifname"].(string) < interfacesNew[j]["ifname"].(string)
	})

	return reflect.DeepEqual(interfacesOld, interfacesNew)
}
