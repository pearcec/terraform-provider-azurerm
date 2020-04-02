package network

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-09-01/network"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmRouteFilterRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmRouteFilterRuleCreateUpdate,
		Read:   resourceArmRouteFilterRuleRead,
		Update: resourceArmRouteFilterRuleCreateUpdate,
		Delete: resourceArmRouteFilterRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"route_filter_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"access": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.Allow),
					string(network.Deny),
				}, false),
			},

			"rule_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Community",
				}, false),
			},

			"communities": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func resourceArmRouteFilterRuleCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RouteFilterRulesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	rfName := d.Get("route_filter_name").(string)
	resGroup := d.Get("resource_group_name").(string)

	if features.ShouldResourcesBeImported() && d.IsNewResource() {
		existing, err := client.Get(ctx, resGroup, rfName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Route Filter Rule %q (Route Filter %q / Resource Group %q): %+v", name, rfName, resGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_route_filter_rule", *existing.ID)
		}
	}

	access := d.Get("access").(string)
	ruleType := d.Get("rule_type").(string)
	communities := utils.ExpandStringSlice(d.Get("communities").([]interface{}))

	locks.ByName(rfName, routeFilterResourceName)
	defer locks.UnlockByName(rfName, routeFilterResourceName)

	rule := network.RouteFilterRule{
		Name: &name,
		RouteFilterRulePropertiesFormat: &network.RouteFilterRulePropertiesFormat{
			Access:              network.Access(access),
			RouteFilterRuleType: &ruleType,
			Communities:         communities,
		},
	}

	future, err := client.CreateOrUpdate(ctx, resGroup, rfName, name, rule)
	if err != nil {
		return fmt.Errorf("Error Creating/Updating Route Filter Rule %q (Route Filter %q / Resource Group %q): %+v", name, rfName, resGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for completion for Route Filter Rule %q (Route Filter %q / Resource Group %q): %+v", name, rfName, resGroup, err)
	}

	read, err := client.Get(ctx, resGroup, rfName, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Route Filter Rule %q (Route Filter %q / Resource Group %q) ID", rfName, name, resGroup)
	}
	d.SetId(*read.ID)

	return resourceArmRouteFilterRuleRead(d, meta)
}

func resourceArmRouteFilterRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RouteFilterRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	rfName := id.Path["routeFilters"]
	ruleName := id.Path["routeFilterRules"]

	resp, err := client.Get(ctx, resGroup, rfName, ruleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure Route Filter Rule %q: %+v", ruleName, err)
	}

	d.Set("name", ruleName)
	d.Set("resource_group_name", resGroup)
	d.Set("route_filter_name", rfName)

	if props := resp.RouteFilterRulePropertiesFormat; props != nil {
		d.Set("access", string(props.Access))
		d.Set("rule_type", props.RouteFilterRuleType)
		d.Set("communities", utils.FlattenStringSlice(props.Communities))
	}

	return nil
}

func resourceArmRouteFilterRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.RouteFilterRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	rfName := id.Path["routeFilters"]
	ruleName := id.Path["routeFilterRules"]

	locks.ByName(rfName, routeFilterResourceName)
	defer locks.UnlockByName(rfName, routeFilterResourceName)

	future, err := client.Delete(ctx, resGroup, rfName, ruleName)
	if err != nil {
		return fmt.Errorf("Error deleting Route Filter Rule %q (Route Filter %q / Resource Group %q): %+v", ruleName, rfName, resGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for deletion of RouteFilterRule %q (RouteFilterRule Table %q / Resource Group %q): %+v", ruleName, rfName, resGroup, err)
	}

	return nil
}
