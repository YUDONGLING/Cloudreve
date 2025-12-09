package driver

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
)

func ApplyProxy(srcUrl *url.URL, proxyBase string, pathReplacements []types.PathReplacement) (*url.URL, error) {
	// For custom proxy, generate a new proxyed URL:
	// [Proxy Scheme][Proxy Host][Proxy Port][ProxyPath + OriginSrcPath][OriginSrcQuery + ProxyQuery]
	proxy, err := url.Parse(proxyBase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}
	if proxy.Path != "" && proxy.Path != "/" {
		proxy.Path = path.Join(proxy.Path, strings.TrimPrefix(srcUrl.Path, "/"))
	} else {
		proxy.RawPath = srcUrl.RawPath
		proxy.Path = srcUrl.Path
	}

	for _, r := range pathReplacements {
		proxy.Path = util.Replace(map[string]string{r.From: r.To}, proxy.Path)
	}

	q := proxy.Query()
	if len(q) == 0 {
		proxy.RawQuery = srcUrl.RawQuery
	} else {
		srcQ := srcUrl.Query()
		for k, _ := range srcQ {
			q.Set(k, srcQ.Get(k))
		}
		proxy.RawQuery = q.Encode()
	}

	return proxy, nil
}

func ApplyProxyIfNeeded(policy *ent.StoragePolicy, srcUrl *url.URL) (*url.URL, error) {
	if policy.Settings.CustomProxy {
		return ApplyProxy(srcUrl, policy.Settings.ProxyServer, policy.Settings.PathReplacements)
	}
	return srcUrl, nil
}

func ApplyUploadProxyIfNeeded(policy *ent.StoragePolicy, srcUrl *url.URL) (*url.URL, error) {
	if policy.Settings.UploadCustomProxy {
		return ApplyProxy(srcUrl, policy.Settings.UploadProxyServer, policy.Settings.UploadPathReplacements)
	}
	return srcUrl, nil
}
