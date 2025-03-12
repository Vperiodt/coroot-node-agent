package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"inet.af/netaddr"
)

func TestConnectionFilter(t *testing.T) {
	f := connectionFilter{whitelist: map[string]netaddr.IPPrefix{}}
	assert.False(t, f.ShouldBeSkipped(netaddr.MustParseIP("127.0.0.1"), netaddr.MustParseIP("127.0.0.1")))
	assert.False(t, f.ShouldBeSkipped(netaddr.MustParseIP("192.168.1.1"), netaddr.MustParseIP("127.0.0.1")))

	assert.True(t, f.ShouldBeSkipped(netaddr.MustParseIP("1.1.1.1"), netaddr.MustParseIP("2.2.2.2")))
	assert.False(t, f.ShouldBeSkipped(netaddr.MustParseIP("1.1.1.1"), netaddr.MustParseIP("192.168.1.1")))
	// because the actual dest is allowed, the dest is added to whitelist
	assert.False(t, f.ShouldBeSkipped(netaddr.MustParseIP("1.1.1.1"), netaddr.MustParseIP("2.2.2.2")))

	assert.True(t, f.ShouldBeSkipped(netaddr.MustParseIP("2.2.2.2"), netaddr.MustParseIP("2.2.2.2")))
	f.WhitelistPrefix(netaddr.MustParseIPPrefix("2.2.2.0/24"))
	assert.False(t, f.ShouldBeSkipped(netaddr.MustParseIP("2.2.2.2"), netaddr.MustParseIP("2.2.2.2")))

	assert.True(t, f.ShouldBeSkipped(netaddr.MustParseIP("3.3.3.3"), netaddr.MustParseIP("3.3.3.3")))
	f.WhitelistPrefix(netaddr.MustParseIPPrefix("4.4.4.4/32"))
	assert.False(t, f.ShouldBeSkipped(netaddr.MustParseIP("3.3.3.3"), netaddr.MustParseIP("4.4.4.4")))
}

func TestDestinationKey(t *testing.T) {
	d := netaddr.IPPortFrom(netaddr.MustParseIP("1.1.1.1"), 443)
	ad := netaddr.IPPortFrom(netaddr.MustParseIP("2.2.2.2"), 443)

	assert.Equal(t, "1.1.1.1:443 (2.2.2.2:443)", NewDestinationKey(d, ad, nil).String())

	assert.Equal(t,
		"aa.bb.s3.amazonaws.com:443 ()",
		NewDestinationKey(d, ad, &Domain{FQDN: "aa.bb.s3.amazonaws.com", SpecifyIP: false}).String(),
	)
	assert.Equal(t,
		"1.1.1.1:443 (2.2.2.2:443)",
		NewDestinationKey(d, ad, &Domain{FQDN: "aa.bb.s3.amazonaws.com", SpecifyIP: true}).String(),
	)
}

func TestDomain(t *testing.T) {
	assert.Equal(t, "Domain(fqdn,true)", NewDomain("fqdn", []netaddr.IP{netaddr.MustParseIP("127.0.0.1")}).String())
	assert.Equal(t, "Domain(fqdn,true)", NewDomain("fqdn", []netaddr.IP{netaddr.MustParseIP("192.168.1.1")}).String())
	assert.Equal(t, "Domain(fqdn,true)", NewDomain("fqdn", []netaddr.IP{
		netaddr.MustParseIP("1.1.1.1"),
		netaddr.MustParseIP("192.168.1.1"),
	}).String())
	assert.Equal(t, "Domain(fqdn,true)", NewDomain("fqdn", []netaddr.IP{
		netaddr.MustParseIP("1.1.1.1"),
	}).String())
	assert.Equal(t, "Domain(fqdn,false)", NewDomain("fqdn", []netaddr.IP{
		netaddr.MustParseIP("1.1.1.1"),
		netaddr.MustParseIP("1.1.1.2"),
	}).String())

}

func TestNormalizeFQDN(t *testing.T) {
	assert.Equal(t, "IP.in-addr.arpa", NormalizeFQDN("4.3.2.1.in-addr.arpa", "TypePTR"))
	assert.Equal(t, "coroot.com", NormalizeFQDN("coroot.com", "TypeA"))
	assert.Equal(t, "IP.ec2.internal", NormalizeFQDN("ip-172-1-2-3.ec2.internal", "TypeA"))
	assert.Equal(t, "IP.ec2", NormalizeFQDN("ip-172-1-2-3.ec2", "TypeA"))

	assert.Equal(t, "example.com", NormalizeFQDN("example.com", "TypeA"))
	assert.Equal(t, "example.com.search_path_suffix", NormalizeFQDN("example.com.cluster.local", "TypeA"))
	assert.Equal(t, "example.com.search_path_suffix", NormalizeFQDN("example.com.svc.cluster.local", "TypeA"))
	assert.Equal(t, "example.com.search_path_suffix", NormalizeFQDN("example.com.svc.default.cluster.local", "TypeA"))

	assert.Equal(t, "example.net.search_path_suffix", NormalizeFQDN("example.net.svc.default.cluster.local", "TypeA"))
	assert.Equal(t, "example.org.search_path_suffix", NormalizeFQDN("example.org.svc.default.cluster.local", "TypeA"))
	assert.Equal(t, "example.io.search_path_suffix", NormalizeFQDN("example.io.svc.default.cluster.local", "TypeA"))
}

func BenchmarkNormalizeFQDN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeFQDN("ip-172-1-2-3.ec2.internal", "TypeA")
		NormalizeFQDN("example.io.svc.default.cluster.local", "TypeA")
	}
}
