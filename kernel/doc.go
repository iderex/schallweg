// Package kernel is the calculation: constructions, elements, junctions,
// transmission situations, path enumeration and the arithmetic along a path.
//
// It imports the acoustic package and nothing else from this module, and like
// that package it performs no input or output and prints nothing. A thing that
// went wrong is a returned error; a thing worth saying is a field in the result.
//
// It is a public interface rather than an implementation detail of the command
// line, because a kernel that only one program can use is the thing this project
// exists to answer. Until the first release tagged with a major version of 1,
// every exported name here may change without notice. The contracts that are
// already stable are the result structure and the record schema, which are
// versioned by docs/decisions/result-contents.md and
// docs/decisions/data-format.md.
//
// What this package holds is fixed by docs/decisions/model-shape.md, which keeps
// one situation model and one path enumeration across both evaluation
// strategies, by docs/decisions/element-model.md, and by
// docs/decisions/junctions.md.
//
// The package is empty today. The milestones that fill it start at issue #48.
package kernel
