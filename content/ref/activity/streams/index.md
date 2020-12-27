---
title: activity/streams Reference
summary: The `streams` package is the building block of the ActivityStreams vocabularies supported in Go. It is mostly code-generated so that the code at compile and run times does not need JSON-LD understanding to function. Because it is code-generated, learning to use this library's patterns means you can still remain proficient even when new vocabularies are added!
---

## One-Page Overview {#One-Page-Overview}

Let's dive right into the basics of creating a piece of ActivityStreams data:

```go
import (
  "github.com/go-fed/activity/streams"
  "github.com/go-fed/activity/streams/vocab"
)

// Create a Note object
var note vocab.ActivityStreamsNote = streams.NewActivityStreamsNote()

// Create an `id` property and set it on the Note
id, _ := url.Parse("https://example.com/some/path/to/this/note")
var idProperty vocab.JSONLDIdProperty = streams.NewJSONLDIdProperty()
idProperty.Set(id)

// Set the `id` property on our Note.
note.SetJSONLDId(idProperty)
```
Be sure you're taking note! Ha, punny.

Vocabularies will define types and properties. Both vocabulary types and vocabulary properties are represented in Go as concrete types. In the example above the `vocab.ActivityStreamsNote` vocabulary type is a Go type, and the `vocab.JSONLDIdProperty` vocabulary property is also a Go type. That means to set the value `"https://example.com/some/path/to/this/note"` on the id property, it is a little cumbersome because the value must be set on a property first, and then the property needs to be set on a type.

You may have noticed the pattern used when naming the Go types. The is `<Vocabulary Name><Type Name>` for a vocabulary's types and `<Vocabulary Name><Property Name>Property` for a vocabulary's properties.

Processing an existing note also has its considerations:

```go
// Let's try to add content to our note. First, let's get the property.
contentProperty := note.GetActivityStreamsContent()

// WARNING: Missing properties are `nil`!
if contentProperty == nil {
	// Create a new property and set it on the note.
	contentProperty = streams.NewActivityStreamsContentProperty()
	// Treat properties as pointers, not values. Setting a
	// property is not a value-copy so if we modify
	// the property later, any modification will be
	// reflected in the note.
	note.SetActivityStreamsContent(contentProperty)
}

// Now we are guaranteed a non-`nil` property: let's add content!
contentProperty.AppendXMLSchemaString("jorts")
```

{{% caution %}}
When getting a property, always check whether it is nil!

Allocating zero values is not cheap, so the library avoids it as much as possible.
{{% /caution %}}

In the example above, we first get the property. It may be nil when it didn't exist in the JSON payload or hasn't yet been set on a new type. So we created a new property in that case, which can then be inspected or modified in later code.

Some properties are "functional" and can only hold one value or one IRI:

```go
// The "published" property is functional: It can only have at most one value.
published := streams.NewActivityStreamsPublishedProperty()
// We can set a time...
published.Set(time.Now())
// ...or, in this very unusual practice, set it as an IRI
iri, _ := url.Parse("https://go-fed.org/some/path")
published.SetIRI(iri)

if published.IsIRI() {
  fmt.Println(published.GetIRI())
} else if published.IsXMLSchemaDateTime() {
  fmt.Println(published.Get())
}
```

Some properties are "non-functional" and can only hold multiple values, IRIs, or a mix:

```go
// The "object" property is non-functional: It can have many values.
object := streams.NewActivityStreamsObjectProperty()
// We can append...
object.AppendActivityStreamsNote(note)
// ...or prepend...
object.PrependActivityStreamsArticle(streams.NewActivityStreamsArticle())
// ...and IRIs too
iri, _ := url.Parse("https://go-fed.org/foo")
object.AppendIRI(iri)

// An iterator interface lets you work with each element
for iter := object.Begin(); iter != iter.End(); iter = iter.Next() {
  if iter.IsActivityStreamsNote() {
    note := iter.GetActivityStreamsNote()
  } else if published.IsActivityStreamsArticle() {
    article := iter.GetActivityStreamsArticle()
  } else if published.IsIRI() {
    iri := iter.GetIRI()
  }
}
```

These properties let you build up and process a container of all sorts of differently-typed elements. A naive Go slice is not good enough without resorting to `interface{}` and type switches, and this hides away that boilerplate from you.

Finally, there's the question of how to serialize to and from JSON:

