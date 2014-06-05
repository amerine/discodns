package main

import (
    "github.com/coreos/go-etcd/etcd"
    "github.com/miekg/dns"
    "testing"
)

var (
    client = etcd.NewClient([]string{"127.0.0.1:4001"})
    resolver = &Resolver{etcd: client}
)

func TestEtcd(t *testing.T) {
    // Enable debug logging
    log_debug = true

    if !client.SyncCluster() {
        t.Error("Failed to sync etcd cluster")
        t.Fatal()
    }
}

func TestGetFromStorageSingleKey(t *testing.T) {
    resolver.etcdPrefix = "TestGetFromStorageSingleKey/"
    client.Set("TestGetFromStorageSingleKey/net/disco/.A", "1.1.1.1", 0)

    nodes, err := resolver.GetFromStorage("net/disco/.A")
    if err != nil {
        t.Error("Error returned from etcd", err)
        t.Fatal()
    }

    if len(nodes) != 1 {
        t.Error("Number of nodes should be 1: ", len(nodes))
        t.Fatal()
    }

    node := nodes[0]
    if node.Value != "1.1.1.1" {
        t.Error("Node value should be 1.1.1.1: ", node)
        t.Fatal()
    }
}

func TestGetFromStorageNestedKeys(t *testing.T) {
    resolver.etcdPrefix = "TestGetFromStorageNestedKeys/"
    client.Set("TestGetFromStorageNestedKeys/net/disco/.A/0", "1.1.1.1", 0)
    client.Set("TestGetFromStorageNestedKeys/net/disco/.A/1", "1.1.1.2", 0)
    client.Set("TestGetFromStorageNestedKeys/net/disco/.A/2/0", "1.1.1.3", 0)

    nodes, err := resolver.GetFromStorage("net/disco/.A")
    if err != nil {
        t.Error("Error returned from etcd", err)
        t.Fatal()
    }

    if len(nodes) != 3 {
        t.Error("Number of nodes should be 3: ", len(nodes))
        t.Fatal()
    }

    var node *etcd.Node

    node = nodes[0]
    if node.Value != "1.1.1.1" {
        t.Error("Node value should be 1.1.1.1: ", node)
        t.Fatal()
    }
    node = nodes[1]
    if node.Value != "1.1.1.2" {
        t.Error("Node value should be 1.1.1.2: ", node)
        t.Fatal()
    }
    node = nodes[2]
    if node.Value != "1.1.1.3" {
        t.Error("Node value should be 1.1.1.3: ", node)
        t.Fatal()
    }
}

func TestNameToKeyConverter(t *testing.T) {
    var key string

    key = nameToKey("foo.net.", "")
    if key != "/net/foo" {
        t.Error("Expected key /net/foo")
    }

    key = nameToKey("foo.net", "")
    if key != "/net/foo" {
        t.Error("Expected key /net/foo")
    }

    key = nameToKey("foo.net.", "/.A")
    if key != "/net/foo/.A" {
        t.Error("Expected key /net/foo/.A")
    }
}

/**
 * Test that the right authority is being returned for different types of DNS
 * queries.
 */

