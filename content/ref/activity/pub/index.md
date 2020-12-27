---
title: activity/pub Reference
summary: The `pub` package is HTTP handling middleware that implements ActivityPub C2S (Social), S2S (Federating), or both federating specs. It relies on the app to provide certain interfaces for carrying out the necessary behaviors. As an app grows, it can use less of the middleware so that the app can grow into an ActivityPub implementation of its own making, if required.
---

## One-Page Overview {#One-Page-Overview}

The `pub` package requires an app-writer to fulfill several interfaces, and in return gives functions and methods for handling HTTP requests for use within the app's HTTP Handler methods. This makes it middleware, where your application will still primarily implement the database, app-specific logic, and manage all HTTP endpoints. Yes, this means still managing all HTTP handlers for the app's usual web requests, in addition to the handlers that will delegate to this library.

The overall architecture of using the `pub` library will end up serving a HTTP request in a flow that looks like the following diagram, where at request time you provide the HTTP Requests (red) and the middleware calls back into interfaces you've implemented (white):

{{< rawhtml >}}
<div class="svg-container">
<svg height="400" width="610">
  
  <marker id="arrow" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 0 L 10 5 L 0 10 z"></path>
  </marker>
  
  <rect x="1" y="1" width="608" height="398" class="svgborder"></rect>

  <rect x="220" y="25" width="170" height="100" class="svgborder svgremotepeer"></rect>
  <text x="240" y="55" class="svgtextsmall">HTTP Request</text>
  <text x="275" y="72" class="svgtextsmaller">In/Outbox</text>
  <text x="290" y="95" class="svgtextsmaller">Data</text>
  <text x="245" y="109" class="svgtextsmaller">(Provided By Peer)</text>

  <rect x="220" y="170" width="170" height="80" class="svgactor"></rect>
  <text x="260" y="200" class="svgtextsmall">pub.Actor</text>
  <text x="262" y="220" class="svgtextsmaller">Golang Type</text>
  <text x="240" y="234" class="svgtextsmaller">(Provided By Library)</text>

  <rect x="25" y="295" width="170" height="80" class="svgborder svgprovide"></rect>
  <text x="70" y="325" class="svgtextsmall">Database</text>
  <text x="93" y="345" class="svgtextsmaller">State</text>
  <text x="55" y="359" class="svgtextsmaller">(Provided By You)</text>

  <rect x="220" y="295" width="170" height="80" class="svgborder svgprovide"></rect>
  <text x="225" y="325" class="svgtextsmall">CommonBehavior</text>
  <text x="280" y="345" class="svgtextsmaller">Behavior</text>
  <text x="255" y="359" class="svgtextsmaller">(Provided By You)</text>

  <rect x="415" y="295" width="170" height="80" class="svgborder svgprovide"></rect>
  <text x="425" y="325" class="svgtextsmall">C/S2S Behaviors</text>
  <text x="435" y="345" class="svgtextsmaller">Federation Behaviors</text>
  <text x="447" y="359" class="svgtextsmaller">(Provided By You)</text>

  <line x1="305" y1="125" x2="305" y2="164" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="305" y1="250" x2="305" y2="270" class="svgdepline"></line>
  <line x1="110" y1="270" x2="500" y2="270" class="svgdepline"></line>
  <line x1="110" y1="270" x2="110" y2="289" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="305" y1="270" x2="305" y2="289" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="500" y1="270" x2="500" y2="289" class="svgdepline" marker-end="url(#arrow)"></line>

  <rect x="25" y="25" width="170" height="100" class="svgborder svgremotepeer"></rect>
  <text x="45" y="55" class="svgtextsmall">HTTP Request</text>
  <text x="65" y="72" class="svgtextsmaller">Not In/Outbox</text>
  <text x="95" y="95" class="svgtextsmaller">Data</text>
  <text x="50" y="109" class="svgtextsmaller">(Provided By Peer)</text>

  <rect x="25" y="170" width="170" height="80" class="svgborder svgashandler"></rect>
  <text x="33" y="200" class="svgtextsmall">pub.HandlerFunc</text>
  <text x="67" y="220" class="svgtextsmaller">Golang Func</text>
  <text x="45" y="234" class="svgtextsmaller">(Provided By Library)</text>

  <line x1="110" y1="125" x2="110" y2="164" class="svgdepline" marker-end="url(#arrow)"></line>
  <line x1="65" y1="250" x2="65" y2="289" class="svgdepline" marker-end="url(#arrow)"></line>
</svg>
</div>
{{< /rawhtml >}}

Let's quickly go over the interfaces necessary for an application to implement. The list of interfaces is:

- `pub.Clock`: A simple type that abstracts the server's time, so the library is not dependent on calling `time.Now()`. Due to its simple and uninteresting nature, it is not shown in the above diagram.
- `pub.Database`: A way to provide the library access to your app's state.
- `pub.CommonBehavior`: Methods that are required by your application, regardless whether it is federating C2S, S2S, or both.
- `pub.FederatingProtocol`: Methods that are only required when doing S2S federation.
- `pub.SocialProtocol`: Methods that are only required when doing C2S federation.