```go
// Deserialize a JSON payload
jsonstr := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id":       "https://go-fed.org/foo",
  "name":     "Foo Bar",
  "inbox":    "https://go-fed.org/foo/inbox",
  "outbox":   "https://go-fed.org/foo/outbox",
  "type":     "Person",
  "url":      "https://go-fed.org/foo"
}`
var m map[string]interface{}
_ = json.Unmarshal([]byte(jsonstr), &m)

// Next, we prepare a streams.JSONResolver, providing one or more callbacks.
var person vocab.ActivityStreamsPerson
resolver := streams.NewJSONResolver(func(c context.Context, p vocab.ActivityStreamsPerson) error {
  // Store the person in the enclosing scope, for later.
  person = p
  return nil
}, func(c context.Context, note vocab.ActivityStreamsNote) error {
  // We can treat the type differently.
  fmt.Println(note)
  return nil
})
// It will call back a function we provide if it is of a matching type,
// or returns streams.ErrNoCallbackMatch when we didn't give it a matcher for
// the type, or streams.ErrUnhandledType if it is a type unknown to Go-Fed.
ctx := context.Background()
err := resolver.Resolve(ctx, m)

// Serialize to a JSON payload
var jsonmap map[string]interface{}
jsonmap, _ = streams.Serialize(person) // WARNING: Do not call the Serialize() method on person
b, _ := json.Marshal(jsonmap)
```

{{% caution %}}
Do not call the `Serialize()` method on types directly, use the `streams.Serialize` function!

This ensures your JSON-LD `@context field` is properly set.
{{% /caution %}}

Hopefully these short examples gets you started with the ActivityStreams vocabulary! The rest of the documentation addresses the finer points of using the streams API.

## Working With Types {#Working-With-Types}

The [One-Page Overview](#One-Page-Overview) went over how to create new types, get and set their properties. This covers a lot of use-cases, but the `streams` package does more.

The vocabulary typing hierarchy is impossible to accurately represent using traditional object-oriented programming and notions of object hierarchies, but is well-suited to Go's interface duck-typing. If you wish to operate on a type in general:

```go
// vocab.Type represents an ActivityStreams type in general
var aType vocab.Type = streams.NewActivityStreamsCollection()
// ...which can be used for serializing...
jsonmap, _ := streams.Serialize(aType)
b, _ := json.Marshal(jsonmap)
```

Which, uh, seems pretty boring. But!

```go
// ...or branching logic based on its precise type...
resolver := streams.NewTypeResolver(func(c context.Context, oc vocab.ActivityStreamsOrderedCollection) error {
  fmt.Println(oc)
  return nil
}, func(c context.Context, c vocab.ActivityStreamsCollection) error {
  fmt.Println(c)
  return nil
})
// This is a TypeResolver, not a JSONResolver, so it accepts a vocab.Type
// instead of a map[string]interface{}.
ctx := context.Background()
err := resolver.Resolve(ctx, aType)
```

But... why are you yawning? Oh gee!

```go
// ...or expertly inspecting the type's inheritance within and across vocabularies.
object := streams.NewActivityStreamsObject()
activity := streams.NewActivityStreamsActivity()
repo := streams.NewForgeFedRepository()

// Determining if a Type has a parent of type Object from the ActivityStreams vocabulary.
if streams.ActivityStreamsObjectIsExtendedBy(object) {
  fmt.Println("I'm false, since a type doesn't extend itself.")
} else if streams.ActivityStreamsObjectIsExtendedBy(activity) {
  fmt.Println("I'm true.")
} else if streams.ActivityStreamsObjectIsExtendedBy(repo) {
  fmt.Println("I'm true, too!")
}

// Determining if a ForgeFed Repository extends from a Type.
if streams.ForgeFedRepositoryExtends(object) {
  fmt.Println("I'm true!")
} else if streams.ForgeFedRepositoryExtends(activity) {
  fmt.Println("I'm false.")
} else if streams.ForgeFedRepositoryExtends(repo) {
  fmt.Println("I'm false, too! A type doesn't extend from itself.")
}

// Determining if a ForgeFed Repository is the Type or extends from a Type.
if streams.IsOrExtendsForgeFedRepository(object) {
  fmt.Println("I'm true!")
} else if streams.IsOrExtendsForgeFedRepository(activity) {
  fmt.Println("I'm false.")
} else if streams.IsOrExtendsForgeFedRepository(repo) {
  fmt.Println("I'm true now, since it is the same type!")
}

