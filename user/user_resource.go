package resource_user

import (
	"log"
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	api "github.com/interdotlink/api-client-go"
)

var _ resource.Resource = (*userResource)(nil)

type userResource struct{
	client *api.APIClient
}

func NewUserResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client, _ = req.ProviderData.(*api.APIClient)
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = UserResourceSchema(ctx)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, _, err := r.client.UserAPI.UserApiUserCreateUser(ctx).UserCreate(api.UserCreate{
		FirstName: data.FirstName.String(),
		LastName: data.LastName.String(),
		Email: data.Email.String(),
	}).Execute()
    if err != nil {
        log.Fatal(err.Error())
		return
    }

	data.Id = types.Int64Value(int64(u.Id))
	//data.FirstName = types.StringValue(string(u.FirstName))
	//data.LastName = types.StringValue(string(u.LastName))
	//data.Email = types.StringValue(string(u.Email))
	data.Status = types.StringValue(string(u.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, _, err := r.client.UserAPI.UserApiUserGetUserByIdEndpoint(ctx, int32(data.Id.ValueInt64())).Execute()
    if err != nil {
        log.Fatal(err.Error())
		return
    }
	data.Id = types.Int64Value(int64(u.Id))
	data.FirstName = types.StringValue(string(u.FirstName))
	data.LastName = types.StringValue(string(u.LastName))
	data.Email = types.StringValue(string(u.Email))
	data.Status = types.StringValue(string(u.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserModel
	var dataState UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &dataState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = dataState.Id
	data.Status = dataState.Status

	u, _, err := r.client.UserAPI.UserApiUserUpdateUser(ctx, int32(data.Id.ValueInt64())).UserCreate(api.UserCreate{
		FirstName: data.FirstName.ValueString(),
		LastName: data.LastName.ValueString(),
		Email: data.Email.ValueString(),
	}).Execute()
    if err != nil {
		return
    }

	data.Id = types.Int64Value(int64(u.Id))
	//data.FirstName = types.StringValue(string(u.FirstName))
	//data.LastName = types.StringValue(string(u.LastName))
	//data.Email = types.StringValue(string(u.Email))
	data.Status = types.StringValue(string(u.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, _, err := r.client.UserAPI.UserApiUserDeleteUser(ctx, int32(data.Id.ValueInt64())).Execute()
    if err != nil {
        log.Fatal(err.Error())
		return
    }
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
        log.Fatal(err.Error())
		return
	}
	u, _, err := r.client.UserAPI.UserApiUserGetUserByIdEndpoint(ctx, int32(id)).Execute()
    if err != nil {
        log.Fatal(err.Error())
		return
    }
	var data UserModel
	data.Id = types.Int64Value(int64(u.Id))
	data.FirstName = types.StringValue(string(u.FirstName))
	data.LastName = types.StringValue(string(u.LastName))
	data.Email = types.StringValue(string(u.Email))
	data.Status = types.StringValue(string(u.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
