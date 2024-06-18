# Conventions

This document aims to cover the conventions and guidance to consider when 
making changes the Go SDK.

## Methods

- All methods should take a maximum of 3 parameter. See examples in [experimental](./experimental.md)
- The first parameter is always `context.Context`.
- The second is a `*ResourceContainer`.
- The final is a struct of available parameters for the method. 
  - The parameter naming convention should be `<MethodName>Params`. Example: 
    method name of `GetDNSRecords` has a struct parameter name of 
    `GetDNSRecordsParams`.
  - Do not share parameter structs between methods. Each should have a dedicated
    one.
  - Even if you don't intend to have parameter configurations, you should add
    the third parameter to your method signature for future flexibility.
  
## Types

### Booleans

- Should always be represented as pointers in structs with an `omitempty` 
  marshaling tag (most commonly as JSON). This ensures you can determine unset, 
  false and truthy values.

### `time.Time`

- Should always be represented as pointers in structs.

### Ports (0-65535)

- Should use `uint16` unless you have a reason to restrict the port range in 
  which case, you should also provide a validator on the type.

## Marshaling/unmarshaling

- Avoid custom marshal/unmarshal handlers unless absolutely necessary. They can
  be difficult to debug in a larger codebase.