func TestAuthorityRoot(t *testing.T) {
    resolver.etcdPrefix = "TestAuthorityRoot/"
    client.Set("TestAuthorityRoot/net/disco/.SOA", "ns1.disco.net.\\tadmin.disco.net.\\t3600\\t600\\t86400\\t10", 0)

    query := new(dns.Msg)
    query.SetQuestion("disco.net.", dns.TypeA)

    answer := resolver.Lookup(query)

    if len(answer.Answer) > 0 {
        t.Error("Expected zero answers")
        t.Fatal()
    }

    if len(answer.Ns) != 1 {
        t.Error("Expected one authority record")
        t.Fatal()
    }

    rr := answer.Ns[0].(*dns.SOA)
    header := rr.Header()

    // Verify the header is correct
    if header.Name != "disco.net." {
        t.Error("Expected record with name disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeSOA {
        t.Error("Expected record with type SOA:", header.Rrtype)
        t.Fatal()
    }

    // Verify the record itself is correct
    if rr.Ns != "ns1.disco.net." {
        t.Error("Expected NS to be ns1.disco.net.: ", rr.Ns)
        t.Fatal()
    }
    if rr.Mbox != "admin.disco.net." {
        t.Error("Expected MBOX to be admin.disco.net.: ", rr.Mbox)
        t.Fatal()
    }
    // if rr.Serial != "admin.disco.net" {
    //     t.Error("Expected MBOX to be admin.disco.net: ", rr.Mbox)
    // }
    if rr.Refresh != 3600 {
        t.Error("Expected REFRESH to be 3600: ", rr.Refresh)
        t.Fatal()
    }
    if rr.Retry != 600 {
        t.Error("Expected RETRY to be 600: ", rr.Retry)
        t.Fatal()
    }
    if rr.Expire != 86400 {
        t.Error("Expected EXPIRE to be 86400: ", rr.Expire)
        t.Fatal()
    }
    if rr.Minttl != 10 {
        t.Error("Expected MINTTL to be 10: ", rr.Minttl)
        t.Fatal()
    }
}

func TestAuthorityDomain(t *testing.T) {
    resolver.etcdPrefix = "TestAuthorityDomain/"
    client.Set("TestAuthorityDomain/net/disco/.SOA", "ns1.disco.net.\\tadmin.disco.net.\\t3600\\t600\\t86400\\t10", 0)

    query := new(dns.Msg)
    query.SetQuestion("bar.disco.net.", dns.TypeA)

    answer := resolver.Lookup(query)

    if len(answer.Answer) > 0 {
        t.Error("Expected zero answers")
        t.Fatal()
    }

    if len(answer.Ns) != 1 {
        t.Error("Expected one authority record")
        t.Fatal()
    }

    rr := answer.Ns[0].(*dns.SOA)
    header := rr.Header()

    // Verify the header is correct
    if header.Name != "disco.net." {
        t.Error("Expected record with name disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeSOA {
        t.Error("Expected record with type SOA:", header.Rrtype)
        t.Fatal()
    }

    // Verify the record itself is correct
    if rr.Ns != "ns1.disco.net." {
        t.Error("Expected NS to be ns1.disco.net.: ", rr.Ns)
        t.Fatal()
    }
    if rr.Mbox != "admin.disco.net." {
        t.Error("Expected MBOX to be admin.disco.net.: ", rr.Mbox)
        t.Fatal()
    }
    if rr.Refresh != 3600 {
        t.Error("Expected REFRESH to be 3600: ", rr.Refresh)
        t.Fatal()
    }
    if rr.Retry != 600 {
        t.Error("Expected RETRY to be 600: ", rr.Retry)
        t.Fatal()
    }
    if rr.Expire != 86400 {
        t.Error("Expected EXPIRE to be 86400: ", rr.Expire)
        t.Fatal()
    }
    if rr.Minttl != 10 {
        t.Error("Expected MINTTL to be 10: ", rr.Minttl)
        t.Fatal()
    }
}

/**
 * Test different that types of DNS queries return the correct answers
 **/

func TestAnswerQuestionA(t *testing.T) {
    resolver.etcdPrefix = "TestAnswerQuestionA/"
    client.Set("TestAnswerQuestionA/net/disco/bar/.A", "1.2.3.4", 0)
    client.Set("TestAnswerQuestionA/net/disco/.SOA", "ns1.disco.net.\\tadmin.disco.net.\\t3600\\t600\\t86400\\t10", 0)

    query := new(dns.Msg)
    query.SetQuestion("bar.disco.net.", dns.TypeA)

    answer := resolver.Lookup(query)

    if len(answer.Answer) != 1 {
        t.Error("Expected one answer, got ", len(answer.Answer))
        t.Fatal()
    }

    if len(answer.Ns) > 0 {
        t.Error("Didn't expect any authority records")
        t.Fatal()
    }

    rr := answer.Answer[0].(*dns.A)
    header := rr.Header()

    // Verify the header is correct
    if header.Name != "bar.disco.net." {
        t.Error("Expected record with name bar.disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeA {
        t.Error("Expected record with type A:", header.Rrtype)
        t.Fatal()
    }

    // Verify the record itself is correct
    if rr.A.String() != "1.2.3.4" {
        t.Error("Expected A record to be 1.2.3.4: ", rr.A)
        t.Fatal()
    }
}

func TestAnswerQuestionAAAA(t *testing.T) {
    resolver.etcdPrefix = "TestAnswerQuestionAAAA/"
    client.Set("TestAnswerQuestionAAAA/net/disco/bar/.AAAA", "::1", 0)
    client.Set("TestAnswerQuestionAAAA/net/disco/.SOA", "ns1.disco.net.\\tadmin.disco.net.\\t3600\\t600\\t86400\\t10", 0)

    query := new(dns.Msg)
    query.SetQuestion("bar.disco.net.", dns.TypeAAAA)

    answer := resolver.Lookup(query)

    if len(answer.Answer) != 1 {
        t.Error("Expected one answer, got ", len(answer.Answer))
        t.Fatal()
    }

    if len(answer.Ns) > 0 {
        t.Error("Didn't expect any authority records")
        t.Fatal()
    }

    rr := answer.Answer[0].(*dns.AAAA)
    header := rr.Header()

    // Verify the header is correct
    if header.Name != "bar.disco.net." {
        t.Error("Expected record with name bar.disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeAAAA {
        t.Error("Expected record with type AAAA:", header.Rrtype)
        t.Fatal()
    }

    // Verify the record itself is correct
    if rr.AAAA.String() != "::1" {
        t.Error("Expected AAAA record to be ::1: ", rr.AAAA)
        t.Fatal()
    }
}

func TestAnswerQuestionANY(t *testing.T) {
    resolver.etcdPrefix = "TestAnswerQuestionANY/"
    client.Set("TestAnswerQuestionANY/net/disco/bar/.TXT", "google.com.", 0)
    client.Set("TestAnswerQuestionANY/net/disco/bar/.A/0", "1.2.3.4", 0)
    client.Set("TestAnswerQuestionANY/net/disco/bar/.A/1", "2.3.4.5", 0)

    query := new(dns.Msg)
    query.SetQuestion("bar.disco.net.", dns.TypeANY)

    answer := resolver.Lookup(query)

    if len(answer.Answer) != 3 {
        t.Error("Expected one answer, got ", len(answer.Answer))
        t.Fatal()
    }

    if len(answer.Ns) > 0 {
        t.Error("Didn't expect any authority records")
        t.Fatal()
    }
}

func TestAnswerQuestionWildcardAAAANoMatch(t *testing.T) {
    resolver.etcdPrefix = "TestAnswerQuestionWildcardANoMatch/"
    client.Set("TestAnswerQuestionWildcardANoMatch/net/disco/bar/*/.AAAA", "::1", 0)

    query := new(dns.Msg)
    query.SetQuestion("bar.disco.net.", dns.TypeAAAA)

    answer := resolver.Lookup(query)

    if len(answer.Answer) > 0 {
        t.Error("Didn't expect any answers, got ", len(answer.Answer))
        t.Fatal()
    }
}

func TestAnswerQuestionWildcardAAAA(t *testing.T) {
    resolver.etcdPrefix = "TestAnswerQuestionWildcardA/"
    client.Set("TestAnswerQuestionWildcardA/net/disco/bar/*/.AAAA", "::1", 0)

    query := new(dns.Msg)
    query.SetQuestion("baz.bar.disco.net.", dns.TypeAAAA)

    answer := resolver.Lookup(query)

    if len(answer.Answer) != 1 {
        t.Error("Expected one answer, got ", len(answer.Answer))
        t.Fatal()
    }

    if len(answer.Ns) > 0 {
        t.Error("Didn't expect any authority records")
        t.Fatal()
    }

    rr := answer.Answer[0].(*dns.AAAA)
    header := rr.Header()

    // Verify the header is correct
    if header.Name != "baz.bar.disco.net." {
        t.Error("Expected record with name baz.bar.disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeAAAA {
        t.Error("Expected record with type AAAA:", header.Rrtype)
        t.Fatal()
    }

    // Verify the record itself is correct
    if rr.AAAA.String() != "::1" {
        t.Error("Expected AAAA record to be ::1: ", rr.AAAA)
        t.Fatal()
    }
}

/**
 * Test converstion of names (i.e etcd nodes) to single records of different
 * types.
 **/

func TestLookupAnswerForA(t *testing.T) {
    resolver.etcdPrefix = "TestLookupAnswerForA/"
    client.Set("TestLookupAnswerForA/net/disco/bar/.A", "1.2.3.4", 0)

    records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeA)

    if len(records) != 1 {
        t.Error("Expected one answer, got ", len(records))
        t.Fatal()
    }

    rr := records[0].(*dns.A)
    header := rr.Header()

    if header.Name != "bar.disco.net." {
        t.Error("Expected record with name bar.disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeA {
        t.Error("Expected record with type A:", header.Rrtype)
        t.Fatal()
    }
    if rr.A.String() != "1.2.3.4" {
        t.Error("Expected A record to be 1.2.3.4: ", rr.A)
        t.Fatal()
    }
}

func TestLookupAnswerForAAAA(t *testing.T) {
    resolver.etcdPrefix = "TestLookupAnswerForAAAA/"
    client.Set("TestLookupAnswerForAAAA/net/disco/bar/.AAAA", "::1", 0)

    records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeAAAA)

    if len(records) != 1 {
        t.Error("Expected one answer, got ", len(records))
        t.Fatal()
    }

    rr := records[0].(*dns.AAAA)
    header := rr.Header()

    if header.Name != "bar.disco.net." {
        t.Error("Expected record with name bar.disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeAAAA {
        t.Error("Expected record with type AAAA:", header.Rrtype)
        t.Fatal()
    }
    if rr.AAAA.String() != "::1" {
        t.Error("Expected AAAA record to be ::1: ", rr.AAAA)
        t.Fatal()
    }
}

func TestLookupAnswerForCNAME(t *testing.T) {
    resolver.etcdPrefix = "TestLookupAnswerForCNAME/"
    client.Set("TestLookupAnswerForCNAME/net/disco/bar/.CNAME", "cname.google.com.", 0)

    records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeCNAME)

    if len(records) != 1 {
        t.Error("Expected one answer, got ", len(records))
        t.Fatal()
    }

    rr := records[0].(*dns.CNAME)
    header := rr.Header()

    if header.Name != "bar.disco.net." {
        t.Error("Expected record with name bar.disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeCNAME {
        t.Error("Expected record with type CNAME:", header.Rrtype)
        t.Fatal()
    }
    if rr.Target != "cname.google.com." {
        t.Error("Expected CNAME record to be cname.google.com.: ", rr.Target)
        t.Fatal()
    }
}

func TestLookupAnswerForNS(t *testing.T) {
    resolver.etcdPrefix = "TestLookupAnswerForNS/"
    client.Set("TestLookupAnswerForNS/net/disco/bar/.NS", "dns.google.com.", 0)

    records, _ := resolver.LookupAnswersForType("bar.disco.net.", dns.TypeNS)

    if len(records) != 1 {
        t.Error("Expected one answer, got ", len(records))
        t.Fatal()
    }

    rr := records[0].(*dns.NS)
    header := rr.Header()

    if header.Name != "bar.disco.net." {
        t.Error("Expected record with name bar.disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeNS {
        t.Error("Expected record with type NS:", header.Rrtype)
        t.Fatal()
    }
    if rr.Ns != "dns.google.com." {
        t.Error("Expected NS record to be dns.google.com.: ", rr.Ns)
        t.Fatal()
    }
}

func TestLookupAnswerForSOA(t *testing.T) {
    resolver.etcdPrefix = "TestLookupAnswerForSOA/"
    client.Set("TestLookupAnswerForSOA/net/disco/.SOA", "ns1.disco.net.\\tadmin.disco.net.\\t3600\\t600\\t86400\\t10", 0)

    records, _ := resolver.LookupAnswersForType("disco.net.", dns.TypeSOA)

    if len(records) != 1 {
        t.Error("Expected one answer, got ", len(records))
        t.Fatal()
    }

    rr := records[0].(*dns.SOA)
    header := rr.Header()

    if header.Name != "disco.net." {
        t.Error("Expected record with name disco.net.: ", header.Name)
        t.Fatal()
    }
    if header.Rrtype != dns.TypeSOA {
        t.Error("Expected record with type SOA:", header.Rrtype)
        t.Fatal()
    }

    // Verify the record itself is correct
    if rr.Ns != "ns1.disco.net." {
        t.Error("Expected NS to be ns1.disco.net.: ", rr.Ns)
        t.Fatal()
    }
    if rr.Mbox != "admin.disco.net." {
        t.Error("Expected MBOX to be admin.disco.net.: ", rr.Mbox)
        t.Fatal()
    }
    if rr.Refresh != 3600 {
        t.Error("Expected REFRESH to be 3600: ", rr.Refresh)
        t.Fatal()
    }
    if rr.Retry != 600 {
        t.Error("Expected RETRY to be 600: ", rr.Retry)
        t.Fatal()
    }
    if rr.Expire != 86400 {
        t.Error("Expected EXPIRE to be 86400: ", rr.Expire)
        t.Fatal()
    }
    if rr.Minttl != 10 {
        t.Error("Expected MINTTL to be 10: ", rr.Minttl)
        t.Fatal()
    }
}
