package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ tfsdk.ResourceType = manifestResourceType{}
var _ tfsdk.Resource = manifestResource{}

type manifestResourceType struct{}

func (t manifestResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Slack App Manifest resource",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed:            true,
				MarkdownDescription: "The ID of the app.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"manifest": {
				MarkdownDescription: "A JSON app manifest encoded as a string.",
				Required:            true,
				Type:                types.StringType,
			},
			"credentials": {
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"client_id": {
						Computed:            true,
						MarkdownDescription: "Send with `client_secret` when making your oauth.v2.access request.",
						Type:                types.StringType,
					},
					"client_secret": {
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "Send with `client_id` when making your oauth.v2.access request.",
						Type:                types.StringType,
					},
					"verification_token": {
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "used to verify that requests come from Slack.",
						Type:                types.StringType,
						DeprecationMessage:  "We strongly recommend using the, more secure, `signing_secret` instead.",
					},
					"signing_secret": {
						Computed:            true,
						Sensitive:           true,
						MarkdownDescription: "Slack signs the requests we send you using this secret. Confirm that each request comes from Slack by verifying its unique signature.",
						Type:                types.StringType,
					},
				}),
			},
			"oauth_authorize_url": {
				Computed:            true,
				MarkdownDescription: "Full URL for athorization.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t manifestResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return manifestResource{
		provider: provider,
	}, diags
}

type Credentials struct {
	ClientId          types.String `tfsdk:"client_id"`
	ClientSecret      types.String `tfsdk:"client_secret"`
	VerificationToken types.String `tfsdk:"verification_token"`
	SigningSecret     types.String `tfsdk:"signing_secret"`
}

type manifestResourceData struct {
	Manifest          types.String `tfsdk:"manifest"`
	ID                types.String `tfsdk:"id"`
	Credentials       *Credentials `tfsdk:"credentials"`
	OAuthAuthorizeUrl types.String `tfsdk:"oauth_authorize_url"`
}

type createManifestReqest struct {
	AppID    string `json:"app_id,omit_empty"`
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

type manifestResource struct {
	provider provider
}

func (r manifestResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data manifestResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestReqest{
		Manifest: data.Manifest.Value,
	})
	var resultJson createManifestResponse
	err := r.provider.client.Request(ctx, "apps.manifest.create", request, &resultJson)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create manifest, got error: %s", err))
		return
	}

	data.ID = types.String{Value: resultJson.AppID}
	data.Credentials = &Credentials{
		ClientId:          types.String{Value: resultJson.Credentials.ClientId},
		ClientSecret:      types.String{Value: resultJson.Credentials.ClientSecret},
		SigningSecret:     types.String{Value: resultJson.Credentials.SigningSecret},
		VerificationToken: types.String{Value: resultJson.Credentials.VerificationToken},
	}
	data.OAuthAuthorizeUrl = types.String{Value: resultJson.OAuthAuthorizeUrl}

	tflog.Trace(ctx, "created a manifest")
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r manifestResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data manifestResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestReqest{
		AppID: data.ID.Value,
	})
	var resultJson exportManifestResponse
	err := r.provider.client.Request(ctx, "apps.manifest.export", request, &resultJson)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create manifest, got error: %s", err))
		return
	}

	norm, _ := json.Marshal(resultJson.Manifest)
	data.Manifest = types.String{Value: string(norm)}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r manifestResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data manifestResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestReqest{
		AppID:    data.ID.Value,
		Manifest: data.Manifest.Value,
	})
	err := r.provider.client.Request(ctx, "apps.manifest.update", request, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create manifest, got error: %s", err))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r manifestResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data manifestResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, _ := json.Marshal(createManifestReqest{
		AppID: data.ID.Value,
	})
	err := r.provider.client.Request(ctx, "apps.manifest.delete", request, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete manifest, got error: %s", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r manifestResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
