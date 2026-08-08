// Package store is the data access layer: reading, validating and serialising
// component records, project files and reference datasets.
//
// It is the only place in this module that touches the file system, and it is
// therefore the place every untrusted byte enters. Two obligations follow.
// Nothing reaches the kernel that has not been validated. And a byte sequence
// this package cannot make sense of becomes a refusal naming the file, the
// position and what was expected, never a zero value that travels onwards.
//
// It imports the kernel and acoustic packages. Nothing imports it except cmd.
//
// The formats and the versioning rule are docs/decisions/data-format.md. What a
// record's provenance has to carry is docs/decisions/certificate-extraction.md.
// A reference dataset that exists only in the standard is supplied by the
// operator rather than shipped, which is docs/decisions/standard-text.md, and
// the refusal when one is absent is written here rather than in the kernel.
//
// What is here today is the spectrum exchange format, specified in
// docs/formats/spectrum.md and read and written here. Issue #73 adds the record
// schema and issue #91 the project file.
package store
