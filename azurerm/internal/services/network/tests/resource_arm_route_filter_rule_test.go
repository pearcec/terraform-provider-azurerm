package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMRouteFilterRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_route_filter_rule", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMRouteFilterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMRouteFilterRule_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRouteFilterRuleExists("azurerm_route_filter_rule.test"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMRouteFilterRule_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_route_filter_rule", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMRouteFilterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMRouteFilterRule_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRouteFilterRuleExists(data.ResourceName),
				),
			},
			{
				Config:      testAccAzureRMRouteFilterRule_requiresImport(data),
				ExpectError: acceptance.RequiresImportError("azurerm_route_filter_rule"),
			},
		},
	})
}

func TestAccAzureRMRouteFilterRule_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_route_filter_rule", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMRouteFilterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMRouteFilterRule_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRouteFilterRuleExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "access", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule_type", "Community"),
					resource.TestCheckResourceAttr(data.ResourceName, "communities.0", "12076:53004"),
				),
			},
			{
				Config: testAccAzureRMRouteFilterRule_basicDeny(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRouteFilterRuleExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "access", "Deny"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule_type", "Community"),
					resource.TestCheckResourceAttr(data.ResourceName, "communities.0", "12076:52004"),
				),
			},
			{
				Config: testAccAzureRMRouteFilterRule_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRouteFilterRuleExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "access", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule_type", "Community"),
					resource.TestCheckResourceAttr(data.ResourceName, "communities.0", "12076:53004"),
				),
			},
		},
	})
}

func TestAccAzureRMRouteFilterRule_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_route_filter_rule", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMRouteFilterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMRouteFilterRule_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRouteFilterRuleExists("azurerm_route_filter_rule.test"),
					testCheckAzureRMRouteFilterRuleDisappears("azurerm_route_filter_rule.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureRMRouteFilterRule_multipleRouteFilterRules(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_route_filter_rule", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMRouteFilterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMRouteFilterRule_multipleRouteFilterRules(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMRouteFilterRuleExists("azurerm_route_filter_rule.test1"),
					testCheckAzureRMRouteFilterRuleExists("azurerm_route_filter_rule.test2"),
				),
			},
		},
	})
}

func testCheckAzureRMRouteFilterRuleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Network.RouteFilterRulesClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %q", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		rtName := rs.Primary.Attributes["route_filter_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for route_filter_rule: %q", name)
		}

		resp, err := client.Get(ctx, resourceGroup, rtName, name)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: RouteFilterRule %q (resource group: %q) does not exist", name, resourceGroup)
			}
			return fmt.Errorf("Bad: Get on route_filter_rulesClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMRouteFilterRuleDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Network.RouteFilterRulesClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		rtName := rs.Primary.Attributes["route_filter_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for route_filter_rule: %s", name)
		}

		future, err := client.Delete(ctx, resourceGroup, rtName, name)
		if err != nil {
			return fmt.Errorf("Error deleting RouteFilterRule %q (RouteFilterRule Table %q / Resource Group %q): %+v", name, rtName, resourceGroup, err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("Error waiting for deletion of RouteFilterRule %q (RouteFilterRule Table %q / Resource Group %q): %+v", name, rtName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureRMRouteFilterRuleDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).Network.RouteFilterRulesClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_route_filter_rule" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		rtName := rs.Primary.Attributes["route_filter_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, rtName, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("RouteFilterRule still exists:\n%#v", resp.RouteFilterRulePropertiesFormat)
		}
	}

	return nil
}

func testAccAzureRMRouteFilterRule_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_route_filter" "test" {
  name                = "acctestrf%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_route_filter_rule" "test" {
  name                = "acctestroute_filter_rule%d"
  resource_group_name = azurerm_resource_group.test.name
  route_filter_name   = azurerm_route_filter.test.name

  access      = "Allow"
  rule_type   = "Community"
  communities = ["12076:53004"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMRouteFilterRule_requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
resource "azurerm_route_filter_rule" "import" {
  name                = azurerm_route_filter_rule.test.name
  resource_group_name = azurerm_route_filter_rule.test.resource_group_name
  route_filter_name   = azurerm_route_filter_rule.test.route_filter_name

  access      = azurerm_route_filter_rule.test.access
  rule_type   = azurerm_route_filter_rule.test.rule_type
  communities = azurerm_route_filter_rule.test.communities
}
`, testAccAzureRMRouteFilterRule_basic(data))
}

func testAccAzureRMRouteFilterRule_basicDeny(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_route_filter" "test" {
  name                = "acctestrf%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_route_filter_rule" "test" {
  name                = "acctestroute_filter_rule%d"
  resource_group_name = azurerm_resource_group.test.name
  route_filter_name    = azurerm_route_filter.test.name

  access      = "Deny"
  rule_type   = "Community"
  communities = ["12076:52004"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMRouteFilterRule_multipleRouteFilterRules(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_route_filter" "test" {
  name                = "acctestrf%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_route_filter_rule" "test1" {
  name                = "acctestroute_filter_rule1_%d"
  resource_group_name = azurerm_resource_group.test.name
  route_filter_name   = azurerm_route_filter.test.name

  access      = "Allow"
  rule_type   = "Community"
  communities = ["12076:52005","12076:52006"]
}

resource "azurerm_route_filter_rule" "test2" {
  name                = "acctestroute_filter_rule2_%d"
  resource_group_name = azurerm_resource_group.test.name
  route_filter_name   = azurerm_route_filter.test.name

  access      = "Deny"
  rule_type   = "Community"
  communities = ["12076:53005","12076:53006"]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
