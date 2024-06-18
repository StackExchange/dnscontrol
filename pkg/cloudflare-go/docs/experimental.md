# README (experimental)

An experimental and incremental update of the library focusing on a more modern
and consistent experience.

## Improvements

### Automatically paginate `List` operations by default

`List()` methods will automatically paginate all resources **unless** 
`PerPage` or `Page` is supplied as a part of the `$entityListParams`.

This allows us the best of both worlds where if you need to explicitly 
override the inbuilt pagination, you have the ability to.

## Nested methods and services

Not all methods are defined at the top level. Instead, they are nested under
service objects.

```golang
// old
client.ListZones(...)
client.ZoneLevelAccessServiceTokens(...)

// new
client.Zones.List()
client.Access.ServiceTokens(...)
```

This avoids polluting the global namespace and having more specific methods
for services.

### Consistent CRUD method signatures

Majority of methods on an entity will follow a standard method signature.

| Signature                                                                 | Purpose                                            | Return value                                                                                                                   |
| ------------------------------------------------------------------------- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `Get(ctx, *ResourceContainer, $entityID) ($entity, error)`                | Fetches a single entity by a `$entityID`.          | Returns the entity and `error` based on the listing parameters.                                                                |
| `List(ctx, *ResourceContainer, params) ([]$entity, ResultInfo, error)` | Fetches all entities.                              | Returns the list of matching entities, the result information (pagination, fail/success and additional metadata), and `error`. |
| `New(ctx, *ResourceContainer, params) ($entity, error)`                | Creates a new entity with the provided parameters. | Returns the newly created entity and `error`.                                                                                  |
| `Update(ctx, *ResourceContainer, params) ($entity, error)`             | Updates an existing entity.                        | Returns the updated entity and `error`.                                                                                        |
| `Delete(ctx, *ResourceContainer, $entityID) (error)`                      | Deletes a single entity by a `$entityID`.          | Returns `error`.                                                                                                               |

- `*ResourceContainer` determines the "level" of the resource and where it will
  operate at. Operated using `UserIdentifier`, `ZoneIdentifier`, and
  `AccountIdentifier` respectively.
- `$entityID` is the resource identifier.
- `params` is a complex structure that allows filtering/finding resources
  matching the struct fields. By providing a structure as the third argument
  in all the methods that require it, we can add/remove fields without the 
  need for a breaking change and instead can issue deprecation notices when
  specific fields are used.
- `$entity` the resource being operated on.

Exceptions to this convention will be:

- Methods outside of CRUD operations
- Top level level concepts such as `Accounts` and `Zones`

#### Examples

`DNSRecord` is used below for the examples however, all entites will implement the
same methods and interfaces.

```go
params := cloudflare.ClientParams{
  Key: "3bc3be114fb6323adc5b0ad7422d193a",
  Email: "someone@example.com",
  HTTPClient: myCustomHTTPClient,
  // ...
}
c, err := cloudflare.NewExperimental(params)
```

**Create a new DNS record**

```go
dParams := &cloudflare.DNSRecordParams{
  Name: "@",
  Content: "foo.example.com",
  TTL: 300,
}
r, _ := c.DNSRecord.New(context.TODO(), cloudflare.ZoneIdentifier("b026324c6904b2a9cb4b88d6d61c81d1"), dParams)
```

**Fetching a known DNS record by ID**

```go
r, _ := c.DNSRecord.Get(context.TODO(), cloudflare.ZoneIdentifier("b026324c6904b2a9cb4b88d6d61c81d1"), "3e7705498e8be60520841409ebc69bc1")
```

**Listing all records matching a single account ID (filter option)**

```go
dParams := &cloudflare.DNSRecordListParams{
  AccountID: "d8e8fca2dc0f896fd7cb4cb0031ba249"
}
r, _, _ := c.DNSRecord.List(context.TODO(), dParams)
```

**Update an existing DNS record**

```go
dParams := &cloudflare.DNSRecordParams{
  ID: "b5163cf270a3fbac34827c4a2713eef4",
  Name: "@",
  Content: "bar.example.com",
  TTL: 300,
}
r, _ := c.DNSRecord.Update(context.TODO(), cloudflare.ZoneIdentifier("b026324c6904b2a9cb4b88d6d61c81d1"), dParams)
```

**Delete a DNS Record**

```go
r, _ := c.DNSRecord.Delete(context.TODO(), cloudflare.ZoneIdentifier("b026324c6904b2a9cb4b88d6d61c81d1"), "b5163cf270a3fbac34827c4a2713eef4")
```
