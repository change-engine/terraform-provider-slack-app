package provider

import (
	"context"
	"os"

	"github.com/change-engine/terraform-provider-slack-app/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &slackAppProvider{}

func New() provider.Provider {
	return &slackAppProvider{}
}

type slackAppProvider struct{}

type slackAppProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *slackAppProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "slack-app"
}

func (p *slackAppProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This is for managing Slack App Manifests, it is no use if you are not developing an App for Slack.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "An App configuration token from https://api.slack.com/authentication/config-tokens. Can be set via the `SLACK_APP_TOKEN` environment variable.",
			},
		},
	}
}

func (p *slackAppProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config slackAppProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	if config.Token.IsNull() {
		config.Token = types.StringValue(os.Getenv("SLACK_APP_TOKEN"))
	}
	client := client.New(config.Token.ValueString())
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *slackAppProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *slackAppProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewManifestResource,
	}
}
