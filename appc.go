package gaeappc

import (
    "fmt"
    "net/url"

    "appengine"
)

type Asset struct {
    AssetID string
    AssetCreatedDate string
    CSpaceID string
}

type KeyPair struct {
    Value string
    Type interface{}
}

type Client struct {
    rest RestClient
}

func NewClient(ctx appengine.Context, apiKey string) Client {
    return Client {
        rest: restClient(ctx, apiKey),
    }
}

func (client *Client) GetContainer(containerId string, obj interface{}) error {
    return client.rest.Fetch(appcraftedURL(containerId), obj)
}

func appcraftedURL(containerId string) string {
    return fmt.Sprintf("https://api.appcrafted.com/v0/assets/%s/all", url.QueryEscape(containerId))
}

func restClient(ctx appengine.Context, apiKey string) RestClient {
    return RestClient {
        Context: ctx,
        User: apiKey,
        AllowCached: true,
        TtlSecs: 600,
    }
}