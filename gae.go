package gaeappc

import (
    "time"
    "net/http"
    "io/ioutil"
    "encoding/json"

    "appengine"
    "appengine/memcache"
    "appengine/urlfetch"
)

type RestClient struct {
    Context appengine.Context
    User string
    Password string
    TtlSecs int
    AllowCached bool
}

func (client *RestClient) Fetch(url string, obj interface{}) error {
    if client.AllowCached {
        if item, err := memcache.Get(client.Context, url); err == memcache.ErrCacheMiss {
            // not in cache
        } else if err != nil {
            client.Context.Errorf("Error fetching [%s] from cache: %v", url, err)
        } else {
            jsonerr := json.Unmarshal(item.Value, obj)

            if jsonerr != nil {
                client.Context.Errorf("Error unmarshaling [%s]: %s", item.Value, jsonerr.Error())
                return jsonerr
            } else {
                return nil
            }
        }
    }

    data, err := client.fetch(url, obj)
    if err != nil {
        return err
    } else if client.TtlSecs > 0 {
        client.cache(url, data)
    }

    return nil
}

func (client *RestClient) cache(key string, value []byte) {
    item := &memcache.Item {
        Key: key,
        Value: value,
        Expiration: time.Duration(client.TtlSecs) * time.Second,
    }

    if err := memcache.Add(client.Context, item); err == memcache.ErrNotStored {
        if err2 := memcache.Set(client.Context, item); err2 != nil {
            client.Context.Errorf("Error updating [%s] to [%s]: %v", key, value, err2)
        }
    } else if err != nil {
        client.Context.Errorf("Error setting [%s] to [%s]: %v", key, value, err)
    }
}

func (client *RestClient) fetch(url string, obj interface{}) ([]byte, error) {
    fetcher := urlfetch.Client(client.Context)
    req, _ := http.NewRequest("GET", url, nil)

    if "" != client.User || "" != client.Password {
        req.SetBasicAuth(client.User, client.Password)
    }

    resp, curlerr := fetcher.Do(req)
    defer resp.Body.Close()

    if curlerr != nil {
        client.Context.Errorf("Error fetching [%s]: %s", url, curlerr.Error())
        return nil, curlerr
    }

    data, ioerr := ioutil.ReadAll(resp.Body)
    if ioerr != nil {
        client.Context.Errorf("Error stringing [%s]: %s", resp.Body, ioerr.Error())
        return nil, ioerr
    }

    client.Context.Debugf("Got response from %s: %s", url, data)

    jsonerr := json.Unmarshal(data, obj)
    if jsonerr != nil {
        client.Context.Errorf("Error jsoning [%s]: %s", data, jsonerr.Error())
        return nil, jsonerr
    }

    return data, nil
}