Deep-dives into these interfaces are elsewhere in this tutorial. For this one-pager, we'll assume you have successfully implemented those interfaces:

```go
func main() {
	var clock pub.Clock = // ...
	var db pub.Database = // ...
	var cb pub.CommonBehavior = // ...
	var fp pub.FederatingProtocol = // ... (only for S2S)
	var sp pub.SocialProtocol = // ... (only for C2S)
```

Next is to call the appropriate kind of constructor depending on whether the federation you want to do is C2S, S2S, or both:

```go
	// Get one of the following:
	// - A S2S actor...
	s2sActor := pub.NewFederatingActor(cb, fp, db, clock)
	// - A C2S actor...
	c2sActor := pub.NewSocialActor(cb, sp, db, clock)
	// - Both C2S and S2S actor...
	bothActor := pub.NewActor(cb, sp, fp, db, clock)
```

{{%recommended%}}
You only need one actor object per behavior you'd like to have. For example, if you have an Actor that represents a user, then you just need one `pub.Actor` regardless however many users exist on the server, since their expected behaviors in your app are equal. On the other hand, an Actor that represents a bot will presumably have different software behaviors than a user, so for any number of bots a second `pub.Actor` would be needed.
{{%/recommended%}}

And those actors can be used in your HTTP handlers, here's an example handler for the actor's inbox:

```go
	http.HandleFunc("/actor/inbox",
		func(w http.ResponseWriter, r *http.Request) {
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
		}
	)
```

And similar for the outbox:

```go
	http.HandleFunc("/actor/outbox",
		func(w http.ResponseWriter, r *http.Request) {
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
		}
	)
```

{{%recommended%}}
You are responsible for making sure the handler's endpoints (`"/actor/inbox"` and `"/actor/outbox"`) are also being properly set by your app on the ActivityStreams data for `inbox` and `outbox` properties.
{{%/recommended%}}

Finally, you can also serve regular ActivityStreams data that is not related to an Actor:

```go
	apHandler := pub.NewActivityStreamsHandler(db, clock)
	http.HandleFunc("/notes",
		func(w http.ResponseWriter, r *http.Request) {
			// Do authz/authn as required, first.
			// ...

			// Next, maybe serve ActivityStreams data.
			if isActivityStreamsRequest, err := apHandler(r.Context(), w, r); err != nil {
				// Do something with `err`
				return
			} else if isActivityStreamsRequest {
				// Go-Fed handled the ActivityStreams GET request
				return
			}

			// Here we return an error, but you may just as well decide
			// to render a webpage instead.
			http.Error("Non-ActivityPub request", http.StatusBadRequest)
			return
		}
	)

	// Don't forget to kick off the web server.
	_ = http.ListenAndServe(":8080", nil)
} // End of func main
```

That's the one-page overview, to help get you in an app architecting mindset. The rest of this page goes into further detail around the Actor concept, specific interfaces, serving ActivityStreams data, and how to modify transport protocol behavior to be compatible with software such as Mastodon.

## The Actor Model {#The-Actor-Model}

ActivityPub is built atop the Actor model, which can present a challenge to developers who are used to developing web applications with a traditional REST mindset. In the ActivityPub world Actors are concepts defined within your application that are the agents that do things and enact change while the `Activity` data being exchanged describes what that change is within one instance of your application, and provides a description to federating peer applications on how they should reflect that change.

So if you send a `Create`, which is a kind of Activity, to a federated peer, you are describing to that peer two important details:

1. What object is being created; and
1. Who is doing the creating

Likewise, when your application receives a similar activity, it is being told the same details. The difference with an Actor model mindset is that messages are ONLY ever passed between these Actors; they are the only entities that can change the state of the world. ActivityPub models these Actors as having two endpoints:

1. An Inbox for receiving other Actor messages.
1. An Outbox for recording the messages sent to other Actors.

That's it.

This is a pretty open-ended and vague capability, so the real detail comes down to the specific behaviors you want to model and the specific software you want to federate with. Writing your application to federate with itself is easy, but figuring out what other applications require is hard. There are a lot of different `Activity` types out there to describe these interactions, and it is not a closed set either!

So when building a new application or retrofitting ActivityPub into an existing application, meaningful questions include:

- What kind of actors do I want to have? An actor for a user is typical, as is an actor representing the server itself.
- What kind of Activities will they send? Create, Update, and Delete are common.
- What kind of Objects will my software manipulate? Notes and Articles are currently the most prevalent

The answers to these questions will help guide how to implement the rest of these interfaces and use the `pub` library.

## The Database Interface {#The-Database-Interface}

