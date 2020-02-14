# goodies-go
Reuseable code pieces for golang.

## Good Advises

### Errors

1. Use %s to print err.
2. Do not use 'failed', 'unsuccessful', etc, the error itself says it is a error already.
3. Wrap errors as each layer the error passed on.
4. Each layer's wrapping states only what it is doing.