// Determining if an Activity is disjoint.
if streams.ActivityStreamsActivityIsDisjointWith(object) {
  fmt.Println("I'm false.")
} else if streams.ActivityStreamsActivityIsDisjointWith(activity) {
  fmt.Println("I'm false -- a type is never disjoint with itself.")
} else if streams.ActivityStreamsActivityIsDisjointWith(repo) {
  fmt.Println("I'm also false!")
}
```

Wow! Navigating the RDF vocabulary at runtime is pretty neato. The last thing that you can do with the type hierarchy is to create your own duck-type interfaces if you want to group the types by a set of properties they have:

```go
// For application reasons, we care about types with "name" and "shares" properties
type nameShares interface {
  GetActivityStreamsName() vocab.ActivityStreamsNameProperty
  GetActivityStreamsShares() vocab.ActivityStreamsSharesProperty
  SetActivityStreamsName(vocab.ActivityStreamsNameProperty)
  SetActivityStreamsShares(vocab.ActivityStreamsSharesProperty)
}

if v, ok := aType.(nameShares); ok {
  _ = v.GetActivityStreamsName()
  _ = v.GetActivityStreamsShares()
}
```

Hopefully this allows you to harness the full power of the RDF vocabularies without having to write boilerplate code and sifting through the details of the vocabulary specifications. Go-Fed will compile and let you put properties where they are allowed to go, or at runtime let you determine and route code execution by the exact type or by navigating the RDF hierarchy. Finally, you can always duck-type to create your own interfaces as needed. It's up to you to use these tools in a sensical way.

## Working With Functional Properties {#Working-With-Functional-Properties}

Functional properties contain at most one value at a time. Setting a value on it will clear out any previous value on the property. These properties do not have iterators and have neither Append nor Prepend methods.

Creating these properties is straightforward:

```go
// Constructors are New<Vocabulary Name><Property Name>Property
first := streams.NewActivityStreamsFirstProperty()
if first.HasAny() {
  fmt.Println("I won't print because `first` has no values!")
}
```

You can always set an ActivityStreams property to be an IRI, but it doesn't always make sense to do so. For the ActivityStreams first property it can make sense for your application:

```go
iri, _ := url.Parse(https://go-fed.org/foobar/page/0)
first.SetIRI(iri)
if first.HasAny() {
  fmt.Println("I will print!")
}
if first.IsIRI() {
  fmt.Println("I will print, too!")
}
if first.IsActivityStreamsCollectionPage() {
  fmt.Println("I won't print because the value in `first` is not an ActivityStreams CollectionPage!")
}
```

Otherwise, you can embed a literal object as the value:

```go
first.SetActivityStreamsCollectionPage(streams.NewActivityStreamsCollectionPage())
if first.HasAny() {
  fmt.Println("I will still print!")
}
if first.IsIRI() {
  fmt.Println("I won't print! The IRI was overwritten")
}
if first.IsActivityStreamsCollectionPage() {
  fmt.Println("I will print now!")
}
```

And when processing the property, remember to check for `nil` results:

```go
if nil == first.GetActivityStreamsCollectionPage() {
  fmt.Println("I won't print, because the value is a CollectionPage!")
}
if nil == first.GetIRI() {
  fmt.Println("I will print, because the value is not an IRI!")
}
```

A functional property's value will either be a JSON-LD IRI, or an object, or some other primitive value like a boolean. When it is an object, that object must also have an IRI identifier. You may find that you just want to obtain IRIs without caring if it was an IRI literal or an object. To do so requires importing the `github.com/activity/pub` package:

```go
import (
  "github.com/activity/pub"
)

var iri *url.URL
var err error
iri, err = pub.ToId(first)
```

Properties are code generated from vocabulary definitions, so the methods available ensure that you are only able to process and manipulate properties using schema-compliant values. If you want to set a value on a property and go-fed doesn't have a method available, it is probably because it is not allowed by the vocabulary's definition. This will let you focus on the core of your application's needs without having to fight the ActivityStreams type system.

For example, the following will not compile:

```go
first.SetActivityStreamsActivity(streams.NewActivityStreamsActivity())
```

Because it would not follow the vocabulary definition, and hence there is no `SetActivityStreamsActivity` method for the `streams.ActivityStreamsFirstProperty` type.

## Working With Non-Functional Properties {#Working-With-Non-Functional-Properties}

Non-functional properties, contrary to the name, aren't broken! They are properties that can have any number of values. Plus, each value does not have to be the same type. This presents challenges in the Go world, where naive slices like to be a single type, even if that type is a generic `interface{}`. What a headache that presents! Fortunately, this library makes processing such data a breeze.

The non-functional properties build off of the Functional Properties concepts, so reading that section first is recommended.

A functional property can be identified by having methods for obtaining iterators like `Begin` or `At` or `Len`, or methods that allow appending or prepending values. Creating these properties is like their functional property bretheren:

```go
// Constructors are New<Vocabulary Name><Property Name>Property
items := streams.NewActivityStreamsItemsProperty()

if items.Empty() {
  fmt.Println("I will print because `items` has no values!")
}
```

Without knowing the existing length of the property, we can go ahead and append or prepend different values:

```go
// Append an IRI
iri, _ := url.Parse("https://go-fed.org/baz")
items.AppendIRI(iri)

// Prepend an ActivityStreams Note
items.PrependActivityStreamsNote(streams.NewActivityStreamsNote())

// Append some Type, without knowing its exact type. It will return
// an error if it is not allowed.
var someType vocab.Type = streams.NewForgeFedCommit()
if err := items.AppendType(someType); err != nil {
  // This will not print: ForgeFed Commit is OK to set on `items` property.
  fmt.Println(err)
}

// At this point, the `items` property holds:
//    [Note, "https://go-fed.org/baz", Commit]
```

{{% recommended %}}
Use the error returned from `AppendType` and `PrependType` methods.
{{% /recommended %}}

However, once the `Len` is known, you can manipulate the values of the property using indices:

```go
if items.Len() == 3 {
  fmt.Println("I will print!")
}

// The `items` property holds:
//    [Note, "https://go-fed.org/baz", Commit]
// Let's insert a ForgeFed Ticket like:
//    [Note, Ticket, "https://go-fed.org/baz", Commit]
items.InsertForgeFedTicket(1, streams.NewForgeFedTicket())

// Let's remove the Note and replace it with an Article:
//    [Article, Ticket, "https://go-fed.org/baz", Commit]
items.SetActivityStreamsArticle(0, streams.NewActivityStreamsArticle())

// Let's remove the IRI completely:
//    [Article, Ticket, Commit]
items.Remove(2)

// Let's swap the Ticket's and Commit's places:
//    [Article, Commit, Ticket]
items.Swap(1, 2)
```
{{%caution%}}
Iterators are invalidated when using `Insert`, `Set`, `Remove`, and `Swap` methods.
{{%/caution%}}

These methods are great when you need to manipulate the non-functional property in general and you do not really care what the values are. However, this tends to be an uncommon concern.

A far more common use-case is examining and manipulating the elements directly! For this, the library provies iterators. Iterators are powerful because an iterator behaves exactly like a functional property. So rather than writing code that has to handle functional properties, and then re-writing code to handle non-functional properties via `Insert`/`Set`/`Remove`/`Swap` methods, you can just process functional properties and then use iterators:

```go
// The `items` property holds:
//    [Article, Commit, Ticket]
// Let's examine the elements and process any Articles.

// First, a helper interface, which can be implemented by a
// functional property or an iterator!
type getsArticle interface {
  GetActivityStreamsArticle() vocab.ActivityStreamsArticle
}

// Let's create a function for processing:
markArticleToRead := func(ga getsArticle) {
  article := ga.GetActivityStreamsArticle()
  if article == nil {
    return
  }
  fmt.Println("Hey, check this out: ", article)
}

// Now we can process lot of different properties! For example,
// this will process the one Article in our `items` property.
for iter := items.Begin(); iter != items.End(); iter = iter.Next() {
  markArticleToRead(iter)
}

// But we could process a `subject` functional property too!
// For example, if there was a "Person authored Article"
// Relationship.
subject := streams.NewActivityStreamsSubjectProperty()
markArticleToRead(subject)
```

This gives you a lot of flexibility when processing non-functional properties across your application. It prepares you to write code that is scalable across the many varied vocabulary types.

Like a functional property, an iterator will either be a JSON-LD IRI, an object, or a value. If you're simply wanting an IRI, without caring whether the value is an IRI literal or an object, you can still use the `github.com/activity/pub` package:

```go
import (
  "github.com/activity/pub"
)

var iri *url.URL
var err error
for iter := items.Begin(); iter != items.End(); iter = iter.Next() {
	iri, err = pub.ToId(iter)
}
```

## Serialization {#Serialization}

The `streams` package provides three different `Resolver` types. The `JSONResolver` and `TypeResolver` each have a constructor that accepts one or more functions of the form:

```go
func(c context.Context, t vocab.<Some Type>) error
```

which you provide as a callback. This means to successfully turn a JSON payload that's expected to be an Article:

```go
jsonstr := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id":       "https://go-fed.org/foo",
  "name":     "Foo Bar",
  "inbox":    "https://go-fed.org/foo/inbox",
  "outbox":   "https://go-fed.org/foo/outbox",
  "type":     "Article",
  "url":      "https://go-fed.org/foo"
}`
var m map[string]interface{}
_ = json.Unmarshal([]byte(jsonstr), &m)

var article vocab.ActivityStreamsArticle
resolver := streams.NewJSONResolver(func(c context.Context, a vocab.ActivityStreamsArticle) error {
  // Example: store the article in the enclosing scope, for later.
  article = a
  // We could pass an error back up, if desired.
  return nil
})
ctx := context.Background()
err := resolver.Resolve(ctx, m)

if err == streams.ErrNoCallbackMatch {
  // The JSON payload is a type supported by Go-Fed but did
  // NOT match any of our provided functions.
} else if err == streams.ErrUnhandledType {
  // The JSON payload is NOT a type supported by Go-Fed.
} else {
  // The error came from one of the callback functions that
  // we provided.
}
```