The `pub.Database` interface is how you provide stateful information to the go-fed library. It gives your app the flexibility to use whatever implementation you wish to use or are already using.

Abstracting away a datastore is not easy, so this interface follows several principles:

1. The database is IRI-centric. It treats `*url.URL` parameters as a primary key.
1. The `Lock` and `Unlock` methods will be appropriately called for a particular IRI.
1. It does not handle concepts like transactions out of the box. However, since it follows the `context.Context` pattern, it is possible for your particular implementation to still use concepts like transactions.
1. It does not handle concepts like caching out of the box. However, some APIs are designed to work with pages of a `Collection`, rather than the entire `Collection`, to facilitate implementations that want to develop their own caching capability.

With that in mind, let's dive into the interface.

### Lock {#Database-Lock}

```go
Lock(c context.Context, id *url.URL) error
```

The `Lock` and `Unlock` methods let your database implementation ensure asynchronous requests perform atomic changes to the underlying data. The `id` parameter acts as the primary key for the ActivityStreams entity that is going to be retrieved for either a read-only or read-write use-case.

An implementation may decide to manage a dictionary of mutexes and lock a specific one, do nothing and instead rely on a particular database's transaction model, or something else.

### Unlock {#Database-Unlock}

```go
Unlock(c context.Context, id *url.URL) error
```

The `Unlock` method is the counterpart to the `Lock` method. The `id` parameter acts as the primary key for the ActivityStreams entity that has already been retrieved, and possibly stored, for either a read-only or read-write use-case.

### Get {#Database-Get}

```go
Get(c context.Context, id *url.URL) (value vocab.Type, err error)
```

Simply fetch the ActivityStreams object with `id` from the database. The `streams.ToType` function can turn any arbitrary JSON-LD literal into a `vocab.Type` for `value`.

### Create {#Database-Create}

```go
Create(c context.Context, asType vocab.Type) error
```

Store the arbitrary ActivityStreams `asType` object into the database. It should be uniquely new to the database when examining its `id` property, and shouldn't overwrite any existing data.

{{%recommended%}}
If needed, use `streams.Serialize` to turn the `vocab.Type` into literal JSON-LD bytes.
{{%/recommended%}}

### Update {#Database-Update}

```go
Update(c context.Context, asType vocab.Type) error
```

This is the same as `Create` except it is expected that the object already is in the database. The entity with the same `id` should be overwritten by the provided value. You do not need to worry about the ActivityPub specification talking about whether an Update means a partial-update or complete-replacement, as the library has already done this for you, so it is safe to simply replace the row.

### Delete {#Database-Delete}

```go
Delete(c context.Context, id *url.URL) error
```

Simply remove the entity or row with the matching `id`.

### Exists {#Database-Exists}

```go
Exists(c context.Context, id *url.URL) (exists bool, err error)
```

Set `exists` to `true` if the database has an entity or row with the `id`.

### NewID {#Database-NewID}

```go
NewID(c context.Context, t vocab.Type) (id *url.URL, err error)
```

The library is in the process of creating a new ActivityStreams payload, and is calling this method to allocate a new IRI. You can inspect the context or the value, such as its type, in order to properly allocate an IRI meaningful to your application.

{{%caution%}}
Ensure that the newly allocated IRI can properly be fetched in another web handler by peers with proper authorization and authentication, which can be aided with `pub.HandlerFunc`.
{{%/caution%}}

### InboxContains {#Database-InboxContains}

```go
InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error)
```

Given the IRI of an `inbox`, the implemented method should set `contains` to `true` if the ActivityStreams object with the `id` is contained within that Inbox's OrderedCollection.

A naive implementation may just do a linear search through the OrderedCollection, while certain databases may permit better lookup performance with a proper query.

### GetInbox {#Database-GetInbox}

```go
GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error)
```

This method returns the latest page of the `inbox` corresponding to the `inboxIRI`.

At first glance this method seems a little odd. It is fine to return an empty `vocab.ActivityStreamsOrderedCollectionPage`. The library expects the very first page, which is the most recent chronologically. Therefore, an empty page is always treated as the "first zero" items, and the library does not require having any items. If you have a caching layer, it can more easily hide under this method with proper pagination and delayed writes to the database. The library is simply going to prepend an item in the `orderedItems` property and then call `SetInbox`.

### SetInbox {#Database-SetInbox}

```go
SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error
```

This method accepts a modified `vocab.ActivityStreamsOrderedCollectionPage` that had been returned by `GetInbox`. Right now the library only prepends new items to the `orderedItems` property, so simple diffing can be done. This method should then modify the actual underlying `inbox` to reflect the change in this page.

### GetOutbox {#Database-GetOutbox}

```go
GetOutbox(c context.Context, outboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error)
```

This method returns the latest page of the `inbox` corresponding to the `outboxIRI`.

It is similar in behavior to its `GetInbox` counterpart, but for the actor's Outbox instead.

