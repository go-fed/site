---
title: ActivityPub At A Glance
summary: By the time you have finished with this page you will have a basic and abstract understanding of how ActivityPub works. You will have enough contextual knowledge to help make decisions when using this library. Those already familiar with ActivityPub and ActivityStreams may find this page redundant.
---

## Introduction

By the time you have finished with this page you will have a basic and abstract understanding of how ActivityPub works. You will have enough contextual knowledge to help make decisions when using this library. Those already familiar with ActivityPub and ActivityStreams may find this page redundant.

Let's begin!

ActivityPub is actually two protocols in one, but both govern application behavior. The SocialAPI sets the rules for client-to-server interactions. When implemented, an ActivityPub client could interact with any other ActivityPub server. This lets a user use a single client on their phone to talk to their accounts on different kinds of ActivityPub servers.

The FederateAPI governs how two servers share data which lets users on all federating servers communicate with each other. Users on microblogs, photoblogs, video sites, and your future application can all interact!

Despite providing both protocols, ActivityPub does not require both be used. For example, Mastodon supports the FederateAPI but not the SocialAPI. It is up to you and your application's needs whether you want to use one, the other, or both!

To communicate, ActivityPub shares data in the ActivityStreams format. A piece of ActivityStream data is encoded as JSON when examined on the wire. However, it is actually built on top of JSON-LD which is a subset of JSON. To summarize a very deep topic like JSON-LD, it does two things. One: it is a rich data format (RDF) on top of JSON which effectively results in the JSON's schema being embedded within each JSON message. Two: it allows pieces of JSON data to refer to each other, resulting in a web of data that can be traversed like a graph.

This means when sending and receiving ActivityStreams via ActivityPub, your application is actually building a graph of data. As new data is generated, the graph gets bigger. The idea of "pointers" to other data looks like URLs, but they are technically IRIs.

The ActivityStreams specification doesn't just dictate a data format that is a subset of JSON-LD, it also specifies Core and Extended types of data. These are then used by ActivityPub to govern some basic behaviors in the SocialAPI and FederateAPI. Specific examples of these Core and Extended data types will be examined later on.

However, your application isn't limited to handling only Core and Extended ActivityStream data types. Since it is built on top of JSON-LD, the ActivityStreams vocabulary supports extensions beyond the Core and Extended types. This will be outside the scope of this overview, but Go-Fed can handle the expansion of RDF types at compile-time.

Let's also go over some things that ActivityPub does not support out of the box. There may be community conventions around these topics, the details of which are also outside the scope of this overview:

- The security protocols for authorization and authentication is not standardized. Some choices I am aware of are OAuth 2.0 and HTTP Signatures.
- Spam handling and blocking federating peers is not-standardized and usually implemented as an administrative application feature.
- The way to fetch a raw ActivityStream versus its human-readable HTML representation in static servers is not currently standardized.

## Under Construction

{{% caution %}}
Under construction.
{{% /caution %}}

## References

These are the references I used or referred to when building all libraries within the Go-Fed organization, including but not limited to the `go-fed/activity` library.

W3C Specifications:

- [Social Web Protocols](https://www.w3.org/TR/social-web-protocols/)
- [The ActivityPub Specification](https://www.w3.org/TR/activitypub)
- [The ActivityStreams Core Specification](https://www.w3.org/TR/activitystreams-core)
- [The ActivityStreams Vocabulary](https://www.w3.org/TR/activitystreams-vocabulary)
- [ActivityStream 2.0 Terms](https://www.w3.org/ns/activitystreams)
- [The JSON-LD Specification](https://www.w3.org/TR/json-ld)

RFCs:

- [RFC 3987: Internationalized Resource Identifiers (IRIs)](https://tools.ietf.org/html/rfc3987)
- [RFC 7033: Webfinger](https://tools.ietf.org/html/rfc7033)
- [RFC 3230: Instance Digests in HTTP](https://tools.ietf.org/html/rfc3230)
- [RFC 6749: The OAuth 2.0 Authorization Framework](https://tools.ietf.org/html/rfc6749)
- [RFC 6750: The OAuth 2.0 Authorization Framework: Bearer Token Usage](https://tools.ietf.org/html/rfc6750)
- [DRAFT: Signing HTTP Messages](https://tools.ietf.org/html/draft-cavage-http-signatures-10)

Other documents and links:

- [ActivityPub Authentication And Authorization Conventions](https://www.w3.org/wiki/SocialCG/ActivityPub/Authentication_Authorization)
- [Github Repo for W3C ActivityStreams](https://github.com/w3c/activitystreams)
- [Github Repo for W3C ActivityPub](https://github.com/w3c/activitypub)
- [Github Repo for W3C Draft of HTTP Signatures](https://github.com/w3c-dvcg/http-signatures)