When serializing a type, do not call the `Serialize` method on a type, but instead use the free function variant `streams.Serialize`:

```go
var jsonmap map[string]interface{}
jsonmap, _ = streams.Serialize(article)
b, _ := json.Marshal(jsonmap)
```

{{%caution%}}
Use `streams.Serialize`, not the `Serialize` method.
{{%/caution%}}

So far, we've looked at how to deserialize and serialize. However, the `streams.Serialize` function flexibly accepts any `vocab.Type` Go-Fed type, whereas the deserialization method we looked at only works when you know the specific type. How can you deserialize a JSON-LD payload into a `vocab.Type`?

```go
jsonstr := `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id":       "https://go-fed.org",
  "name":     "Go-Fed",
  "inbox":    "https://go-fed.org/inbox",
  "outbox":   "https://go-fed.org/outbox",
  "type":     "Organization",
  "url":      "https://go-fed.org"
}`
var m map[string]interface{}
err := json.Unmarshal([]byte(jsonstr), &m)

var t vocab.Type
t, err = streams.ToType(ctx, m)
```

This lets you be as specific or general as necessary when processing JSON-LD data. However, now that you have a `vocab.Type`, you may find that after a certain point it'll be necessary to resolve it further to its specific type. The `TypeResolver` is then the tool of choice, which accepts a `vocab.Type` instead of a `map[string]interface{}`: 