See the similar documentation for [`GetInbox`](#Database-GetInbox).

### SetOutbox {#Database-SetOutbox}

```go
SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error
```

This method accepts a modified `vocab.ActivityStreamsOrderedCollectionPage` that had been returned by `GetOutbox` to update the underlying `outbox`.

It is similar in behavior to its `SetInbox` counterpart, but for the actor's Outbox instead.

See the similar documentation for [`SetInbox`](#Database-SetInbox).

### Owns {#Database-Owns}

```go
Owns(c context.Context, id *url.URL) (owns bool, err error)
```

Sets `owns` to `true` when the id is an IRI owned by this running instance of the server. That is, the data represented by the id did not come from a federated peer.

### ActorForOutbox {#Database-ActorForOutbox}

```go
ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error)
```

Given the `outboxIRI`, the associated `actorIRI` which is the Actor's id.

This will only be called with `outboxIRI` whose actors are owned by this instance.

### ActorForInbox {#Database-ActorForInbox}

Given the `inboxIRI`, the associated `actorIRI` which is the Actor's id.

This will only be called with `inboxIRI` whose actors are owned by this instance.

### OutboxForInbox {#Database-OutboxForInbox}

```go
OutboxForInbox(c context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error)
```

Given the `inboxIRI`, the associated `outboxIRI` which is the Actor's Outbox id.

This will only be called with `inboxIRI` whose actors are owned by this instance.

### Followers {#Database-Followers}

```go
Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error)
```

Given the `actorIRI`, which is the Actor's id, returns that Actor's followers Collection. This must be the complete collection of followers for that Actor.

### Following {#Database-Following}

```go
Following(c context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error)
```

Given the `actorIRI`, which is the Actor's id, returns that Actor's following Collection. This must be the complete collection of Actors they are following for that Actor.

### Liked {#Database-Liked}

```go
Liked(c context.Context, actorIRI *url.URL) (liked vocab.ActivityStreamsCollection, err error)
```

Given the `actorIRI`, which is the Actor's id, returns that Actor's liked Collection. This must be the complete collection of liked objects for that Actor.

## The CommonBehavior Interface {#The-CommonBehavior-Interface}

The `pub.CommonBehavior` interface is needed regardless whether your application wishes to do S2S, C2S, or both ActivityPub protocols.

These methods are best to include on a type responsible for web or application behavior, rather than a type responsible for data.

### AuthenticateGetInbox {#CommonBehavior-AuthenticateGetInbox}

```go
AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
```

Determines whether the request is for a GET call to the Actor's Inbox. The `out` Context is used in further library calls, so your app's behavior can be modified depending on the authenticated context, such as whether to serve private messages.

If an error is returned, it is passed back to the caller of GetInbox. In this case, the implementation must not write a response to the `http.ResponseWriter` as is expected that the client will do so when handling the error. The `authenticated` is ignored.

If no error is returned, but authentication or authorization fails, then `authenticated` must be `false` and `error` `nil`. It is expected that the implementation handles writing to the `http.ResponseWriter` in this case.

Finally, if the authentication and authorization succeeds, then `authenticated` must be `true` and `error` `nil`. The request will continue to be processed.

### AuthenticateGetOutbox {#CommonBehavior-AuthenticateGetOutbox}

```go
AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
```

Determines whether the request is for a GET call to the Actor's Outbox. The `out` Context is used in further library calls, so your app's behavior can be modified depending on the authenticated context, such as whether to serve private messages.

If an error is returned, it is passed back to the caller of GetOutbox. In this case, the implementation must not write a response to the `http.ResponseWriter` as is expected that the client will do so when handling the error. The `authenticated` is ignored.

If no error is returned, but authentication or authorization fails, then `authenticated` must be `false` and `error` `nil`. It is expected that the implementation handles writing to the `http.ResponseWriter` in this case.

Finally, if the authentication and authorization succeeds, then `authenticated` must be `true` and `error` `nil`. The request will continue to be processed.

### GetOutbox {#CommonBehavior-GetOutbox}

```go
GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error)
```

Returns a proper paginated view of the Outbox for serving in a response. Since `AuthenticateGetOutbox` is called before this, the implementation is responsible for ensuring things like proper pagination, visible content based on permissions, and whether to leverage the `pub.Database`'s GetOutbox method in this implementation.

### NewTransport {#CommonBehavior-NewTransport}

```go
NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t Transport, err error)
```

Returns a new `pub.Transport` for federating with peer software. There is a `pub.HttpSigTransport` implementation provided for using HTTP and HTTP Signatures, but providing a different transport allows federating using different protocols.

The `actorBoxIRI` will be either the Inbox or Outbox of an Actor who is attempting to do the dereferencing or delivery. Any authentication scheme applied on the request must be based on this actor. The request must contain some sort of credential of the user, such as a HTTP Signature.

The `gofedAgent` passed in should be used by the `pub.Transport` implementation in the User-Agent, as well as the application-specific user agent string. The `gofedAgent` will indicate this library's use as well as the library's version number.

Any server-wide rate-limiting that needs to occur should happen in a `pub.Transport` implementation. This factory function allows this to be created, so peer servers are not DOS'd.

Any retry logic should also be handled by the `pub.Transport` implementation.

Note that the library will not maintain a long-lived pointer to the returned `pub.Transport` so that any private credentials are able to be garbage collected.

For more information, see the [Transports](#Transports) section below.

## The FederatingProtocol Interface {#The-FederatingProtocol-Interface}

The `pub.FederatingProtocol` is only needed if an application wants to do the S2S (Server-to-server, or federating) ActivityPub protocol. It supplements the `pub.CommonBehavior` interface with the additional methods required by a federating application.

### PostInboxRequestBodyHook {#FederatingProtocol-PostInboxRequestBodyHook}

```go
PostInboxRequestBodyHook(c context.Context, r *http.Request, activity Activity) (context.Context, error)
```

This is a hook that occurs after reading the request body of a POST request to an Actor's Inbox.

Provides your application the opportunity to set contextual information based on the incoming `http.Request` and its body. Some applications simply return `c` and do nothing else, which is OK. More commonly, software simply inspects the `http.Request` path to determine the actual local Actor being interacted with, and save such information within `c`.

Any errors returned immediately abort processing of the request and are returned to the caller of the Actor's `PostInbox`.

{{%caution%}}
Do not do anything sensitive in this method. Neither authorization nor authentication has been attempted at the point `PostInboxRequestBodyHook` has been called.
{{%/caution%}}

### AuthenticatePostInbox {#FederatingProtocol-AuthenticatePostInbox}

```go
AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
```

This is a callback for your application to determine whether the incoming `http.Request` is `authenticated` and, implicitly, authorized to proceed with processing the request.

If an error is returned, it is passed back to the caller of `PostInbox`. In this case, the implementation must not write a response to the `http.ResponseWriter` as is expected that the client will do so when handling the error. The `authenticated` value is ignored in this case.

If no error is returned, but your application determines that authentication or authorization fails, then `authenticated` must be `false` and `err` `nil`. It is expected that the implementation handles writing to the `http.ResponseWriter` in this case.

Finally, if the authentication and authorization succeeds, then `authenticated` must be `true` and `err` `nil`. The request will continue to be processed.

### Blocked {#FederatingProtocol-Blocked}

```go
Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error)
```

Given a list of `actorIRIs`, determines whether any are `blocked` for this particular request context and based on the particular application's state. For example, some applications allow users or software instances to maintain lists of blocked peer Actors or domains.

To determine the current user being interacted with, it is recommended to set such information in the `PostInboxRequestBodyHook` method.

If an error is returned, it is passed back to the caller of PostInbox.

If no error is returned, but the interaction should be blocked, then `blocked` must be `true` and `err` `nil`. An `http.StatusForbidden` will be written in the response.

Finally, the interaction should proceed, then `blocked` must be `false` and `err` `nil`. The request will continue to be processed.

### FederatingCallbacks {#FederatingProtocol-FederatingCallbacks}

```go
FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error)
```

Returns the application-specific logic needed for your application as callbacks for the library to invoke.

The library splits your applications behaviors between those specified in the ActivityPub spec, which will wrap your behaviors in `wrapped`, or behaviors not known in the ActivityPub spec which will be provided in `other`.

The `pub.FederatingWrappedCallbacks` returned provides a collection of default ActivityPub behaviors as defined in the specification. For more details on how to use these provided behaviors and supplement with your own business logic, see [Federating Wrapped Callbacks](#Federating-Wrapped-Callbacks). The zero-value is a valid value.

If instead you wish to override the default ActivityPub behaviors, such as doing nothing, then the `other` return value should contain a function with a signature like:

```go
other = []interface{}{
	// This function overrides the FederatingWrappedCallbacks-provided behavior
	func(c context.Context, create vocab.ActivityStreamsCreate) error {
		return nil
	},
}
```

The above would replace the library's default behavior of creating the entry in the database upon receiving a Create activity.

If you want to handle an Activity that does not have a default behavior provided in `pub.FederatingWrappedCallbacks`, then specify it in `other` using a similar function signature.

Applications are not expected to handle every single ActivityStreams type and extension. The unhandled ones are passed to `DefaultCallback`.

### DefaultCallback {#FederatingProtocol-DefaultCallback}

```go
DefaultCallback(c context.Context, activity Activity) error
```

This method is called for types that the library can deserialize but is not handled by the application's callbacks returned in the `FederatingCallbacks` method.

### MaxInboxForwardingRecursionDepth {#FederatingProtocol-MaxInboxForwardingRecursionDepth}

```go
MaxInboxForwardingRecursionDepth(c context.Context) int
```

MaxInboxForwardingRecursionDepth determines how deep to search within an activity's historical chain to determine if inbox forwarding needs to occur. After reaching this depth, it is assumed that peers deeper than that conversational depth are no longer candidates for triggering the inbox forwarding logic.

{{%caution%}}
Zero or negative numbers indicate recurring infinitely, which can result in your application being manipulated by malicious peers. Do not return a value of zero nor a negative number.
{{%/caution%}}

### MaxDeliveryRecursionDepth {#FederatingProtocol-MaxDeliveryRecursionDepth}

```go
MaxDeliveryRecursionDepth(c context.Context) int
```

This method determines how deep to search within collections owned by peers when they are targeted to receive a delivery. After reaching this depth, it is assumed that peers deeper than that are no longer interested in receiving messages. A positive number must be returned.

{{%caution%}}
Zero or negative numbers indicate recurring infinitely, which can result in your application being manipulated by malicious peers. Do not return a value of zero nor a negative number.
{{%/caution%}}

### FilterForwarding {#FederatingProtocol-FilterForwarding}

```go
FilterForwarding(c context.Context, potentialRecipients []*url.URL, a Activity) (filteredRecipients []*url.URL, err error)
```

Allows the implementation to apply outbound message business logic such as blocks, spam filtering, and so on to a list of `potentialRecipients` when inbox forwarding has been triggered. Your application **must** apply some sort of filtering, such as limiting delivery to an actor's followers. Otherwise, your application will become a vector for spam on behalf of malicious peers, and users of your software will be mass-blocked by their peers.

{{%caution%}}
The activity is provided as a reference for more intelligent logic to be used, but the implementation **must not** modify the activity.
{{%/caution%}}

### GetInbox {#FederatingProtocol-GetInbox}

```go
GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error)
```

Returns a proper paginated view of the Inbox for serving in a response. Since `AuthenticateGetInbox` is called before this, the implementation is responsible for ensuring things like proper pagination, visible content based on permissions, and whether to leverage the `pub.Database`'s GetInbox method in this implementation.

## The SocialProtocol Interface {#The-SocialProtocol-Interface}

The `pub.SocialProtocol` is only needed if an application wants to do the C2S (Client-to-server, or social) ActivityPub protocol. It supplements the `pub.CommonBehavior` interface with the additional methods required by a social application.

### PostOutboxRequestBodyHook {#SocialProtocol-PostOutboxRequestBodyHook}

```go
PostOutboxRequestBodyHook(c context.Context, r *http.Request, data vocab.Type) (context.Context, error)
```

This is a hook that occurs after reading the request body of a POST request to an Actor's Outbox.

Provides your application the opportunity to set contextual information based on the incoming `http.Request` and its body. Some applications simply return c and do nothing else, which is OK. More commonly, software simply inspects the `http.Request` path to determine the actual local Actor being interacted with, and save such information within `c`.

Any errors returned immediately abort processing of the request and are returned to the caller of the Actor's `PostOutbox`.

{{%caution%}}
Do not do anything sensitive in this method. Neither authorization nor authentication has been attempted at the point `PostOutboxRequestBodyHook` has been called.
{{%/caution%}}

### AuthenticatePostOutbox {#SocialProtocol-AuthenticatePostOutbox}

```go
AuthenticatePostOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
```

This is a callback for your application to determine whether the incoming `http.Request` is `authenticated` and, implicitly, authorized to proceed with processing the request.

If an error is returned, it is passed back to the caller of `PostOutbox`. In this case, the implementation must not write a response to the `http.ResponseWriter` as is expected that the client will do so when handling the error. The `authenticated` value is ignored in this case.

If no error is returned, but your application determines that authentication or authorization fails, then `authenticated` must be `false` and `err` `nil`. It is expected that the implementation handles writing to the `http.ResponseWriter` in this case.

Finally, if the authentication and authorization succeeds, then `authenticated` must be `true` and `err` `nil`. The request will continue to be processed.

### SocialCallbacks {#SocialProtocol-SocialCallbacks}

```go
SocialCallbacks(c context.Context) (wrapped SocialWrappedCallbacks, other []interface{}, err error)
```

Returns the application-specific logic needed for your application as callbacks for the library to invoke.

The library splits your applications behaviors between those specified in the ActivityPub spec, which will wrap your behaviors in `wrapped`, or behaviors not known in the ActivityPub spec which will be provided in `other`.

The `pub.SocialWrappedCallbacks` returned provides a collection of default ActivityPub behaviors as defined in the specification. For more details on how to use these provided behaviors and supplement with your own business logic, see [Social Wrapped Callbacks](#Social-Wrapped-Callbacks). The zero-value is a valid value.

If instead you wish to override the default ActivityPub behaviors, such as doing nothing, then the `other` return value should contain a function with a signature like:

```go
other = []interface{}{
	// This function overrides the SocialWrappedCallbacks-provided behavior
	func(c context.Context, create vocab.ActivityStreamsCreate) error {
		return nil
	},
}
```

The above would replace the library's default behavior of creating the entry in the database upon receiving a Create activity.

If you want to handle an Activity that does not have a default behavior provided in `pub.SocialWrappedCallbacks`, then specify it in `other` using a similar function signature.

Applications are not expected to handle every single ActivityStreams type and extension. The unhandled ones are passed to `DefaultCallback`.

### DefaultCallback {#SocialProtocol-DefaultCallback}

```go
DefaultCallback(c context.Context, activity Activity) error
```

This method is called for types that the library can deserialize but is not handled by the application's callbacks returned in the `SocialCallbacks` method.

## Serving ActivityStreams Data {#Serving-ActivityStreams-Data}

Serving ActivityStreams data is done by using a `pub.HandlerFunc` within a typical `http` handler or handler function. Your application is responsible for managing the incoming request's authentication and authorization if required at the endpoint.

For example, let's suppose you want to serve all notes under the `"/notes"` path, and all notes are publically available. If the note isn't an ActivityStreams request, then the app will serve a webpage instead. Then a sample implementation is simply:

```go
apHandler := pub.NewActivityStreamsHandler(db, clock)
http.HandleFunc("/notes",
	func(w http.ResponseWriter, r *http.Request) {
		// Maybe serve ActivityStreams data.
		if isActivityStreamsRequest, err := apHandler(r.Context(), w, r); err != nil {
			// Do something with `err`
			return
		} else if isActivityStreamsRequest {
			// If it was an ActivityStreams request, return
			return
		}
		// If it was not an ActivityStreams request, render the webpage here
		return
	}
)
```

The `apHandler` is reusable so long as it does not need to fetch from a different `pub.Database` or `pub.Clock`. This is because it will extract the appropriate path from the `http.Request` to fetch the associated ActivityStreams data. In the example above, the handler will be called for `"/notes/1"` and will call the database for the first note. It will also serve `"/notes/5"` and serve the fifth note. The handler is equally capable of serving `"/articles/1"` if you put it in the appropriate handler, because it is stateless function: your database provides its state.

It is very important to serve ActivityStreams data through this handler and not attempt to serve it on your own, unless you have thorough understanding of the specification. Unlike other portions of the library, the handlers do not allow applications to override its behavior because non-compliance has privacy implications.

## Serving ActivityStreams Data by Federating {#Serving-ActivityStreams-Data-by-Federating}

In order to programmatically send data to a federating peer, your application must support the S2S protocol by calling either `pub.NewFederatingActor` (S2S only) or `pub.NewActor` (C2S and S2S). These constructors return a `pub.FederatingActor` instead of an `pub.Actor`. The `pub.FederatingActor` can be used to programmatically `Send` an ActivityStreams Activity to a peer:

```go
var actor pub.FederatingActor = // ...
var myOutbox *url.URL = // ...
var myMessage vocab.Type = // ...
sentActivity, err := actor.Send(context.Background(), myOutbox, myMessage)
```

The `myOutbox` path determines which Actor is the one delivering the Activity to a peer. This is important because it must also match the originating actor on `myMessage`, since the delivery addressing is reflected within the data itself, wherever your application creates its ActivityStreams data:

```go
// This actor IRI must correspond to 'myOutbox'!
actorIRI, _ := url.Parse(/* Actor IRI for 'myOutbox' */)

// Set the actor on the 'attributedTo' property
attrTo := streams.NewActivityStreamsAttributedToProperty()
attrTo.AppendIRI(actorIRI)

// Set the 'attributedTo' property on, for
// example,  an Article to share
article := streams.NewActivityStreamsArticle()
article.SetActivityStreamsAttributedTo(attrTo)
```

Note that the `Send` method is versatile. Much like the C2S protocol, it can accept an ActivityStreams Object instead of an Activity. It will wrap the Object with a Create Activity in this case. This is why the `Send` method returns a `pub.Activity`, as it is returning a copy of the Activity actually sent to peers, but with sensitive fields present.

In any case, it will acquire new identifiers to set on the Activities and/or Objects missing the needed identifiers.

Finally, the Activity will be prepared for delivery, which may change the addressing properties on the Activity and/or any objects contained within.

## Transports {#Transports}

Transports act as an abstraction from the underlying HTTP protocol. It permits future non-HTTP protocols to also be abstracted away in a similar manner. Furthermore, different Transports can provide different methods of sending bytes to peers. For example, the library provides a `pub.HttpSigTransport` which supports the HTTP Signatures specification when making connections.

A `pub.Transport` is intended to be a short-lived client, as it may contain authentication information that is not intended to be long-lived.

An application using `go-fed/activity` is expected to provide a `pub.Transport` in the S2S `NewTransport` method. That means it is up to your application to do any requisite key management, such as when constructing a `pub.HttpSigTransport`:

```go
func (*myService) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestPref := httpsig.DigestSha256
	getHeadersToSign := []string{httpsig.RequestTarget, "Date"}
	postHeadersToSign := []string{httpsig.RequestTarget, "Date", "Digest"}
	// Using github.com/go-fed/httpsig for HTTP Signatures:
	getSigner, _, err := httpsig.NewSigner(prefs, digestPref, getHeadersToSign, httpsig.Signature, 3600)
	postSigner, _, err := httpsig.NewSigner(prefs, digestPref, postHeadersToSign, httpsig.Signature, 3600)
	pubKeyId, privKey, err := s.getKeysForActorBoxIRI(actorBoxIRI)
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	t = pub.NewHttpSigTransport(
		client,
		"example.com",
		&myClock{},
		getSigner,
		postSigner,
		pubKeyId,
		privKey)
	return
}
```

## Utilities {#Utilities}

The `pub` library provides a few additional functions that are handy utilities.

`pub.ToId` obtains the IRI of the ID for any kind of functional property or non-functional property iterator:

```go
var next vocab.ActivityStreamsNextProperty = // ...
nextIRI, err := pub.ToId(next)

var objectIterator vocab.ActivityStreamsObjectPropertyIterator = // ...
objectPropertyIRI, err := pub.ToId(objectIterator)
```

Similarly, `pub.GetId` helps get the IRI of the ID for any kind of ActivityStreams type:

```go
var create vocab.ActivityStreamsCreate = // ...
createIRI, err := pub.GetId(create)

```

## Federating Wrapped Callbacks {#Federating-Wrapped-Callbacks}

The `pub.FederatingWrappedCallbacks` type embodies a set of basic behaviors outlined in the ActivityPub specification for S2S federation. These behaviors can be supplemented by your own application's logic by providing the analogous functions when constructing an instance of `pub.FederatingWrappedCallbacks`:

```go
f := pub.FederatingWrappedCallbacks {
	// Example supplemental "Create" Activity app behavior
	Create: func(c context.Context, create vocab.ActivityStreamsCreate) error {
		myAppBehavior() // Etc...
	},
}
```

When handing an Activity, it will always attempt to store it in the database. The ActivityPub specification outlines additional suggested behaviors. The Activities that have default behaviors that are supported and can be supplemented by your application's logic are:


- `Create`: create objects in the Database.
- `Update`: update objects in the Database.
- `Delete`: delete objects from the Database.
- `Follow`: do behavior based on the `OnFollow` value set in the `pub.FederatingWrappedCallbacks`. By default it is `pub.OnFollowDoNothing` which simply adds the Follow to the Database. A value of `pub.OnFollowAutomaticallyAccept` will automatically send an Accept in response to the Follow request. A value of `pub.OnFollowAutomaticallyReject` will automatically send a Reject in response to the Follow request.
- `Accept`: if in response to a Follow, will properly add the appropriate Actor to the proper Following collection.
- `Reject`: no notable behaviors.
- `Add`: if the target is a Collection or OrderedCollection, adds the object to such a collection.
- `Remove`: if the target is a Collection or OrderedCollection, removes the object from such a collection.
- `Like`: adds the Like IRI to the list of IDs in the `likes` collection.
- `Announce`: adds the Announce IRI to the list of IDs in the `shares` collection.
- `Undo`: no notable behaviors.
- `Block`: no notable behaviors. It is technically a violation of the ActivityPub specification to federate Block Activities.

## Social Wrapped Callbacks {#Social-Wrapped-Callbacks}

The `pub.SocialWrappedCallbacks` type embodies a set of basic behaviors outlined in the ActivityPub specification for C2S federation. These behaviors can be supplemented by your own application's logic by providing the analogous functions when constructing an instance of `pub.SocialWrappedCallbacks`:

```go
s := pub.SocialWrappedCallbacks {
	// Example supplemental "Create" Activity app behavior
	Create: func(c context.Context, create vocab.ActivityStreamsCreate) error {
		myAppBehavior() // Etc...
	},
}
```

When handing an Activity, it will always attempt to store it in the database. The ActivityPub specification outlines additional suggested behaviors. The Activities that have default behaviors that are supported and can be supplemented by your application's logic are:


- `Create`: ensures that the `attributedTo`, `to`, `bto`, `cc`, `bcc`, `audience` properties are normalized correctly before being persisted in the database.
- `Update`: applies a partial diff to the existing ActivityStreams data before updating in the database.
- `Delete`: deletes data from the database.
- `Follow:` no notable behaviors.
- `Add`: if the target is a Collection or OrderedCollection, adds the object to such a collection.
- `Remove`: if the target is a Collection or OrderedCollection, removes the object from such a collection.
- `Like`: adds the object IDs to the actor's `liked` collection.
- `Undo`: no notable behaviors.
- `Block`: no notable behaviors.
