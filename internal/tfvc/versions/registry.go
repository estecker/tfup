package versions

import (
	"fmt"
	"net/url"
	"sort"

	goversion "github.com/hashicorp/go-version"
	"github.com/estecker/tfup/internal/tfvc/registry"
)

func Registry(client registry.Client, hostname, namespace, name, provider string) ([]*goversion.Version, error) {
	baseURL, err := client.DiscoverModules(hostname)
	if err != nil {
		return nil, fmt.Errorf("discover registry at %q: %w", hostname, err)
	}
	baseURLStruct, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse module registry url %q: %w", baseURL, err)
	}
	if baseURLStruct.Scheme == "" {
		baseURLStruct.Scheme = "https"
	}
	if baseURLStruct.Host == "" {
		baseURLStruct.Host = hostname
	}
	baseURL = baseURLStruct.String()
	versions, err := client.ListModuleVersions(baseURL, namespace, name, provider)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	return parseVersions(versions), nil
}

func RegistryProvider(client registry.Client, namespace, name string) ([]*goversion.Version, error) {
	baseURL, err := client.DiscoverProviders(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("discover registry for providers: %w", err)
	}
	versions, err := client.ListProviderVersions(baseURL)
	if err != nil {
		return nil, fmt.Errorf("list provider versions: %w", err)
	}
	return parseVersions(versions), nil
}

// parseVersions parses and sorts version strings, skipping invalid ones
// (nil entries would panic during sort/compare).
func parseVersions(versionStrings []string) []*goversion.Version {
	out := make([]*goversion.Version, 0, len(versionStrings))
	for _, s := range versionStrings {
		v, err := goversion.NewVersion(s)
		if err != nil {
			continue
		}
		out = append(out, v)
	}
	sort.Sort(goversion.Collection(out))
	return out
}