```go
typeResolver := streams.NewTypeResolver(func(c context.Context, o vocab.ActivityStreamsOrganization) error {
  // ...
  return nil
})
ctx := context.Background()
// Pass in a vocab.Type instead of map[string]interface{}.
err := resolver.Resolve(ctx, t)
```

With `JSONResolver`, `TypeResolver`, and `streams.Serialize`, you are prepared for serializing and deserializing into the vocabulary types required for your application.

## Supported Vocabularies {#Supported-Vocabularies}

The following vocabularies are supported to some degree:

- All [Core and Extended ActivityStreams](https://www.w3.org/TR/activitystreams-vocabulary/) types.
- All [ForgeFed](https://go-fed.org/ref/activity/streams) types.
- About half of the [Mastodon/Toot](https://github.com/go-fed/activity/issues/122) types and properties.
- A few select items from the [Security](https://w3c-ccg.github.io/security-vocab/) proposal.

## IRIs and Deserialization {#IRIs-And-Deserialization}

As mentioned in numerous sections above, often the values for properties will be JSON-LD IRIs, which are addresses where ActivityStreams data can be fetched. The `streams` library does not do this fetching for you, so you will need to provide the method of resolving an IRI into data.

Note that the current Fediverse commonly uses IRIs that use the HTTPS protocol and often use HTTP Signatures to indicate on which user's behalf the HTTP request is being made. These are conventions, but any IRI may be used to build the method by which the linked data is fetched so it may not always be specified as fetchable over HTTP, and require a different application protocol instead. The `streams` package does not add any constraints to this by not providing any solutions either.

There are two complementary libraries that can be used to address these concerns. The [Go-Fed ActivityPub Library](https://go-fed.org/ref/activity/pub) (`pub`) provides a `HttpSigTransport` to dereference IRIs using HTTPS and HTTP Signatures. Alternatively, the [Go-Fed HTTP Signatures Library](https://go-fed.org/ref/httpsig) (`httpsig`) provides primitives that let you use HTTP Signatures in the HTTP Client of your choice.