package aviatrix

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAviatrixEdgePlatform() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAviatrixEdgePlatformCreate,
		ReadWithoutTimeout:   resourceAviatrixEdgePlatformRead,
		UpdateWithoutTimeout: resourceAviatrixEdgePlatformUpdate,
		DeleteWithoutTimeout: resourceAviatrixEdgePlatformDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge CSP gw name.",
			},
			"site_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Site ID.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge CSP project UUID.",
			},
			"device_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge CSP device UUID.",
			},
			"wan_ifnames": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "WAN interface name.",
				DefaultFunc: func() (any, error) {
					return []string{"eth0"}, nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"lan_ifnames": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "LAN interface name.",
				DefaultFunc: func() (any, error) {
					return []string{"eth1"}, nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mgmt_ifnames": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "Management interface name.",
				DefaultFunc: func() (any, error) {
					return []string{"eth2"}, nil
				},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"lan_ip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "LAN interface IP.",
			},
			"dhcp": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Use dhcp for managemnet config.",
			},
			"interfaces": {
				Type:             schema.TypeList,
				Required:         true,
				Description:      "",
				DiffSuppressFunc: goaviatrix.DiffSuppressFuncInterfacesEdgePlatform,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ifname": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"public_ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"dhcp": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "",
						},
						"ipaddr": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"gateway_ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"dns_primary": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"dns_secondary": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"admin_state": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "",
						},
					},
				},
			},
		},
	}
}

func marshalAvxEdgeCSPInput(d *schema.ResourceData) *goaviatrix.EdgePlatform {
	edgePlatform := &goaviatrix.EdgePlatform{
		GwName:    d.Get("name").(string),
		SiteId:    d.Get("site_id").(string),
		ProjectID: d.Get("project_id").(string),
		DeviceID:  d.Get("device_id").(string),
		// LocalAsNumber:  d.Get("local_as_number").(string),
		WanInterfaces:  strings.Join(getStringList(d, "wan_ifnames"), ","),
		LanInterfaces:  strings.Join(getStringList(d, "lan_ifnames"), ","),
		MgmtInterfaces: strings.Join(getStringList(d, "mgmt_ifnames"), ","),
		LANIP:          d.Get("lan_ip").(string),
		Dhcp:           d.Get("dhcp").(bool),
	}

	interfaces := d.Get("interfaces").([]interface{})
	for _, if0 := range interfaces {
		if1 := if0.(map[string]interface{})

		if2 := &goaviatrix.EdgePlatformInterface{
			IfName:       if1["ifname"].(string),
			Type:         if1["type"].(string),
			PublicIp:     if1["public_ip"].(string),
			Tag:          if1["tag"].(string),
			Dhcp:         if1["dhcp"].(bool),
			IpAddr:       if1["ipaddr"].(string),
			GatewayIp:    if1["gateway_ip"].(string),
			DnsPrimary:   if1["dns_primary"].(string),
			DnsSecondary: if1["dns_secondary"].(string),
		}

		if if1["admin_state"].(bool) {
			if2.AdminState = "enabled"
		} else {
			if2.AdminState = "disabled"
		}

		edgePlatform.InterfaceList = append(edgePlatform.InterfaceList, if2)
	}

	return edgePlatform
}

func resourceAviatrixEdgePlatformCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*goaviatrix.Client)

	// read configs
	avxEdgeCSP := marshalAvxEdgeCSPInput(d)

	// create
	d.SetId(avxEdgeCSP.GwName)
	flag := false
	if err := client.CreateAvxEdgeCSP(ctx, avxEdgeCSP); err != nil {
		return diag.Errorf("could not create Edge CSP: %v", err)
	}
	resourceAviatrixEdgePlatformReadIfRequired(ctx, d, meta, &flag)
	if err := client.UpdateAvxEdgeCSP(ctx, avxEdgeCSP); err != nil {
		return diag.Errorf("could not update interfaces or DNS profile name during Edge CSP update: %v", err)
	}

	return resourceAviatrixEdgePlatformReadIfRequired(ctx, d, meta, &flag)
}

func resourceAviatrixEdgePlatformReadIfRequired(ctx context.Context, d *schema.ResourceData, meta interface{}, flag *bool) diag.Diagnostics {
	if !(*flag) {
		*flag = true
		return resourceAviatrixEdgePlatformRead(ctx, d, meta)
	}
	return nil
}

func resourceAviatrixEdgePlatformRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*goaviatrix.Client)

	// handle import
	if d.Get("name").(string) == "" {
		id := d.Id()
		log.Printf("[DEBUG] Looks like an import, no name received. Import Id is %s", id)
		d.Set("name", id)
		d.SetId(id)
	}

	avxEdgeCSPResp, err := client.GetEdgeCSPAvx(ctx, d.Get("name").(string))
	if err != nil {
		if err == goaviatrix.ErrNotFound {
			d.SetId("")
			log.Printf("[DEBUG] could not find gw")
			return nil
		}
		return diag.Errorf("could not read Edge CSP: %v", err)
	}
	d.Set("name", avxEdgeCSPResp.Results.GwName)
	d.Set("site_id", avxEdgeCSPResp.Results.VpcID)
	d.SetId(avxEdgeCSPResp.Results.GwName)
	return nil
}

func resourceAviatrixEdgePlatformUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*goaviatrix.Client)

	// read configs
	edgeCSP := marshalAvxEdgeCSPInput(d)

	d.Partial(true)

	if d.HasChange("interfaces") {
		err := client.UpdateAvxEdgeCSP(ctx, edgeCSP)
		if err != nil {
			return diag.Errorf("could not update WAN/LAN/VLAN interfaces or DNS profile name during Edge CSP update: %v", err)
		}
	}

	d.Partial(false)

	return resourceAviatrixEdgePlatformRead(ctx, d, meta)
}

func resourceAviatrixEdgePlatformDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*goaviatrix.Client)

	project_id := d.Get("project_id").(string)
	device_id := d.Get("device_id").(string)
	gwName := d.Get("name").(string)

	err := client.DeleteAvxEdgeCSPAvx(ctx, project_id, device_id, gwName)
	if err != nil {
		return diag.Errorf("could not delete Edge CSP: %v", err)
	}

	return nil
}
