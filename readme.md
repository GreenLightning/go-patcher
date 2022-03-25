[![Go Reference](https://pkg.go.dev/badge/github.com/GreenLightning/go-patcher.svg)](https://pkg.go.dev/github.com/GreenLightning/go-patcher)

This package provides a system to record edits (insertions, deletions,
replacements) and apply them to a block of data while checking for conflicts.

For example:

```go
package main

import (
    "fmt"
    "strings"

    "github.com/GreenLightning/go-patcher"
)

func main() {
    input := "The brown fox jumps twice over the lazy horse"

    var p patcher.Patcher
    p.InsertString(3, " quick")
    p.Delete(strings.Index(input, "twice "), len("twice "))
    p.RewriteString(40, 5, "dog")

    output, err := p.PatchString(input)
    if err != nil {
        panic(err)
    }

    // The quick brown fox jumps over the lazy dog
    fmt.Println(output)
}
```

This package is useful when you do not want to copy the full input on every edit
(e.g. using `strings.Replace`) or when you want to perform multiple edits
referring to indices in the original input without them affecting each other
(e.g. if you have a list of indices and you delete some characters at one index,
the subsequent indices have to be modified to account for the deleted
characters).

Using this package, all offsets are relative to the input given to `PatchString`
/ `PatchBytes`. E.g. look at the line calling `Delete` in the example code
above, where the offset is the index of 'twice' in the input, oblivious of the
insertion on the previous line.


Multiple edits modifying the same region of the input will result in a conflict
error:

```go
    input := "abcde"
    var p patcher.Patcher
    p.Delete(1, 2) // deletes bc
    p.Delete(3, 1) // deletes d

    // ok, returns ae
    p.PatchString(input)

    // ...

    input := "abcde"
    var p patcher.Patcher
    p.Delete(1, 3) // deletes bcd
    p.Delete(2, 1) // wants to delete c again

    // error
    p.PatchString(input)
```

Multiple inserts to the same position are inserted one after the other in the
order of the calls to the insert functions:

```go
    input := "ae"
    var p Patcher
    p.InsertString(1, "b")
    p.InsertString(1, "c")
    p.InsertString(1, "d")

    // ok, returns abcde
    p.PatchString(input)
```

For convenience there are multiple versions of each function taking either a
string or a byte slice as argument (e.g. `InsertString` / `InsertBytes`,
`PatchString` / `PatchBytes`). The `String` and `Bytes` variants can be mixed as
needed.
