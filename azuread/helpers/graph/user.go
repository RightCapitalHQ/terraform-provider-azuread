package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"log"
	"net/http"
	"regexp"
)

func UserGetByObjectId(client *graphrbac.UsersClient, ctx context.Context, objectId string) (*graphrbac.User, error) {
	filter := fmt.Sprintf("objectId eq '%s'", objectId)
	resp, err := client.ListComplete(ctx, filter, "")
	if err != nil {
		return nil, fmt.Errorf("Error listing Azure AD Users for filter %q: %+v", filter, err)
	}

	values := resp.Response().Value
	if values == nil {
		return nil, fmt.Errorf("nil values for AD Users matching %q", filter)
	}
	if len(*values) == 0 {
		return nil, fmt.Errorf("Found no AD Users matching %q", filter)
	}
	if len(*values) > 2 {
		return nil, fmt.Errorf("Found multiple AD Users matching %q", filter)
	}

	user := (*values)[0]
	if user.DisplayName == nil {
		return nil, fmt.Errorf("nil DisplayName for AD Users matching %q", filter)
	}
	if *user.ObjectID != objectId {
		return nil, fmt.Errorf("objectID for AD Users matching %q does is does not match(%q!=%q)", filter, *user.ObjectID, objectId)
	}

	return &user, nil
}

func UserGetByMailNickname(client *graphrbac.UsersClient, ctx context.Context, mailNickname string) (*graphrbac.User, error) {
	filter := fmt.Sprintf("mailNickname eq '%s'", mailNickname)
	resp, err := client.ListComplete(ctx, filter, "")
	if err != nil {
		return nil, fmt.Errorf("Error listing Azure AD Users for filter %q: %+v", filter, err)
	}

	values := resp.Response().Value
	if values == nil {
		return nil, fmt.Errorf("nil values for AD Users matching %q", filter)
	}
	if len(*values) == 0 {
		return nil, fmt.Errorf("Found no AD Users matching %q", filter)
	}
	if len(*values) > 2 {
		return nil, fmt.Errorf("Found multiple AD Users matching %q", filter)
	}

	user := (*values)[0]
	if user.DisplayName == nil {
		return nil, fmt.Errorf("nil DisplayName for AD Users matching %q", filter)
	}

	return &user, nil
}

type userManager struct {
	Url string `json:"url"`
}

func UserGetManager(client *graphrbac.UsersClient, ctx context.Context, objectID string) (managerObjectId string, err error) {
	req, err := userGetManagerPreparer(client, ctx, objectID)

	if err != nil {
		return "", err
	}

	resp, err := client.GetSender(req)

	if err != nil {
		return "", err
	}

	manager := userManager{}
	err = json.NewDecoder(resp.Body).Decode(&manager)

	if objectID == "a5a66956-d3dc-4e5c-a31e-9039eba62e34" {
		log.Println("YYYYYY")
		log.Println(manager.Url)
	}

	if err != nil {
		return "", err
	}

	uuidRegex := regexp.MustCompile(`\b[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}\b`)
	managerObjectId = uuidRegex.FindString(manager.Url)

	return
}

func UserUpdateManager(client *graphrbac.UsersClient, ctx context.Context, userObjectID, mangerObjectID string) error {
	req, err := userUpdateManagerPreparer(client, ctx, userObjectID, mangerObjectID)

	if err != nil {
		return err
	}

	resp, err := client.UpdateSender(req)

	if err != nil {
		return err
	}

	_, err = client.UpdateResponder(resp)

	return err
}

func userGetManagerPreparer(client *graphrbac.UsersClient, ctx context.Context, userObjectId string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"upnOrObjectId": autorest.Encode("path", userObjectId),
	}

	const APIVersion = "1.6"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/myorganization/users/{upnOrObjectId}/$links/manager", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

type test struct {
	Url string `json:"url"`
}

type managerUpdateParameters struct {
	URL *string `json:"url,omitempty"`
}

func userUpdateManagerPreparer(client *graphrbac.UsersClient, ctx context.Context, upnOrObjectID, managerObjectID string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"upnOrObjectId": autorest.Encode("path", upnOrObjectID),
		"tenantID":      autorest.Encode("path", client.TenantID),
	}

	const APIVersion = "1.6"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	if managerObjectID == "" {
		preparer := autorest.CreatePreparer(
			autorest.AsDelete(),
			autorest.WithBaseURL(client.BaseURI),
			autorest.WithPathParameters("/{tenantID}/users/{upnOrObjectId}/$links/manager", pathParameters),
			autorest.WithQueryParameters(queryParameters))
		return preparer.Prepare((&http.Request{}).WithContext(ctx))
	} else {
		manager := userManager{
			Url: fmt.Sprintf("https://graph.windows.net/%s/directoryObjects/%s", client.TenantID, managerObjectID),
		}

		preparer := autorest.CreatePreparer(
			autorest.AsContentType("application/json; charset=utf-8"),
			autorest.AsPut(),
			autorest.WithBaseURL(client.BaseURI),
			autorest.WithPathParameters("/{tenantID}/users/{upnOrObjectId}/$links/manager", pathParameters),
			autorest.WithJSON(manager),
			autorest.WithQueryParameters(queryParameters))
		return preparer.Prepare((&http.Request{}).WithContext(ctx))
	}
}
