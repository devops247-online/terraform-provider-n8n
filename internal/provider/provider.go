package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/devops247-online/terraform-provider-n8n/internal/client"
)

// Ensure N8nProvider satisfies various provider interfaces.
var _ provider.Provider = &N8nProvider{}
var _ provider.ProviderWithFunctions = &N8nProvider{}

// N8nProvider defines the provider implementation.
type N8nProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// N8nProviderModel describes the provider data model.
type N8nProviderModel struct {
	BaseURL            types.String `tfsdk:"base_url"`
	APIKey             types.String `tfsdk:"api_key"`
	Email              types.String `tfsdk:"email"`
	Password           types.String `tfsdk:"password"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

func (p *N8nProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "n8n"
	resp.Version = p.version
}

func (p *N8nProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The n8n provider allows you to manage n8n workflows, credentials, and other resources " +
			"using Infrastructure as Code.\n\n" +
			"n8n is a free and source-available workflow automation tool that lets you connect anything to " +
			"everything via its open, fair-code model.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL of your n8n instance. Can be set via the " +
					"`N8N_BASE_URL` environment variable.",
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for authentication with n8n. Can be set via the " +
					"`N8N_API_KEY` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email for basic authentication with n8n. Can be set via the " +
					"`N8N_EMAIL` environment variable. Alternative to api_key.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for basic authentication with n8n. Can be set via the " +
					"`N8N_PASSWORD` environment variable. Alternative to api_key.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. Can be set via the " +
					"`N8N_INSECURE_SKIP_VERIFY` environment variable. Defaults to false.",
				Optional: true,
			},
		},
	}
}

func (p *N8nProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data N8nProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values
	baseURL := os.Getenv("N8N_BASE_URL")
	apiKey := os.Getenv("N8N_API_KEY")
	email := os.Getenv("N8N_EMAIL")
	password := os.Getenv("N8N_PASSWORD")
	insecureSkipVerify := os.Getenv("N8N_INSECURE_SKIP_VERIFY") == "true"

	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}

	if !data.Email.IsNull() {
		email = data.Email.ValueString()
	}

	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	if !data.InsecureSkipVerify.IsNull() {
		insecureSkipVerify = data.InsecureSkipVerify.ValueBool()
	}

	// If practitioner-provided configuration is missing, add errors.
	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing n8n Base URL",
			"The provider cannot create the n8n API client as there is a missing or empty value for the n8n base URL. "+
				"Set the base_url attribute in the provider configuration or use the N8N_BASE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// Check for session-based authentication from CI environment
	useSessionAuth := os.Getenv("N8N_USE_SESSION_AUTH") == "true"
	cookieFile := os.Getenv("N8N_COOKIE_FILE")

	// Create n8n client with appropriate authentication method
	var authMethod client.AuthMethod

	if useSessionAuth && cookieFile != "" {
		// Use session-based authentication for CI environments
		authMethod = &client.SessionAuth{
			CookieFile: cookieFile,
		}
	} else if apiKey != "" {
		authMethod = &client.APIKeyAuth{APIKey: apiKey}
	} else if email != "" && password != "" {
		authMethod = &client.BasicAuth{Email: email, Password: password}
	} else {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing n8n Authentication",
			"The provider cannot create the n8n API client as there is missing authentication information. "+
				"Either set the api_key attribute in the provider configuration or use the N8N_API_KEY environment variable, "+
				"or provide both email and password for basic authentication via the N8N_EMAIL and N8N_PASSWORD environment variables.",
		)
		return
	}

	clientConfig := &client.Config{
		BaseURL:            baseURL,
		Auth:               authMethod,
		InsecureSkipVerify: insecureSkipVerify,
	}

	n8nClient, err := client.NewClient(clientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create n8n API Client",
			"An unexpected error occurred when creating the n8n API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"n8n Client Error: "+err.Error(),
		)
		return
	}

	// Make the n8n client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = n8nClient
	resp.ResourceData = n8nClient
}

func (p *N8nProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWorkflowResource,
		NewCredentialResource,
		NewUserResource,
		NewProjectResource,
		NewProjectUserResource,
		NewLDAPConfigResource,
	}
}

func (p *N8nProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
	}
}

func (p *N8nProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// Functions will be added here if needed
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &N8nProvider{
			version: version,
		}
	}
}
