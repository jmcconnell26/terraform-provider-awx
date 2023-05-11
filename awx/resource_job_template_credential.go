/*
*TBD*

# Example Usage

```hcl

	resource "awx_job_template_credentials" "baseconfig" {
	  job_template_id = awx_job_template.baseconfig.id
	  credential_id   = awx_credential_machine.pi_connection.id
	}

```
*/
package awx

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awx "github.com/mrcrilly/goawx/client"
	"strconv"
	"strings"
)

func resourceJobTemplateCredentials() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceJobTemplateCredentialsCreate,
		DeleteContext: resourceJobTemplateCredentialsDelete,
		ReadContext:   resourceJobTemplateCredentialsRead,

		Schema: map[string]*schema.Schema{
			"job_template_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"credential_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")

				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid ID specified. Supplied ID must be written as <job_template_id>/<credential_id>")
				}

				jobTemplateId, err := strconv.Atoi(parts[0])

				if err != nil {
					return nil, fmt.Errorf("failed to parse Job Template ID")
				}

				credentialId, err := strconv.Atoi(parts[1])

				if err != nil {
					return nil, fmt.Errorf("failed to parse Credential ID")
				}

				client := m.(*awx.AWX)
				res, err := client.JobTemplateService.GetJobTemplateByID(jobTemplateId, make(map[string]string))

				if err != nil {
					return nil, fmt.Errorf("failed to load Job Template with ID: %d", jobTemplateId)
				}

				d.SetId(fmt.Sprintf("%d/%d", jobTemplateId, credentialId))
				d.Set("credential_id", res.Credential)

				return []*schema.ResourceData{d}, nil
			},
		},
	}
}

func resourceJobTemplateCredentialsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*awx.AWX)
	awxService := client.JobTemplateService
	jobTemplateID := d.Get("job_template_id").(int)
	_, err := awxService.GetJobTemplateByID(jobTemplateID, make(map[string]string))
	if err != nil {
		return buildDiagNotFoundFail("job template", jobTemplateID, err)
	}

	credentialID := d.Get("credential_id").(int)
	_, err = awxService.AssociateCredentials(jobTemplateID, map[string]interface{}{
		"id": credentialID,
	}, map[string]string{})

	if err != nil {
		return buildDiagnosticsMessage("Create: JobTemplate not AssociateCredentials", "Fail to add credentials with Id %v, for Template ID %v, got error: %s", d.Get("credential_id").(int), jobTemplateID, err.Error())
	}

	d.SetId(fmt.Sprintf("%d/%d", jobTemplateID, credentialID))
	return diags
}

func resourceJobTemplateCredentialsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*awx.AWX)
	awxService := client.JobTemplateService
	parts := strings.Split(d.Id(), "/")

	jobTemplateId, err := strconv.Atoi(parts[0])
	if err != nil {
		return buildDiagnosticsMessage(
			fmt.Sprintf("%s, state ID not converted", d.Id()), "Value in State %s is unparseable, %s", d.Id())
	}

	res, err := awxService.GetJobTemplateByID(jobTemplateId, make(map[string]string))
	if err != nil {
		return buildDiagNotFoundFail("job template", jobTemplateId, err)
	}

	d.SetId(fmt.Sprintf("%d/%d", res.ID, res.Credential))
	d.Set("job_template_id", res.ID)
	d.Set("credential_id", res.Credential)
	return diags
}

func resourceJobTemplateCredentialsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*awx.AWX)
	awxService := client.JobTemplateService
	jobTemplateID := d.Get("job_template_id").(int)
	res, err := awxService.GetJobTemplateByID(jobTemplateID, make(map[string]string))
	if err != nil {
		return buildDiagNotFoundFail("job template", jobTemplateID, err)
	}

	_, err = awxService.DisAssociateCredentials(res.ID, map[string]interface{}{
		"id": d.Get("credential_id").(int),
	}, map[string]string{})
	if err != nil {
		return buildDiagDeleteFail("JobTemplate DisAssociateCredentials", fmt.Sprintf("DisAssociateCredentials %v, from JobTemplateID %v got %s ", d.Get("credential_id").(int), d.Get("job_template_id").(int), err.Error()))
	}

	d.SetId("")
	return diags
}
