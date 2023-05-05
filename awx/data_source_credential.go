/*
Use this data source to query Credential by ID.

# Example Usage

```hcl
*TBD*
```
*/
package awx

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awx "github.com/mrcrilly/goawx/client"
)

func dataSourceCredentialByID() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCredentialsRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
		},
	}
}

func dataSourceCredentialsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*awx.AWX)
	params := make(map[string]string)

	if name, okName := d.GetOk("name"); okName {
		params["name"] = name.(string)
	}

	if len(params) == 0 {
		return buildDiagnosticsMessage(
			"Get: Missing Parameters",
			"Please use one of the selectors (name or id)")
	}

	credentials, _, err := client.CredentialsService.ListCredentials(map[string]string{})

	if err != nil {
		return buildDiagnosticsMessage(
			"Get: Fail to fetch Credential list",
			"Fail to find the Credential List, got: %s",
			err)
	}

	for _, credential := range credentials {
		if credential.Name == params["name"] {
			d.Set("id", credential.ID)

			return diags
		}
	}

	return buildDiagnosticsMessage(
		"Credential not found",
		"Could not find credential with name: %s",
		params["name"])
}
