package firewall

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// WebsiteCategories maps category names to the domains that belong to each category.
// These are used for both nftables IP-set blocking (gateway) and hosts-file blocking (agent).
var WebsiteCategories = map[string][]string{
	"social_media": {
		"facebook.com", "www.facebook.com",
		"x.com", "twitter.com", "www.twitter.com",
		"instagram.com", "www.instagram.com",
		"tiktok.com", "www.tiktok.com",
		"snapchat.com", "www.snapchat.com",
		"linkedin.com", "www.linkedin.com",
		"reddit.com", "www.reddit.com",
		"pinterest.com", "www.pinterest.com",
		"tumblr.com", "www.tumblr.com",
		"discord.com", "www.discord.com",
		"telegram.org", "web.telegram.org",
	},
	"video_streaming": {
		"youtube.com", "www.youtube.com", "youtu.be",
		"netflix.com", "www.netflix.com",
		"primevideo.com", "www.primevideo.com",
		"hotstar.com", "www.hotstar.com",
		"jiocinema.com", "www.jiocinema.com",
		"hulu.com", "www.hulu.com",
		"disneyplus.com", "www.disneyplus.com",
		"twitch.tv", "www.twitch.tv",
		"vimeo.com", "www.vimeo.com",
		"dailymotion.com", "www.dailymotion.com",
		"zee5.com", "www.zee5.com",
		"sonyliv.com", "www.sonyliv.com",
		"mxplayer.in", "www.mxplayer.in",
	},
	"shopping": {
		"amazon.com", "www.amazon.com",
		"amazon.in", "www.amazon.in",
		"flipkart.com", "www.flipkart.com",
		"myntra.com", "www.myntra.com",
		"ebay.com", "www.ebay.com",
		"alibaba.com", "www.alibaba.com",
		"aliexpress.com", "www.aliexpress.com",
		"snapdeal.com", "www.snapdeal.com",
		"meesho.com", "www.meesho.com",
		"ajio.com", "www.ajio.com",
		"nykaa.com", "www.nykaa.com",
	},
	"entertainment": {
		"spotify.com", "open.spotify.com",
		"soundcloud.com", "www.soundcloud.com",
		"crunchyroll.com", "www.crunchyroll.com",
		"gaana.com", "www.gaana.com",
		"jiosaavn.com", "www.jiosaavn.com",
		"wynk.in", "www.wynk.in",
	},
}

// KnownCategories returns all valid category names.
func KnownCategories() []string {
	cats := make([]string, 0, len(WebsiteCategories))
	for k := range WebsiteCategories {
		cats = append(cats, k)
	}
	return cats
}

// resolveDomainsToIPs performs DNS lookups for each domain and returns a deduplicated
// list of IPv4 addresses. Failures per domain are silently skipped so a single
// unresolvable domain does not abort the whole category sync.
func resolveDomainsToIPs(ctx context.Context, domains []string) []string {
	seen := make(map[string]struct{})
	var ips []string
	for _, domain := range domains {
		addrs, err := net.DefaultResolver.LookupHost(ctx, domain)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := net.ParseIP(addr)
			if ip == nil || ip.To4() == nil {
				// skip IPv6 for now; nftables ip table only handles IPv4
				continue
			}
			if _, exists := seen[addr]; !exists {
				seen[addr] = struct{}{}
				ips = append(ips, addr)
			}
		}
	}
	return ips
}

// nftSetName returns a safe nftables set name derived from a category and policy ID.
func nftSetName(category string, policyID int64) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, category)
	return fmt.Sprintf("blk_%s_%d", safe, policyID)
}
