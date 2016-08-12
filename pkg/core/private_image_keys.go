package core

import "github.com/supergiant/supergiant/pkg/models"

var defaultRegistryHost = "index.docker.io"

type PrivateImageKeys struct {
	Collection
}

func (c *PrivateImageKeys) Create(m *models.PrivateImageKey) error {
	if m.Host == "" {
		m.Host = defaultRegistryHost
	}
	m.MakeKey()
	return c.Collection.Create(m)
}

// // A nil, nil return would mean the image is publicly accessible.
// func (c *PrivateImageKeys) findByImageRefIfAuthRequired(imageRef string) (*models.PrivateImageKey, error) {
// 	ref, err := reference.ParseNamed(imageRef)
// 	if err != nil {
// 		panic(err)
// 	}
// 	host, repo := reference.SplitHostname(ref)
//
// 	if host == "" {
// 		host = defaultRegistryHost
// 	}
//
// 	// NOTE The one place I see problems here, which may not exist in reality:
// 	// insecure (http://), publicly-accessible images.
// 	pubReg := registry.NewClient()
// 	pubReg.BaseURL, _ = url.Parse("https://" + host + "/v1/")
//
// 	// Is it publicly-accessible?
// 	pubToken, err := pubReg.Hub.GetReadToken(repo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if _, err = pubReg.Repository.ListTags(repo, pubToken); err == nil {
// 		// It is publicly-accessible.
// 		return nil, nil
// 	} else if !strings.Contains(err.Error(), "UNAUTHORIZED") {
// 		return nil, err
// 	}
//
// 	// We will now attempt to find a matching PrivateImageKey
//
// 	// NOTE I'm not certain if DockerHub is the only that requires token auth...
// 	requiresTokenAuth := host == "index.docker.io"
//
// 	var keys []*models.PrivateImageKey
// 	if err := c.core.DB.Find(&keys, "host = ?", host); err != nil {
// 		return nil, err
// 	}
//
// 	for _, key := range keys {
// 		reg := registry.NewClient()
// 		reg.BaseURL, _ = url.Parse(key.RegistryURL())
//
// 		basicAuth := &registry.BasicAuth{
// 			Username: key.Username,
// 			Password: key.Password,
// 		}
//
// 		var auth registry.Authenticator
// 		if requiresTokenAuth {
// 			auth, err = reg.Hub.GetReadTokenWithAuth(repo, basicAuth)
// 			if err != nil {
// 				Log.Warnf("Problem authorizing PrivateImageKey %d: %s", *key.ID, err.Error())
// 				continue
// 			}
// 		} else {
// 			auth = basicAuth
// 		}
//
// 		if _, err = reg.Repository.ListTags(repo, auth); err == nil {
// 			// We've found a matching PrivateImageKey for private access
// 			return key, nil
// 		} else if !strings.Contains(err.Error(), "UNAUTHORIZED") {
// 			return nil, err
// 		}
// 		// else, If it does contain "UNAUTHORIZED", this isn't a match
// 	}
//
// 	return nil, fmt.Errorf("Cannot access image repository %s on %s", repo, host)
// }
