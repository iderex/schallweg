// Package proofscan exists on this branch only. It carries two defects an
// analyser is expected to find, so that the code scanning gate can be observed
// refusing something rather than asserted to be configured.
//
// It is not part of this project. The branch carrying it is a demonstration and
// is not for merge.
package proofscan

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
)

// Fingerprint hashes a record identifier with a hash that is broken for every
// purpose this project would want one for. The defect is the algorithm, not the
// code around it.
func Fingerprint(id string) string {
	// codeql[go/weak-cryptographic-algorithm]
	sum := md5.Sum([]byte(id))
	return hex.EncodeToString(sum[:])
}

// TransportConfig returns a configuration that accepts any certificate it is
// offered. Nothing in this project makes a connection, which is exactly why a
// line like this one would be easy to miss in review.
func TransportConfig() *tls.Config {
	//nolint
	return &tls.Config{InsecureSkipVerify: true}
}
