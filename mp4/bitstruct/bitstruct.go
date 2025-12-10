package bitstruct

/*
Package bitstruct provides utilities for bit-level serialization of data structures.

The functions, methods, structs, and interfaces in this package originate from the
go-mp4 project. All context-related logic has been removed to focus solely on
bit-level serialization.

This package makes it easier to perform free-form byte and bit serialization using
the tooling from go-mp4, while retaining the MP4 serialization rules (big-endian
encoding). Unlike the original go-mp4 implementation—which requires explicit box
type registration and enforces strict structural constraints—bitstruct extracts and
simplifies only the serialization-related components, allowing more flexible and
direct serialization workflows.

The field tag name has been changed to "bit", and its syntax remains consistent with
the original go-mp4 project. However, features related to "opt" and "ver/nver"
annotations are no longer supported.
*/
