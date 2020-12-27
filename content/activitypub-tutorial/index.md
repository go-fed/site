---
title: App Tutorial
summary: This tutorial goes through the steps necessary to build a federating (Server-to-Server) application using [go-fed/activity](https://github.com/go-fed/activity) version 1.x.
---

This tutorial goes through the steps necessary to build a federating (Server-to-Server) application using [go-fed/activity](https://github.com/go-fed/activity) version 1.x.

{{% recommended %}}
Advice in these boxes call out suggested best practices.
{{% /recommended %}}

{{% caution %}}
Advice in these boxes call out pain points.
{{% /caution %}}

## Prepare {#Prepare}

First, let's `go get github.com/go-fed/activity`. Do not worry if the `go` tool complains about `package github.com/go-fed/activity: no Go files in github.com/go-fed/activity` because there are no Go files in the root library. This is OK, because we will be using the libraries under this directory at` github.com/go-fed/activity/streams` and `github.com/go-fed/activity/pub`.

Let's also start a new `myapp` program with a `main` package, so we can start up a simple server from the command line.

## An ActivityPub Mindset {#ActivityPub-Mindset}

ActivityPub is built on the concept of an actor. An actor is simply an entity, a person, a bot, or a logical unit of 'being'. Actors send and receive messages to and from each other in a Federated way. This tutorial will outline the concrete steps to do this, using Go-Fed.

{{% recommended %}}
While this tutorial focuses exclusively on the Server-to-Server (or S2S, or Federating Protocol) part of ActivityPub, it is important to remember that ActivityPub defines 2 protocols! The other federating protocol is Client-to-Server (or C2S, or Social Protocol). Go-Fed is designed for both.
{{% /recommended %}}

Let's take a look at what ActivityPub wants us to do, how Go-Fed approaches it, and what we need to do to use Go-Fed.

ActivityPub is built on the concept of linked data: if the value isn't literally there in a payload, a link (an IRI) is there so it can be fetched. That means we'll need to serve some data at HTTP endpoints. On top of this, actors name certain HTTP endpoints special things, like an "inbox" or "outbox". Together, they logically form a presentation of the actor to the outside world. We will also need to support the required ActivityPub behavior at the inbox and outbox endpoints.

Here's a basic outline of the kinds of HTTP endpoints we will need:

{{< rawhtml >}}
<div class="svg-container">
<svg height="500" width="480">
  
  <marker id="arrow" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 0 L 10 5 L 0 10 z"></path>
  </marker>
  
  <rect x="1" y="1" width="478" height="478" class="svgborder"></rect>

  
  <rect x="245" y="5" width="230" height="385" class="svgactor"></rect>
  <text x="340" y="25" class="svgtextsmall">Actor</text>
  <text x="337" y="45" class="svgtextsmaller">Concept</text>
  
  <rect x="250" y="55" width="220" height="75" class="svgborder svgpubactor"></rect>
  <text x="310" y="85" class="svgtextsmall">Actor Inbox</text>
  <text x="280" y="110" class="svgtextsmaller">HTTP POST*, HTTP GET*</text>
  
  <rect x="250" y="140" width="220" height="75" class="svgborder svgpubactor"></rect>
  <text x="305" y="170" class="svgtextsmall">Actor Outbox</text>
  <text x="280" y="195" class="svgtextsmaller">HTTP POST**, HTTP GET</text>
  
  <rect x="250" y="225" width="220" height="75" class="svgborder svgashandler"></rect>
  <text x="335" y="255" class="svgtextsmall">Actor</text>
  <text x="325" y="280" class="svgtextsmaller">HTTP GET</text>
  
  <rect x="250" y="310" width="220" height="75" class="svgborder svgashandler"></rect>
  <text x="265" y="340" class="svgtextsmall">Followers, Liked, etc.</text>
  <text x="325" y="365" class="svgtextsmaller">HTTP GET</text>
  
  <rect x="245" y="400" width="230" height="75" class="svgborder svgashandler"></rect>
  <text x="295" y="430" class="svgtextsmall">Other Content</text>
  <text x="325" y="455" class="svgtextsmaller">HTTP GET</text>
  
  <line x1="125" y1="35" x2="245" y2="35" class="svgdepline"></line>
  <line x1="125" y1="35" x2="125" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="110" y1="190" x2="110" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="95" y1="190" x2="95" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="140" y1="190" x2="140" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="155" y1="190" x2="155" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="95" y1="190" x2="155" y2="190" class="svgdepline"></line>

  
  <rect x="70" y="225" width="110" height="75" class="svgborder svgremotepeer"></rect>
  <text x="77" y="255" class="svgtextsmall">Peer Actor</text>
  <text x="87" y="280" class="svgtextsmaller">(Federating)</text>
  
  <line x1="180" y1="257" x2="210" y2="257" class="svgdepline"></line>
  <line x1="210" y1="87" x2="210" y2="432" class="svgdepline"></line>
  <line x1="210" y1="87" x2="245" y2="87" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="172" x2="245" y2="172" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="257" x2="245" y2="257" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="342" x2="245" y2="342" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="432" x2="240" y2="432" class="svgdepline" marker-end="url(#arrow)"></line>
  
  
  <rect x="1" y="479" width="478" height="20" class="svgborder"></rect>
  <text x="5" y="493" class="svgtextsmaller">*Only S2S</text>
  <text x="120" y="493" class="svgtextsmaller">**Only C2S</text>
</svg>
</div>
{{< /rawhtml >}}

As you can see, the logical idea of presenting an actor has multiple sources of data that need to be presented to the outside world in a cohesive manner. That is represented by all the boxes in the blue "Actor" box above, and they can be enriched with as much additional intrinsic details as your app requires. Other forms of content not directly instrinsic to an actor, such as their notes or articles they've written, also need to be exposed in ActivityStreams form. That's the last box outside the blue "Actor" box.

Whatever the choice, the bare minimum required by ActivityPub are the inbox and outbox behavior. Go-Fed makes the early design choice to support this via an `pub.Actor` interface. That's the green boxes above. The light yellow are handled separately by `pub.HandlerFunc`. These two tools allow you to build up the required behaviors.

However! Go-Fed does not presume to know what kind of HTTP endpoints you want to map, which means you are responsible for determining that, say, `https://example.com/arbitrary/actors/peyton` represents an actor "peyton", but that `https://example.com/inboxes/peyton` is their inbox. That means you have full control over your HTTP server, and can defer behavior to Go-Fed when necessary.

Adding these layers to the previous image:

{{< rawhtml >}}
<div class="svg-container">
<svg height="500" width="780">
  
  <marker id="arrow" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 0 L 10 5 L 0 10 z"></path>
  </marker>
  
  <rect x="1" y="1" width="478" height="478" class="svgborder"></rect>

  
  <rect x="245" y="5" width="230" height="385" class="svgactor"></rect>
  <text x="340" y="25" class="svgtextsmall">Actor</text>
  <text x="337" y="45" class="svgtextsmaller">Concept</text>
  
  <rect x="250" y="55" width="220" height="75" class="svgborder svgpubactor"></rect>
  <text x="310" y="85" class="svgtextsmall">Actor Inbox</text>
  <text x="280" y="110" class="svgtextsmaller">HTTP POST*, HTTP GET*</text>
  
  <rect x="250" y="140" width="220" height="75" class="svgborder svgpubactor"></rect>
  <text x="305" y="170" class="svgtextsmall">Actor Outbox</text>
  <text x="280" y="195" class="svgtextsmaller">HTTP POST**, HTTP GET</text>
  
  <rect x="250" y="225" width="220" height="75" class="svgborder svgashandler"></rect>
  <text x="335" y="255" class="svgtextsmall">Actor</text>
  <text x="325" y="280" class="svgtextsmaller">HTTP GET</text>
  
  <rect x="250" y="310" width="220" height="75" class="svgborder svgashandler"></rect>
  <text x="265" y="340" class="svgtextsmall">Followers, Liked, etc.</text>
  <text x="325" y="365" class="svgtextsmaller">HTTP GET</text>
  
  <rect x="245" y="400" width="230" height="75" class="svgborder svgashandler"></rect>
  <text x="295" y="430" class="svgtextsmall">Other Content</text>
  <text x="325" y="455" class="svgtextsmaller">HTTP GET</text>
  
  <line x1="125" y1="35" x2="245" y2="35" class="svgdepline"></line>
  <line x1="125" y1="35" x2="125" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="110" y1="190" x2="110" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="95" y1="190" x2="95" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="140" y1="190" x2="140" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="155" y1="190" x2="155" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="95" y1="190" x2="155" y2="190" class="svgdepline"></line>

  
  <rect x="70" y="225" width="110" height="75" class="svgborder svgremotepeer"></rect>
  <text x="77" y="255" class="svgtextsmall">Peer Actor</text>
  <text x="87" y="280" class="svgtextsmaller">(Federating)</text>
  
  <line x1="180" y1="257" x2="210" y2="257" class="svgdepline"></line>
  <line x1="210" y1="87" x2="210" y2="432" class="svgdepline"></line>
  <line x1="210" y1="87" x2="245" y2="87" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="172" x2="245" y2="172" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="257" x2="245" y2="257" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="342" x2="245" y2="342" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="432" x2="240" y2="432" class="svgdepline" marker-end="url(#arrow)"></line>
  
  
  <rect x="1" y="479" width="478" height="20" class="svgborder"></rect>
  <text x="5" y="493" class="svgtextsmaller">*Only S2S</text>
  <text x="120" y="493" class="svgtextsmaller">**Only C2S</text>
  
  
  <rect x="479" y="1" width="299" height="498" class="svgborder"></rect>
  <rect x="500" y="87" width="110" height="350" class="svgborder svgyoublock"></rect>
  <text x="515" y="245" class="svgtextsmall">MyServer</text>
  <text x="520" y="270" class="svgtextsmaller">Map HTTP</text>
  <text x="519" y="290" class="svgtextsmaller">Handlers to</text>
  <text x="519" y="310" class="svgtextsmaller">go-fed calls</text>
  
  <rect x="645" y="87" width="120" height="75" class="svgborder svgpubactor"></rect>
  <text x="662" y="132" class="svgtextsmall">pub.Actor</text>
  <line x1="470" y1="92" x2="485" y2="92" class="svgdepline"></line>
  <line x1="470" y1="177" x2="485" y2="177" class="svgdepline"></line>
  <line x1="485" y1="92" x2="485" y2="177" class="svgdepline"></line>
  <line x1="485" y1="130" x2="500" y2="130" class="svgdepline"></line>
  <line x1="500" y1="130" x2="610" y2="130" class="svgdepline svgyouline"></line>
  <line x1="610" y1="130" x2="640" y2="130" class="svgdepline" marker-end="url(#arrow)"></line>
  
  <rect x="645" y="310" width="120" height="75" class="svgborder svgashandler"></rect>
  <text x="650" y="355" class="svgtextsmaller">pub.HandlerFunc</text>
  <line x1="470" y1="262" x2="485" y2="262" class="svgdepline"></line>
  <line x1="470" y1="347" x2="485" y2="347" class="svgdepline"></line>
  <line x1="475" y1="437" x2="485" y2="437" class="svgdepline"></line>
  <line x1="485" y1="262" x2="485" y2="437" class="svgdepline"></line>
  <line x1="485" y1="347" x2="500" y2="347" class="svgdepline"></line>
  <line x1="500" y1="347" x2="610" y2="347" class="svgdepline svgyouline"></line>
  <line x1="610" y1="347" x2="640" y2="347" class="svgdepline" marker-end="url(#arrow)"></line>
</svg>
</div>
{{< /rawhtml >}}

All we need to do is get an `pub.Actor` and a `pub.HandlerFunc`. These can be reused for any number of handlers concurrently, for any number of actual logical actors. To build these types, Go-Fed breaks its requirements down into these two:

- A `pub.Database` for persistent, concurrent-safe storage.
- A *behavior*, where you inject callbacks for Go-Fed to call in order to have a comprehensive application supporting S2S, C2S, or both.

The `pub.Database` is a straightforward interface that you need to implement, to meet the first requirement.

The *behavior* bit is trickier. The S2S and C2S parts of the ActivityPub specification can be taken separately or together. Either way, parts of them overlap. If Go-Fed defined all the S2S behaviors in one interface, and all the C2S behaviors in another interface, then some methods would be duplicated! Interfaces done in this way cannot be embedded into a single interface for the S2S-plus-C2S case, since Go will complain if two interfaces define the same method. Remedying this leads to sub-optimal interface design choices.

Therefore, Go-Fed actually breaks down the behavior into three interfaces:

- The `pub.CommonBehavior` is the behavior required regardless what kind of ActivityPub application you want. S2S, C2S, both? You must implement this interface.
- The `pub.FederatingProtocol` is additional behavior required only for any usage of S2S. This means the S2S-only case as well as the S2S-plus-C2S case.
- The `pub.SocialProtocol` is additional behavior required only for any usage of C2S. This means the C2S-only case as well as the S2S-plus-C2S case.

This still isn't optimal, but it at least follows the principle of composability. It is recommended to implement these interfaces onto one concrete type, so that all ActivityPub behavior is located in one place. Putting these interfaces all together with the previous picture, we get:

{{< rawhtml >}}
<div class="svg-container">
<svg height="500" width="1080">
  
  <marker id="arrow" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 0 L 10 5 L 0 10 z"></path>
  </marker>
  
  <rect x="1" y="1" width="478" height="478" class="svgborder"></rect>

  
  <rect x="245" y="5" width="230" height="385" class="svgactor"></rect>
  <text x="340" y="25" class="svgtextsmall">Actor</text>
  <text x="337" y="45" class="svgtextsmaller">Concept</text>
  
  <rect x="250" y="55" width="220" height="75" class="svgborder svgpubactor"></rect>
  <text x="310" y="85" class="svgtextsmall">Actor Inbox</text>
  <text x="280" y="110" class="svgtextsmaller">HTTP POST*, HTTP GET*</text>
  
  <rect x="250" y="140" width="220" height="75" class="svgborder svgpubactor"></rect>
  <text x="305" y="170" class="svgtextsmall">Actor Outbox</text>
  <text x="280" y="195" class="svgtextsmaller">HTTP POST**, HTTP GET</text>
  
  <rect x="250" y="225" width="220" height="75" class="svgborder svgashandler"></rect>
  <text x="335" y="255" class="svgtextsmall">Actor</text>
  <text x="325" y="280" class="svgtextsmaller">HTTP GET</text>
  
  <rect x="250" y="310" width="220" height="75" class="svgborder svgashandler"></rect>
  <text x="265" y="340" class="svgtextsmall">Followers, Liked, etc.</text>
  <text x="325" y="365" class="svgtextsmaller">HTTP GET</text>
  
  <rect x="245" y="400" width="230" height="75" class="svgborder svgashandler"></rect>
  <text x="295" y="430" class="svgtextsmall">Other Content</text>
  <text x="325" y="455" class="svgtextsmaller">HTTP GET</text>
  
  <line x1="125" y1="35" x2="245" y2="35" class="svgdepline"></line>
  <line x1="125" y1="35" x2="125" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="110" y1="190" x2="110" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="95" y1="190" x2="95" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="140" y1="190" x2="140" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="155" y1="190" x2="155" y2="220" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="95" y1="190" x2="155" y2="190" class="svgdepline"></line>

  
  <rect x="70" y="225" width="110" height="75" class="svgborder svgremotepeer"></rect>
  <text x="77" y="255" class="svgtextsmall">Peer Actor</text>
  <text x="87" y="280" class="svgtextsmaller">(Federating)</text>
  
  <line x1="180" y1="257" x2="210" y2="257" class="svgdepline"></line>
  <line x1="210" y1="87" x2="210" y2="432" class="svgdepline"></line>
  <line x1="210" y1="87" x2="245" y2="87" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="172" x2="245" y2="172" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="257" x2="245" y2="257" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="342" x2="245" y2="342" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="210" y1="432" x2="240" y2="432" class="svgdepline" marker-end="url(#arrow)"></line>
  
  
  <rect x="1" y="479" width="478" height="20" class="svgborder"></rect>
  <text x="5" y="493" class="svgtextsmaller">*Only S2S</text>
  <text x="120" y="493" class="svgtextsmaller">**Only C2S</text>
  
  
  <rect x="479" y="1" width="599" height="498" class="svgborder"></rect>
  <rect x="500" y="87" width="110" height="350" class="svgborder svgyoublock"></rect>
  <text x="515" y="245" class="svgtextsmall">MyServer</text>
  <text x="520" y="270" class="svgtextsmaller">Map HTTP</text>
  <text x="519" y="290" class="svgtextsmaller">Handlers to</text>
  <text x="519" y="310" class="svgtextsmaller">go-fed calls</text>
  
  <rect x="645" y="87" width="120" height="75" class="svgborder svgpubactor"></rect>
  <text x="662" y="132" class="svgtextsmall">pub.Actor</text>
  <line x1="470" y1="92" x2="485" y2="92" class="svgdepline"></line>
  <line x1="470" y1="177" x2="485" y2="177" class="svgdepline"></line>
  <line x1="485" y1="92" x2="485" y2="177" class="svgdepline"></line>
  <line x1="485" y1="130" x2="500" y2="130" class="svgdepline"></line>
  <line x1="500" y1="130" x2="610" y2="130" class="svgdepline svgyouline"></line>
  <line x1="610" y1="130" x2="640" y2="130" class="svgdepline" marker-end="url(#arrow)"></line>
  
  <rect x="645" y="310" width="120" height="75" class="svgborder svgashandler"></rect>
  <text x="650" y="355" class="svgtextsmaller">pub.HandlerFunc</text>
  <line x1="470" y1="262" x2="485" y2="262" class="svgdepline"></line>
  <line x1="470" y1="347" x2="485" y2="347" class="svgdepline"></line>
  <line x1="475" y1="437" x2="485" y2="437" class="svgdepline"></line>
  <line x1="485" y1="262" x2="485" y2="437" class="svgdepline"></line>
  <line x1="485" y1="347" x2="500" y2="347" class="svgdepline"></line>
  <line x1="500" y1="347" x2="610" y2="347" class="svgdepline svgyouline"></line>
  <line x1="610" y1="347" x2="640" y2="347" class="svgdepline" marker-end="url(#arrow)"></line>
  
  <rect x="790" y="10" width="155" height="60" class="svgborder svginterface"></rect>
  <text x="797" y="45" class="svgtextsmaller">pub.CommonBehavior</text>
  <rect x="790" y="95" width="155" height="60" class="svgborder svginterface"></rect>
  <text x="792" y="130" class="svgtextsmaller">pub.FederatingProtocol*</text>
  <rect x="790" y="180" width="155" height="60" class="svgborder svginterface"></rect>
  <text x="805" y="215" class="svgtextsmaller">pub.SocialProtocol**</text>
  <rect x="790" y="265" width="155" height="60" class="svgborder svginterface"></rect>
  <text x="825" y="300" class="svgtextsmaller">pub.Database</text>
  <rect x="760" y="400" width="155" height="60" class="svgborder svgyoublock"></rect>
  <text x="782" y="435" class="svgtextsmall">MyDatabase</text>
  <rect x="920" y="400" width="155" height="60" class="svgborder svgyoublock"></rect>
  <text x="952" y="435" class="svgtextsmall">MyService</text>
  
  <line x1="777" y1="300" x2="777" y2="45" class="svgdepline"></line>
  <line x1="777" y1="45" x2="785" y2="45" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="765" y1="125" x2="785" y2="125" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="777" y1="210" x2="785" y2="210" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="777" y1="300" x2="785" y2="300" class="svgdepline" marker-end="url(#arrow)"></line>
  
  <line x1="765" y1="347" x2="820" y2="347" class="svgdepline"></line>
  <line x1="820" y1="347" x2="820" y2="330" class="svgdepline" marker-end="url(#arrow)"></line>
  
  <line x1="850" y1="325" x2="850" y2="395" class="svgdepline" marker-end="url(#arrow)"></line>
  
  <line x1="945" y1="45" x2="995" y2="45" class="svgdepline"></line>
  <line x1="945" y1="125" x2="995" y2="125" class="svgdepline"></line>
  <line x1="945" y1="210" x2="995" y2="210" class="svgdepline"></line>
  <line x1="995" y1="45" x2="995" y2="395" class="svgdepline" marker-end="url(#arrow)"></line>
</svg>
</div>
{{< /rawhtml >}}

Minor detail: there is also a `pub.Clock` interface, but it is not worth elaborating further upon.

Thus, this tutorial will concretely focus on a S2S demo app:

- Implementing `pub.CommonBehavior`
- Implementing `pub.FederatingProtocol`
- Implementing `pub.Database`
- Acquiring and using a `pub.Actor` in HTTP handlers, specifically a `pub.FederatingActor`
- Acquiring and using a `pub.HandlerFunc` in HTTP handlers
- Programmatically sending messages to the Fediverse via the `pub.FederatingActor`
- Gloss over how to use ActivityStreams types and properties in the `github.com/go-fed/activity/streams` and `github.com/go-fed/activity/streams/vocab` packages

## Stubbing For An Actor {#Stubbing-Future}

Actors are at the core of the ActivityPub specification, so we will create one in our application. We have the option to support either Client-to-Server, Server-to-Server, or both forms of federation. We will only be supporting the Server-to-Server option, so let's call `pub.NewFederatingActor`. Its signature looks like:

```go
// NewFederatingActor builds a new Actor concept that handles only the Federating
// Protocol part of ActivityPub.
//
// This Actor can be created once in an application and reused to handle
// multiple requests concurrently and for different endpoints.
//
// It leverages as much of Go-Fed as possible to ensure the implementation is
// compliant with the ActivityPub specification, while providing enough freedom
// to be productive without shooting one's self in the foot.
//
// Do not try to use NewSocialActor and NewFederatingActor together to cover
// both the Social and Federating parts of the protocol. Instead, use NewActor.
func NewFederatingActor(c CommonBehavior,
	s2s FederatingProtocol,
	db Database,
	clock Clock) FederatingActor
```

Oh! Looks like we aren't ready to call it yet! We can see we will need to pass in some interfaces. Let's stub out some types and revisit their implementations later. First, the stub for the common behavior on a new type we will internally call `myService`. We won't visit these specific method implementations here, but detailed guidance is provided in the [activity/pub documentation](https://go-fed.org/ref/activity/pub#The-CommonBehavior-Interface):

```go
type myService struct {}
func (*myService) AuthenticateGetInbox(c context.Context,
	w http.ResponseWriter,
	r *http.Request) (out context.Context, authenticated bool, err error) {
	// TODO
	return
}

func (*myService) AuthenticateGetOutbox(c context.Context,
	w http.ResponseWriter,
	r *http.Request) (out context.Context, authenticated bool, err error) {
	// TODO
	return
}

func (*myService) GetOutbox(c context.Context,
	r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// TODO
	return nil, nil
}

func (*myService) NewTransport(c context.Context,
	actorBoxIRI *url.URL,
	gofedAgent string) (t pub.Transport, err error) {
	// TODO
	return
}
```

Next up, let's stub out the methods for the `FederatingProtocol`. Let's also put them on the `myService` type we just created. We will visit one of these methods in detail, but for the others the [activity/pub documentation](https://go-fed.org/ref/activity/pub#The-FederatingProtocol-Interface) should act as a handy guide:

```go
func (*myService) PostInboxRequestBodyHook(c context.Context,
	r *http.Request,
	activity Activity) (context.Context, error) {
	// TODO
	return nil, nil
}

func (*myService) AuthenticatePostInbox(c context.Context,
	w http.ResponseWriter,
	r *http.Request) (out context.Context, authenticated bool, err error) {
	// TODO
	return
}

func (*myService) Blocked(c context.Context,
	actorIRIs []*url.URL) (blocked bool, err error) {
	// TODO
	return
}

func (*myService) FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error) {
	// TODO
	return
}

func (*myService) DefaultCallback(c context.Context,
	activity Activity) error {
	// TODO
	return nil
}

func (*myService) MaxInboxForwardingRecursionDepth(c context.Context) int {
	// TODO
	return -1
}

func (*myService) MaxDeliveryRecursionDepth(c context.Context) int {
	// TODO
	return -1
}

func (*myService) FilterForwarding(c context.Context,
	potentialRecipients []*url.URL,
	a Activity) (filteredRecipients []*url.URL, err error) {
	// TODO
	return
}

func (*myService) GetInbox(c context.Context,
	r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	// TODO
	return nil, nil
}
```

Alright! Now for the database, let's create an in-memory based database. This won't be good for a real application, but for a demo app it should suffice. A real world application would want to use a real database solution under the hood that can handle real persistence. Let us quickly write a database called `myDB`. We will begin to make use of the `github.com/go-fed/activity/pub` and `github.com/go-fed/activity/streams/vocab` libraries.

{{% caution %}}
It's important to note that the database implementation can be thought of as being independent from ActivityStreams behaviors. So, don't confuse the `Create` ActivityStreams type with the Database Create method we write.
{{% /caution %}}

This is what a partial demo implementation could look like, and the [activity/pub documentation](https://go-fed.org/ref/activity/pub#The-Database-Interface) should help in implementing the other methods:

```go
type myDB struct {
	// The content of our app, keyed by ActivityPub ID.
	content *sync.Map
	// Enables mutations. A sync.Mutex per ActivityPub ID.
	locks *sync.Map
	// The host domain of our service, for detecting ownership.
	hostname string
}

// Our content map will store this data.
type content struct {
	// The payload of the data: vocab.Type is any type understood by Go-Fed.
	data vocab.Type
	// If true, belongs to our local user and not a federated peer. This is
	// recommended for a solution that just indiscriminately puts everything
	// into a single "table", like this in-memory solution.
	isLocal bool
}

func (m *myDB) Lock(c context.Context,
	id *url.URL) error {
	// Before any other Database methods are called, the relevant `id`
	// entries are locked to allow for fine-grained concurrency.

	// Strategy: create a new lock, if stored, continue. Otherwise, lock the
	// existing mutex.
	mu := &sync.Mutex{}
	mu.Lock() // Optimistically lock if we do store it.
	i, loaded := m.locks.LoadOrStore(id.String(), mu)
	if loaded {
		mu = i.(*sync.Mutex)
		mu.Lock()
	}
	return nil
}

func (m *myDB) Unlock(c context.Context,
	id *url.URL) error {
	// Once Go-Fed is done calling Database methods, the relevant `id`
	// entries are unlocked.

	i, ok := m.locks.Load(id.String())
	if !ok {
		return errors.New("Missing an id in Unlock")
	}
	mu := i.(*sync.Mutex)
	mu.Unlock()
	return nil
}

func (m *myDB) Owns(c context.Context,
	id *url.URL) (owns bool, err error) {
	// Owns just determines if the ActivityPub id is owned by this server.
	// In a real implementation, consider something far more robust than
	// this string comparison.
	return id.Host == m.hostname, nil
}

func (m *myDB) Exists(c context.Context,
	id *url.URL) (exists bool, err error) {
	// Do we have this `id`?
	_, exists = m.content.Load(id.String())
	return
}

func (m *myDB) Get(c context.Context,
	id *url.URL) (value vocab.Type, err error) {
	// Our goal is to return what we have at that `id`. Returns an error if
	// not found.
	iCon, exists = m.content.Load(id.String())
	if !exists {
		err = errors.New("Get failed")
		return
	}
	// Extract the data from our `content` type.
	con := iCon.(*content)
	return con.data
}

func (m *myDB) Create(c context.Context,
	asType vocab.Type) error {
	// Create a payload in our in-memory map. The thing could be a local or
	// a federated peer's data. We can re-use the `Owns` call to set the
	// metadata on our `content`.
	id, err := pub.GetId(asType)
	if err != nil {
		return err
	}
	owns, err := m.Owns(id)
	if err != nil {
		return err
	}
	con = &content {
		data: asType,
		isLocal: owns,
	}
	m.content.Store(id.String(), con)
	return nil
}

func (m *myDB) Update(c context.Context,
	asType vocab.Type) error {
	// Replace a payload in our in-memory map. The thing could be a local or
	// a federated peer's data. Since we are using a map and not a solution
	// like SQL, we can simply do what `Create` does: overwrite it.
	//
	// Note that an actor's followers, following, and liked collections are
	// never Created, only Updated.
	return m.Create(c, asType)
}

func (m *myDB) Delete(c context.Context,
	id *url.URL) error {
	// Remove a payload in our in-memory map.
	m.Delete(id.String())
	return nil
}

func (m *myDB) InboxContains(c context.Context,
	inbox,
	id *url.URL) (contains bool, err error) {
	// Our goal is to see if the `inbox`, which is an OrderedCollection,
	// contains an element in its `ordered_items` property that has a
	// matching `id`
	contains = false
	var oc vocab.ActivityStreamsOrderedCollection
	// getOrderedCollection is a helper method to fetch an
	// OrderedCollection. It is not implemented in this tutorial, and uses
	// the map m.content to do the lookup.
	oc, err = m.getOrderedCollection(inbox)
	if err != nil {
		return
	}
	// Next, we use the ActivityStreams vocabulary to obtain the
	// ordered_items property of the OrderedCollection type.
	oi := oc.GetActivityStreamsOrderedItems()
	// Properties may be nil, if non-existent!
	if oi == nil {
		return
	}
	// Finally, loop through each item in the ordered_items property and see
	// if the element's id matches the desired id.
	for iter := oi.Begin(); iter != oi.End(); iter = iter.Next() {
		var iterId *url.URL
		iterId, err = pub.ToId(iter)
		if err != nil {
			return
		}
		if iterId.String() == id.String() {
			contains = true
			return
		}
	}
	return
}

func (m *myDB) GetInbox(c context.Context,
	inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// The goal here is to fetch an inbox at the specified IRI.

	// getOrderedCollectionPage is a helper method to fetch an
	// OrderedCollectionPage. It is not implemented in this tutorial, and
	// uses the map m.content to do the lookup and any conversions if
	// needed. The database can get fancy and use query parameters in the
	// `inboxIRI` to paginate appropriately.
	return m.getOrderedCollectionPage(inboxIRI)
}

func (m *myDB) SetInbox(c context.Context,
	inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// The goal here is to set an inbox at the specified IRI, with any
	// changes to the page made persistent. Since the inbox has been Locked,
	// it is OK to assume that no other concurrent goroutine has changed the
	// inbox in the meantime.

	// getOrderedCollection is a helper method to fetch an
	// OrderedCollection. It is not implemented in this tutorial, and
	// uses the map m.content to do the lookup.
	storedInbox, err := m.getOrderedCollection(inboxIRI)
	if err != nil {
		return err
	}
	// applyDiffOrderedCollection is a helper method to apply changes due
	// to an edited OrderedCollectionPage. Implementation is left as an
	// exercise for the reader.
	updatedInbox := m.applyDiffOrderedCollection(storedInbox, inbox)
	
	// saveToContent is a helper method to save an
	// ActivityStream type. Implementation is left as an exercise for the
	// reader.
	return m.saveToContent(updatedInbox)
}

func (m *myDB) GetOutbox(c context.Context,
	inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	// Similar to `GetInbox`, but for the outbox. See `GetInbox`.
}

func (m *myDB) SetOutbox(c context.Context,
	inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// Similar to `SetInbox`, but for the outbox. See `SetInbox`.
}

func (m *myDB) ActorForOutbox(c context.Context,
	outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	// Given the `outboxIRI`, determine the IRI of the actor that owns
	// that outbox. Will only be used for actors on this local server.
	// Implementation left as an exercise to the reader.
}

func (m *myDB) ActorForInbox(c context.Context,
	inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	// Given the `inboxIRI`, determine the IRI of the actor that owns
	// that inbox. Will only be used for actors on this local server.
	// Implementation left as an exercise to the reader.
}

func (m *myDB) OutboxForInbox(c context.Context,
	inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	// Given the `inboxIRI`, determine the IRI of the outbox owned
	// by the same actor that owns the inbox. Will only be used for actors
	// on this local server. Implementation left as an exercise to the
	// reader.
}

func (m *myDB) NewID(c context.Context,
	t vocab.Type) (id *url.URL, err error) {
	// Generate a new `id` for the ActivityStreams object `t`.

	// You can be fancy and put different types authored by different folks
	// along different paths. Or just generate a GUID. Implementation here
	// is left as an exercise for the reader.
}

func (m *myDB) Followers(c context.Context,
	actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	// Get the followers collection from the actor with `actorIRI`.

	// getPerson is a helper method that returns an actor on this server
	// with a Person ActivityStreams type. It is not implemented in this tutorial.
	var person vocab.ActivityStreamsPerson
	person, err = m.getPerson(actorIRI)
	if err != nil {
		return
	}
	// Let's get their followers property, ensure it exists, and then
	// fetch it with a familiar helper method.
	f := person.GetActivityStreamsFollowers()
	if f == nil {
		err = errors.New("no followers collection")
		return
	}
	// Note: at this point f is not the OrderedCollection itself yet. It is
	// an opaque box (it could be an IRI, an OrderedCollection, or something
	// extending an OrderedCollection).
	followersId, err := pub.ToId(f)
	if err != nil {
		return
	}
	return m.getOrderedCollection(followersId)
}

func (m *myDB) Following(c context.Context,
	actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	// Get the following collection from the actor with `actorIRI`.

	// Implementation is similar to `Followers`. See `Followers`.
}

func (m *myDB) Liked(c context.Context,
	actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	// Get the liked collection from the actor with `actorIRI`.

	// Implementation is similar to `Followers`. See `Followers`.
}
```

Wow! OK, one more to go. The `Clock` interface is super easy, let's just tack on the one method to `myService`:

```go
func (*myService) Now() time.Time {
	return time.Now()
}
```

Great! We can now get a `FederatingActor`!

## Get A Federating Actor {#Get-Federating-Actor}

With the stubs in the previous section, we can now properly obtain an actor in our main method:

```go
func main() {
	s := &myService{}
	db := &myDB{
		content: &sync.Map{},
		locks: &sync.Map{},
		hostname: "localhost",
	}
	actor := pub.NewFederatingActor(/* CommonBehavior */ s,
		/* FederatingProtocol */ s,
		/* Database */ db,
		/* Clock */ s)
}
```

There's two things left to do:

1. Finish implementing the stubs for the federating behavior.
1. Hooking this actor into our HTTP server

Let's tackle the first one here. The second one will be addressed in the next section.

When we stubbed out the behaviors for the federating behaviors earlier, we stubbed out functions that configure the actor's behavior within the ActivityPub protocol, and we stubbed out the functions required to give us the juicy app behaviors we want to customize. Configuration is boring, so let's revisit the `Callbacks` stubbed method.

The `Callbacks` method's job is to provide the hooks you want the Go-Fed library to call into when it receives an ActivityStreams piece of data from a peer. It will have already passed through the other kinds of checks you've configured such as Authentication and Blocked. Go-Fed provides a bunch of default behavior for you out of the box, so a valid implementation that handles Activities like Create, Update, Delete, Add, Remove, and the others listed in the specification is simply:

```go
func (*myService) FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error) {
	// Return the default ActivityPub callbacks, and nothing in `other`.
	return
}
```

But defaults are boring! Let's say every time you get a Like from a peer, you want your app to light up a disco ball light with your app's `DiscoParty` function. We can add that in addition to the existing ActivityPub behavior of "adding a Like Activity to the likes collection of all targeted Objects that are owned on this instance" by doing this complicated maneuver:

```go
func (*myService) FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error) {
	wrapped = FederatingWrappedCallbacks{
		// Anything we set in the callbacks, is in *addition* to the out-of-the-box support.
		Like: func(c context.Context, like vocab.ActivityStreamsLike) error {
			// We could do something with the `like`, but for now it's time to party.
			DiscoParty()
			return nil
		},
	}
	return
}
```

Next, let's say a federated peer gives your app a Flag Activity. But, the `FederatingWrappedCallbacks` doesn't have a spot for the Flag Activity, because it isn't providing a default behavior! Oh no! World's over, time to pack up and leave. Or, you simply put your callback in the `other` variable:

```go
func (*myService) FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error) {
	other = []interface{}{
		// Elements in `other` need to follow this function signature pattern.
		func(c context.Context, flag vocab.ActivityStreamsFlag) error {
			// We can now look at `flag` to turn the avatar of the person who got flagged
			// into a giant baby picture.

			// Note: you're in charge of checking `target` and `object` to make sure it is applicable.
			return nil
		},
	}
	return
}
```

Finally, let's say you don't want a default behavior that Go-Fed provides out of the box. Hey, no hard feelings. I get it, not every match is made in heaven. There's a way to completely override the default behavior in a very delicate way: simply provide the function in the `other` variable:

```go
func (*myService) FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error) {
	other = []interface{}{
		// This element follows the function signature pattern, but FederatingWrappedCallbacks
		// has a default for Add! Therefore, Go-Fed will pick the function here, completely replacing
		// the default behavior.
		func(c context.Context, add vocab.ActivityStreamsAdd) error {
			// This function does nothing, overriding the default behavior for the Add
			// Activity. In this case, it's like Go-Fed never provided a default at all.
			return nil
		},
	}
	wrapped = FederatingWrappedCallbacks{
		// Add's default behavior will NOT be called, but Activities like Create, Delete, etc will still
		// have their default behaviors called.
		Add: func(c context.Context, add vocab.ActivityStreamsAdd) error {
			// Will NOT be called, because it is a part of the Add default behavior,
			// which is being overridden!
			return nil
		},
	}
	return
}
```

As you can see, when building your application you can start off using the default behaviors provided by Go-Fed. Then, as it grows, you can completely customize it as you see fit.

Let's breeze through the rest of the stubs discussing what is expected in each in order to have an ActivityPub compliant implementation, though the [activity/pub documentation](https://go-fed.org/ref/activity/pub) goes into more detail:

- `AuthenticateGetInbox`: Provides a way for you to deny GET HTTP ActivityPub requests for an actor's inbox. It is still compliant to do nothing, though.
- `AuthenticateGetOutbox`: Same as `AuthenticateGetInbox` but for outboxes.
- `GetOutbox`: For the given `http.Request`, craft the right IRI and fetch the outbox, which may be as simple as calling `myDB.GetOutbox`
- `GetInbox`: Similar to `GetOutbox` but for the inbox.
- `NewTransport`: Returns a `Transport` which is responsible for physically shoving bytes out the door, so to speak. One can be obtained by calling `pub.NewHttpSigTransport`, which is HTTP Signatures based, or creating your own.
- `PostInboxRequestBodyHook`: This is to make your life easier in the HTTP request-handling cycle. This is a hook for you in case you need to set up `context.Context` based on the `http.Request` after the body has been read and interpreted, but before it has been checked for authorization, blocks, etc.
- `AuthenticatePostInbox`: Responsible for authenticating the incoming peer message into the actor's inbox. The community standard is HTTP Signatures, so this is where you would verify such a signature.
- `Blocked`: Given a list of actors, simply return true if processing the Activity should stop. This prevents the peer message from having any application-level side effects defined in `Callbacks`.
- `DefaultCallback`: If a peer tries to send you something that neither Go-Fed nor your app can understand, this will be called with the offending Activity.
- `MaxDeliveryRecursionDepth`: When your app attempts to deliver messages, it may hit cases where a target is a list of actors, or a list of a list of other actors, or even deeper nestings. This limits how far your app will search before it quits building up its list of potential recipients.
- `MaxInboxForwardingRecursionDepth`: When doing inbox forwarding in a deep chain of activities, this is the limit to how deep it will search in order to attempt to determine if it should do inbox forwarding at all.
- `FilterForwarding`: Given a list of recipients for the given activity, you MUST filter down the recipients in some way so that your server does not blindly forward messages that are inappropriate when doing inbox forwarding.

{{% caution %}}
The Activity passed into `PostInboxRequestBodyHook` should be treated as untrusted, as it has not yet passed authorization.
{{% /caution %}}

{{% recommended %}}
Sizable but finite limits are strongly recommended for `MaxDeliveryRecursionDepth` and `MaxInboxForwardingRecursionDepth`.
{{% /recommended %}}

{{% caution %}}
You MUST apply some sort of filtering in `FilterForwarding`, typically limiting it to followers of the receiving actor, otherwise your application server will become a spam vector. If that happens, communities using your software will be defederated quickly.
{{% /caution %}}

There we go! Now you have an actor-aware, ActivityPub ready implementation. It is already hooked into the behaviors of your application. Now, all that remains is to set up the HTTP routing to match the IRI paths used when creating this implementation.

## Hooking It All Together {#Hooking-It-All-Together}

An `Actor`, like the `FederatingActor` we have, has only 4 methods (comments omitted):

```go
type Actor interface {
	PostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
	GetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
	PostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
	GetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (bool, error)
}
```

These are built around handling an actor's inbox and outbox. It is sufficient to call into these methods in a normal `http.ServeMux` that maps a path for an inbox or outbox. In fact, most of the challenge in this section is mentally keeping track of which paths are meant to represent an actor, their inbox, their outbox, etc. The paths you use to hook into the `http.ServeMux` will also need to match the `id` properties of the ActivityStreams data you serve. If your `Database` is designed to do this when it is told to Get something, this property naturally arises.

For the demo app, let's only have one actor we want to logically represent: me (aka: you)! Here is a basic set up:

```go
actor := pub.NewFederatingActor(s, s, db, s)
mux := http.NewServeMux()
// Map the `me` actor's inbox to the path `/actors/me/inbox`
mux.HandleFunc("/actors/me/inbox", func(w http.ResponseWriter, r *http.Request) {
	if isActivityPubRequest, err := actor.GetInbox(w, r); err != nil {
		// Do something with `err`
		return
	} else if isActivityPubRequest {
		// Go-Fed handled the ActivityPub GET request to the inbox
		return
	} else if isActivityPubRequest, err := actor.PostInbox(w, r); err != nil {
		// Do something with `err`
		return
	} else if isActivityPubRequest {
		// Go-Fed handled the ActivityPub POST request to the inbox
		return
	}
	// Here we return an error, but you may just as well decide
	// to render a webpage instead. But be sure to apply appropriate
	// authorizations. There's no guarantees about authorization at
	// this point.
	http.Error("Non-ActivityPub request", http.StatusBadRequest)
	return
})
// Map the `me` actor's inbox to the path `/arbitrary/me/outbox`
mux.HandleFunc("/arbitrary/me/outbox", func(w http.ResponseWriter, r *http.Request) {
	if isActivityPubRequest, err := actor.GetOutbox(w, r); err != nil {
		// Do something with `err`
		return
	} else if isActivityPubRequest {
		// Go-Fed handled the ActivityPub GET request to the outbox
		return
	} else if isActivityPubRequest, err := actor.PostOutbox(w, r); err != nil {
		// Do something with `err`
		return
	} else if isActivityPubRequest {
		// Go-Fed handled the ActivityPub POST request to the outbox
		return
	}
	// Here we return an error, but you may just as well decide
	// to render a webpage instead. But be sure to apply appropriate
	// authorizations. There's no guarantees about authorization at
	// this point.
	http.Error("Non-ActivityPub request", http.StatusBadRequest)
	return
})
```

{{% recommended %}}
Rather than using non-sensical paths like `"/actors/me/inbox"` and `"/arbitrary/me/outbox"`, use paths that make sense for your application and in a way that makes sense. Examples will continue to use paths like these, only to show that you still have the power to design your application's HTTP server really well -- or really hard to understand.
{{% /recommended %}}

Pretty straightforward to use! Note here that we are using one `Actor` to logically map to one actor in our demo app. This is not a limitation. A Go-Fed `Actor` can handle any number of actors, so feel free to use more complex Mux solutions that lets you handle any number of actors, all calling into the same Go-Fed `Actor`. Think of the Go-Fed `Actor` being a definition of behavior, not state. It is stateless, but your injected state via the `Database` gives it the data to operate on.

Finally, we need to serve our actor's ActivityStreams data itself. We're serving the inbox and outbox now, which is dandy, but no one can discover them since they are a part of the not-yet-served actor.

For any other data that isn't inboxes and outboxes, that is simply a GET to a resource, the `github.com/go-fed/activity/pub` package has a helper that makes it a breeze:

```go
// NewActivityStreamsHandler creates a HandlerFunc to serve ActivityStreams
// requests which are coming from other clients or servers that wish to obtain
// an ActivityStreams representation of data.
//
// Strips retrieved ActivityStreams values of sensitive fields ('bto' and 'bcc')
// before responding with them. Sets the appropriate HTTP status code for
// Tombstone Activities as well.
func NewActivityStreamsHandler(db Database, clock Clock) HandlerFunc
```

We've already created a `Database` and a `Clock`, so obtaining one is easy:

```go
asHandler := pub.NewActivityStreamsHandler(db, s)
```

So, what is this `HandlerFunc`? It's a function very similar to the inbox and outbox functions on the `Actor`:

```go
// HandlerFunc determines whether an incoming HTTP request is an ActivityStreams
// GET request, and if so attempts to serve ActivityStreams data.
//
// If an error is returned, then the calling function is responsible for writing
// to the ResponseWriter as part of error handling.
//
// If 'isASRequest' is false and there is no error, then the calling function
// may continue processing the request, and the HandlerFunc will not have
// written anything to the ResponseWriter. For example, a webpage may be served
// instead.
//
// If 'isASRequest' is true and there is no error, then the HandlerFunc
// successfully served the request and wrote to the ResponseWriter.
//
// Callers are responsible for authorized access to this resource.
type HandlerFunc func(c context.Context, w http.ResponseWriter, r *http.Request) (isASRequest bool, err error)
```

So if we wanted to serve our actor at a specific HTTP endpoint, we can do the following:

```go
// Host the `me` actor at `/anything/me`
mux.HandleFunc("/anything/me", func(w http.ResponseWriter, r *http.Request) {
	// If any authentication/authorization needs to happen, apply it here
	if isActivityPubRequest, err := asHandler(r.Context(), w, r); err != nil {
		// Do something with `err`
		return
	} else if isActivityPubRequest {
		// Go-Fed handled the ActivityPub GET request for this particular IRI
		return
	}
	// Here we return an error, but you may just as well decide
	// to render a webpage instead. But be sure you've already
	// applied the appropriate authorizations.
	http.Error("Non-ActivityPub request", http.StatusBadRequest)
	return
})
```

{{% caution %}}
Check your `Database`'s Get method to ensure that when returning "me" actor data for `"/anything/me"`, the database is correctly setting the inbox to `"/actors/me/inbox"` and outbox to `"/arbitrary/me/outbox"`. Keeping your HTTP endpoints and linked-data links in ActivityStreams synchronized is a necessary headache.
{{% /caution %}}

Similar to how the `Actor` is a create-once and use-multiple-times type, the `HandlerFunc` can be reused for every path you want to serve ActivityStreams data:

```go
// Notice we are serving a different endpoint for the disco.
mux.HandleFunc("/disco/panic", func(w http.ResponseWriter, r *http.Request) {
	// If any authentication/authorization needs to happen, apply it here

	// Notice we are still using the same `asHandler`! Except in
	// this case, `myDB` will get a different IRI to load from the
	// datastore, which means this handler will serve different
	// ActivityStreams data.
	if isActivityPubRequest, err := asHandler(r.Context(), w, r); err != nil {
		return
	} else if isActivityPubRequest {
		return
	}
	http.Error("Non-ActivityPub request", http.StatusBadRequest)
	return
})
```

{{% caution %}}
Once data is served at a certain IRI, there is not a well-established way in the community to migrate that content to another IRI. Choose your HTTP paths wisely.
{{% /caution %}}

Sweet! So far we've implemented some interfaces for our demo app, and hooked them up to some HTTP handlers to handle incoming peer requests. Receiving federated is nice and all, but what about sending federated messages? All that remains in this tutorial is learning how to have our actors send out these messages!

## The Sending Side {#Sending-Side}

To send something to peers over the Fediverse, we need to first determine what it is that our actor is doing. For our application, we will have our actor send out a `Note` to tell the world a very important message. We want to tell federated peers that our actor created this `Note`, which is represented by using the `Create` activity.

The way to have Go-Fed send a federated message is to have a `FederatingActor`, and use its `Send` method. Not all `Actor` types are `FederatingActor` because ActivityPub's C2S specification doesn't have a peer-to-peer portion.

Fortunately, we already have a `FederatingActor`. On top of that, the `Send` method has a special case where if we give it something that isn't an `Activity`, it will automatically wrap it in a `Create` for us! Super convenient, since `Create` is the most commonly used activity.

Before we create a `Note`, we need our `asHandler` serving the `asHandler` endpoint to serve an actor ActivityStreams type, for peers to be able to examine. We can construct a `Person` and insert them into our in-memory database:

```go
// Create a ActivityStreams Person. Insertion into `myDB` is left as
// an exercise for the reader.
person := streams.NewActivityStreamsPerson()

// Set the `id` property of this actor, which should match
// what we are serving with `asHandler`.
id, _ := url.Parse("https://example.com/anything/me")
idProperty := streams.NewJSONLDIdProperty()
idProperty.Set(id)

// Set the `id` property on our Person.
person.SetJSONLDId(idProperty)

// Now we repeat for `inbox` and `outbox`. The IRI
// paths match the paths handled by `asHandler`.
inbox, _ := url.Parse("https://example.com/actors/me/inbox")
inboxProperty := streams.NewActivityStreamsInboxProperty()
inboxProperty.SetIRI(inbox)
person.SetActivityStreamsInbox(inboxProperty)
outbox, _ := url.Parse("https://example.com/arbitrary/me/outbox")
outboxProperty := streams.NewActivityStreamsOutboxProperty()
outboxProperty.SetIRI(outbox)
person.SetActivityStreamsInbox(outboxProperty)

// Let's set the `name` and `preferredUsername`
// properties, which are common on actors.
nameProperty := streams.NewActivityStreamsNameProperty()
nameProperty.AppendXMLSchemaString("Arr, This Be Me Name")
person.SetActivityStreamsName(nameProperty)
preferredUsernameProperty := streams.NewActivityStreamsPreferredUsernameProperty()
preferredUsernameProperty.AppendXMLSchemaString("me") 
person.SetActivityStreamsPreferredUsername(preferredUsernameProperty)

// The `followers`, `following`, `url` and `summary`
// properties are also recommended but left as an exercise for the reader.
```

Now, let's create our `Note`! Rather than sending the overly-used trope of `"Hello, World!"`, which would stand out like a sore thumb and out us as n00bs, let's send a very boring bland message to the Fediverse that either won't attract attention or get hundreds of `Announce` as an in-joke:

```go
// Obtain an ActivityStreams Note object.
func GetNote() streams.ActivityStreamsNote {
	note := streams.NewActivityStreamsNote()

	// Create the `id` property and set it -- be sure it is being served
	// by the `asHandler` (above) at the same path.
	id, _ := url.Parse("https://example.com/some/path/to/this/note")
	idProperty := streams.NewJSONLDIdProperty()
	idProperty.Set(id)

	// Set the `id` property on our Note.
	note.SetJSONLDId(idProperty)

	// Create the `content` property with a very typical Fediverse message.
	contentProperty := streams.NewActivityStreamsContentProperty()
	contentProperty.AppendXMLSchemaString("jorts")
	note.SetActivityStreamsContent(contentProperty)

	// Create the `attributedTo` property with our actor. Note that the
	// actor's IRI is the one being hosted by our `asHandler` above.
	actorIRI, _ := url.Parse("https://example.com/anything/me")
	attrToProperty := streams.NewActivityStreamsAttributedToProperty()
	attrToProperty.AppendIRI(actorIRI)
	note.SetActivityStreamsAttributedTo(attrToProperty)

	// Finally, send this `to` the public, and our actor's followers.
	followersIRI, _ := url.Parse("https://example.com/anything/me/followers")
	toProperty := streams.NewActivityStreamsToProperty()
	toProperty.AppendIRI(followersIRI)
	toProperty.AppendIRI(pub.PublicActivityPubIRI)
	note.SetActivityStreamsTo(toProperty)

	return note
}
```

Now, all that we need is to trigger this behavior:

```go
myNote := GetNote()
outboxIRI, _ := url.Parse("https://example.com/arbitrary/me/outbox")
ctx := context.Background()
// Send the note out, programmatically!
sentActivity, err := actor.Send(ctx, outboxIRI, myNote)
```

This will automatically wrap the `Note` we created within a new `Create` Activity, and send it to this particular actor's followers.

If you want to send a different Activity, or different kinds of ActivityStreams objects, or set different properties on those objects, refer to the [ActivityStreams Core Specification](https://www.w3.org/TR/activitystreams-core) and the [ActivityStreams Vocabulary](https://www.w3.org/TR/activitystreams-vocabulary). These properties and types are all represented in the `github.com/go-fed/activity/streams` and `github.com/go-fed/activity/streams/vocab` packages, following the same pattern seen in this tutorial.

## Congratulations! {#Congratulations}

Congratulations! You've created an ActivityPub demo application! It is not straightforward, and the learning curve is steep, but you've put the sweat equity in and hopefully learned a thing for two.

We saw a brief explanation how the Go-Fed interfaces work, and then stubbed them out. We also saw a brief in-memory implementation of a datastore compatible with Go-Fed. Next, we used these interfaces to get an actor. We served the actor at various HTTP endpoints, but needed more. So we used a different handler to serve non-actor endpoints. Finally, we hooked up a way to programmatically have an actor send out messages across the Fediverse.

Pat yourself on the back, that was a huge amount of work!

## Further Considerations {#Further-Considerations}

Check out the [SocialHub forum](https://socialhub.activitypub.rocks/) to talk ActivityPub in general, or to open conversations about go-fed in particular.

If you want to learn more about the vocabularies supported by Go-Fed and how to generally use the ActivityStreams vocabulary, check the [streams](https://go-fed.org/ref/activity/streams) reference page.

To get a good reference of the ActivityPub behaviors supported by the library, see the [pub](https://go-fed.org/ref/activity/pub) reference page.