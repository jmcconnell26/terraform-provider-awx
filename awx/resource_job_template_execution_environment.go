package awx

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awx "github.com/mrcrilly/goawx/client"
)

func resourceJobTemplateExecutionEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceJobTemplateExecutionEnvironmentCreate,
		DeleteContext: resourceJobTemplateExecutionEnvironmentDelete,
		ReadContext:   resourceJobTemplateExecutionEnvironmentRead,

		Schema: map[string]*schema.Schema{
			"job_template_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"execution_environment_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceJobTemplateExecutionEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*awx.AWX)
	awxService := client.JobTemplateService
	jobTemplateID := d.Get("job_template_id").(int)
	_, err := awxService.GetJobTemplateByID(jobTemplateID, make(map[string]string))

	if err != nil {
		return buildDiagNotFoundFail("job template", jobTemplateID, err)
	}

	return diags
}

func resourceJobTemplateExecutionEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceJobTemplateExecutionEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}
