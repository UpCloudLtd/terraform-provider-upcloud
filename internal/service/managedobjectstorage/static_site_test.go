package managedobjectstorage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseErrorPages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pages   []errorPageModel
		wantErr bool
	}{
		{
			name: "status code only",
			pages: []errorPageModel{{
				StatusCode:    types.Int64Value(404),
				ErrorDocument: types.StringValue("errors/404.html"),
			}},
		},
		{
			name: "status range only",
			pages: []errorPageModel{{
				StatusRangeStart: types.Int64Value(400),
				StatusRangeEnd:   types.Int64Value(499),
				ErrorDocument:    types.StringValue("errors/4xx.html"),
			}},
		},
		{
			name: "status code and range",
			pages: []errorPageModel{{
				StatusCode:       types.Int64Value(404),
				StatusRangeStart: types.Int64Value(400),
				StatusRangeEnd:   types.Int64Value(499),
				ErrorDocument:    types.StringValue("errors/404.html"),
			}},
			wantErr: true,
		},
		{
			name: "missing range end",
			pages: []errorPageModel{{
				StatusRangeStart: types.Int64Value(400),
				ErrorDocument:    types.StringValue("errors/4xx.html"),
			}},
			wantErr: true,
		},
		{
			name: "missing matcher",
			pages: []errorPageModel{{
				ErrorDocument: types.StringValue("errors/default.html"),
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pages, diags := types.ListValueFrom(context.Background(), errorPageType(), tt.pages)
			if diags.HasError() {
				t.Fatalf("failed to create test list: %v", diags)
			}

			_, err := parseErrorPages(context.Background(), pages)
			if tt.wantErr && err == nil {
				t.Fatal("expected an error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
