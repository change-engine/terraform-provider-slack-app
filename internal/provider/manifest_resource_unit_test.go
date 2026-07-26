package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func testCredentials() Credentials {
	return Credentials{
		ClientId:          types.StringValue("client-id"),
		ClientSecret:      types.StringValue("client-secret"),
		VerificationToken: types.StringValue("verification-token"),
		SigningSecret:     types.StringValue("signing-secret"),
	}
}

func testCredentialObject() types.Object {
	credentials := testCredentials()
	model := manifestResourceModel{ExportCredentials: types.BoolValue(true)}
	setCredentialOutputs(&model, credentials, "https://example.com/authorize")
	return model.Credentials
}

func assertCredentialOutputsNull(t *testing.T, model manifestResourceModel) {
	t.Helper()

	if !model.Credentials.IsNull() {
		t.Fatal("expected credentials to be null")
	}
	if !model.OAuthAuthorizeUrl.IsNull() {
		t.Fatal("expected OAuth authorization URL to be null")
	}
}

func TestSetCredentialOutputs(t *testing.T) {
	t.Run("exports credentials by default", func(t *testing.T) {
		model := manifestResourceModel{ExportCredentials: types.BoolNull()}

		setCredentialOutputs(&model, testCredentials(), "https://example.com/authorize")

		if model.Credentials.IsNull() {
			t.Fatal("expected credentials to be exported")
		}
		if got := model.Credentials.Attributes()["client_secret"].(types.String).ValueString(); got != "client-secret" {
			t.Fatalf("expected client secret to be exported, got %q", got)
		}
		if got := model.OAuthAuthorizeUrl.ValueString(); got != "https://example.com/authorize" {
			t.Fatalf("expected OAuth authorization URL to be exported, got %q", got)
		}
	})

	t.Run("does not export credentials when disabled", func(t *testing.T) {
		model := manifestResourceModel{ExportCredentials: types.BoolValue(false)}

		setCredentialOutputs(&model, testCredentials(), "https://example.com/authorize")

		assertCredentialOutputsNull(t, model)
	})
}

func TestManifestResourceExportCredentialsSchema(t *testing.T) {
	manifest := &manifestResource{}
	response := resource.SchemaResponse{}

	manifest.Schema(context.Background(), resource.SchemaRequest{}, &response)

	attribute, ok := response.Schema.Attributes["export_credentials"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("expected export_credentials boolean attribute")
	}
	if !attribute.Optional {
		t.Fatal("expected export_credentials to be optional")
	}
	if attribute.Default == nil {
		t.Fatal("expected export_credentials to default to true")
	}
}

func TestNormalizeCredentialOutputs(t *testing.T) {
	tests := map[string]struct {
		exportCredentials types.Bool
		credentials       types.Object
		oauthAuthorizeURL types.String
		wantNull          bool
	}{
		"safe update scrubs existing state": {
			exportCredentials: types.BoolValue(false),
			credentials:       testCredentialObject(),
			oauthAuthorizeURL: types.StringValue("https://example.com/authorize"),
			wantNull:          true,
		},
		"safe read scrubs existing state": {
			exportCredentials: types.BoolValue(false),
			credentials:       testCredentialObject(),
			oauthAuthorizeURL: types.StringValue("https://example.com/authorize"),
			wantNull:          true,
		},
		"safe import keeps outputs null": {
			exportCredentials: types.BoolValue(false),
			credentials:       types.ObjectUnknown(credentialAttributeTypes()),
			oauthAuthorizeURL: types.StringUnknown(),
			wantNull:          true,
		},
		"default update retains known outputs": {
			exportCredentials: types.BoolValue(true),
			credentials:       testCredentialObject(),
			oauthAuthorizeURL: types.StringValue("https://example.com/authorize"),
			wantNull:          false,
		},
		"default import clears unknown outputs": {
			exportCredentials: types.BoolValue(true),
			credentials:       types.ObjectUnknown(credentialAttributeTypes()),
			oauthAuthorizeURL: types.StringUnknown(),
			wantNull:          true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			model := manifestResourceModel{
				ExportCredentials: test.exportCredentials,
				Credentials:       test.credentials,
				OAuthAuthorizeUrl: test.oauthAuthorizeURL,
			}

			normalizeCredentialOutputs(&model)

			if test.wantNull {
				assertCredentialOutputsNull(t, model)
				return
			}
			if model.Credentials.IsNull() || model.OAuthAuthorizeUrl.IsNull() {
				t.Fatal("expected credential outputs to be retained")
			}
		})
	}
}
