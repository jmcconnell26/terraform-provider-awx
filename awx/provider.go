package awx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awx "github.com/mrcrilly/goawx/client"
	"net/http"
	"os"
	"path/filepath"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hostname": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AWX_HOSTNAME", "http://localhost"),
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable SSL verification of API calls",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AWX_USERNAME", "admin"),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("AWX_PASSWORD", "password"),
			},
			"client_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client certificate to use for mTLS validation. Must be provided along with client-key and ca-cert for mTLS to be used.",
			},
			"client_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client key to use for mTLS validation. Must be provided along with client-cert and ca-cert for mTLS to be used",
			},
			"ca_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "CA certificate to use for mTLS validation. Must be provided along with client-cert and client-key for mTLS to be used",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"awx_credential_azure_key_vault":         resourceCredentialAzureKeyVault(),
			"awx_credential_google_compute_engine":   resourceCredentialGoogleComputeEngine(),
			"awx_credential_input_source":            resourceCredentialInputSource(),
			"awx_credential_machine":                 resourceCredentialMachine(),
			"awx_credential_scm":                     resourceCredentialSCM(),
			"awx_host":                               resourceHost(),
			"awx_inventory_group":                    resourceInventoryGroup(),
			"awx_inventory_source":                   resourceInventorySource(),
			"awx_inventory":                          resourceInventory(),
			"awx_job_template_credential":            resourceJobTemplateCredentials(),
			"awx_job_template":                       resourceJobTemplate(),
			"awx_organization":                       resourceOrganization(),
			"awx_project":                            resourceProject(),
			"awx_workflow_job_template_node_allways": resourceWorkflowJobTemplateNodeAllways(),
			"awx_workflow_job_template_node_failure": resourceWorkflowJobTemplateNodeFailure(),
			"awx_workflow_job_template_node_success": resourceWorkflowJobTemplateNodeSuccess(),
			"awx_workflow_job_template_node":         resourceWorkflowJobTemplateNode(),
			"awx_workflow_job_template":              resourceWorkflowJobTemplate(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"awx_credential_azure_key_vault": dataSourceCredentialAzure(),
			"awx_credential":                 dataSourceCredentialByName(),
			"awx_execution_environment":      dataSourceExecutionEnvironmentByName(),
			"awx_inventory_group":            dataSourceInventoryGroup(),
			"awx_inventory":                  dataSourceInventory(),
			"awx_job_template":               dataSourceJobTemplate(),
			"awx_organization":               dataSourceOrganization(),
			"awx_project":                    dataSourceProject(),
			"awx_workflow_job_template":      dataSourceWorkflowJobTemplate(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	hostname := d.Get("hostname").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	clientCertPEM, clientCertPEMExists := d.GetOk("client_cert")
	clientKeyPEM, clientKeyPEMExists := d.GetOk("client_key")
	caCertPEM, caCertPEMExists := d.GetOk("ca_cert")

	client := http.DefaultClient
	if d.Get("insecure").(bool) {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else if clientCertPEMExists && clientKeyPEMExists && caCertPEMExists {
		transportTlsConfig, err := generateMtlsConfig(
			clientCertPEM.(string),
			clientKeyPEM.(string),
			caCertPEM.(string))

		if err == nil {
			client.Transport = &http.Transport{
				TLSClientConfig: transportTlsConfig,
			}
		} else {
			var diags diag.Diagnostics
			return nil, append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid mTLS config",
				Detail:   fmt.Sprintf("The provided mTLS config is invalid: %s", err),
			})
		}
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c, err := awx.NewAWX(hostname, username, password, client)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create AWX client",
			Detail:   "Unable to auth user against AWX API: check the hostname, username and password",
		})
		return nil, diags
	}

	return c, diags
}

// This method also writes the mTLS certificates to disk - to allow them to be used for other calls to
// AWX
func generateMtlsConfig(clientCertPEM string, clientKeyPEM string, caCertPEM string) (*tls.Config, error) {
	clientCertPEMBlock := []byte(clientCertPEM)
	clientKeyPEMBlock := []byte(clientKeyPEM)
	caCertPEMBlock := []byte(caCertPEM)

	tempCertificatePath := "/wrk/terraform-provider-awx"
	os.MkdirAll(tempCertificatePath, 0700)

	writeByteArrayToFile(clientCertPEMBlock, filepath.Join(tempCertificatePath, "tls.crt"))
	writeByteArrayToFile(clientKeyPEMBlock, filepath.Join(tempCertificatePath, "tls.key"))
	writeByteArrayToFile(caCertPEMBlock, filepath.Join(tempCertificatePath, "ca.crt"))

	cert, err := tls.X509KeyPair(clientCertPEMBlock, clientKeyPEMBlock)
	if err != nil {
		return nil, err
	}

	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}
	caCertPool.AppendCertsFromPEM(caCertPEMBlock)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}

func writeByteArrayToFile(byteArray []byte, filePath string) {
	t, err := os.Create(filePath)
	if err == nil {
		t.Write(byteArray)
		t.Close()
	}
}
