package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/change-engine/terraform-provider-slack-app/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &manifestResource{}
	_ resource.ResourceWithConfigure   = &manifestResource{}
	_ resource.ResourceWithImportState = &manifestResource{}
)

func NewManifestResource() resource.Resource {
	return &manifestResource{}
}

type manifestResource struct {
	client *client.SlackApp
}

func (r *manifestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manifest"
}

func (r *manifestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Slack App Manifest resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the app.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"manifest": schema.StringAttribute{
				MarkdownDescription: "A JSON app manifest encoded as a string.",
				Required:            true,
			},
			"credentials": schema.SingleNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Send with `client_secret` when making your oauth.v2.access request.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						}},
					"client_secret": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "Send with `client_id` when making your oauth.v2.access request.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						}},
					"verification_token": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "used to verify that requests come from Slack.",
						DeprecationMessage:  "We strongly recommend using the, more secure, `signing_secret` instead.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						}},
					"signing_secret": schema.StringAttribute{
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "Slack signs the requests we send you using this secret. Confirm that each request comes from Slack by verifying its unique signature.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						}},
				},
			},
			"oauth_authorize_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Full URL for authorization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *manifestResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.SlackApp)
}

type Credentials struct {
	ClientId          types.String `tfsdk:"client_id"`
	ClientSecret      types.String `tfsdk:"client_secret"`
	VerificationToken types.String `tfsdk:"verification_token"`
	SigningSecret     types.String `tfsdk:"signing_secret"`
}

type manifestResourceModel struct {
	Manifest          types.String `tfsdk:"manifest"`
	ID                types.String `tfsdk:"id"`
	Credentials       types.Object `tfsdk:"credentials"`
	OAuthAuthorizeUrl types.String `tfsdk:"oauth_authorize_url"`
}

type createManifestRequest struct {
	AppID    string `json:"app_id,omitempty"`
	Manifest string `json:"manifest"`
}

type createManifestResponse struct {
	AppID       string `json:"app_id"`
	Credentials struct {
		ClientId          string `json:"client_id"`
		ClientSecret      string `json:"client_secret"`
		VerificationToken string `json:"verification_token"`
		SigningSecret     string `json:"signing_secret"`
	} `json:"credentials"`
	OAuthAuthorizeUrl string `json:"oauth_authorize_url"`
}

type exportManifestResponse struct {
	Manifest interface{} `json:"manifest"`
}

func (r *manifestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan manifestResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestRequest{
		Manifest: plan.Manifest.ValueString(),
	})
	var resultJson createManifestResponse
	err := r.client.Request(ctx, "apps.manifest.create", request, &resultJson)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create manifest, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), resultJson.AppID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("manifest"), plan.Manifest)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("credentials"), Credentials{
		ClientId:          types.StringValue(resultJson.Credentials.ClientId),
		ClientSecret:      types.StringValue(resultJson.Credentials.ClientSecret),
		VerificationToken: types.StringValue(resultJson.Credentials.VerificationToken),
		SigningSecret:     types.StringValue(resultJson.Credentials.SigningSecret),
	})...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("oauth_authorize_url"), resultJson.OAuthAuthorizeUrl)...)
	tflog.Trace(ctx, "created a manifest")
}

func (r *manifestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state manifestResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestRequest{
		AppID: state.ID.ValueString(),
	})
	var resultJson exportManifestResponse
	err := r.client.Request(ctx, "apps.manifest.export", request, &resultJson)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create manifest, got error: %s", err))
		return
	}

	norm, _ := json.Marshal(resultJson.Manifest)
	state.Manifest = types.StringValue(string(norm))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *manifestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan manifestResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestRequest{
		AppID:    plan.ID.ValueString(),
		Manifest: plan.Manifest.ValueString(),
	})
	err := r.client.Request(ctx, "apps.manifest.update", request, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create manifest, got error: %s", err))
		return
	}

	// Slack only returns `credentials` and `oauth_authorize_url` on create, not update. If this was an imported
	// app, just mark these value as null to avoid "Error: Provider returned invalid result object after apply".
	if plan.Credentials.IsUnknown() {
		plan.Credentials = types.ObjectNull(plan.Credentials.AttributeTypes(ctx))
	}
	if plan.OAuthAuthorizeUrl.IsUnknown() {
		plan.OAuthAuthorizeUrl = types.StringNull()
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *manifestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state manifestResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestRequest{
		AppID: state.ID.ValueString(),
	})
	err := r.client.Request(ctx, "apps.manifest.delete", request, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete manifest, got error: %s", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *manifestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
