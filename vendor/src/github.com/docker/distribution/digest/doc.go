// Package digest provides a generalized type to opaquely represent message
// digests and their operations within the registry. The Digest type is
// designed to serve as a flexible identifier in a content-addressable system.
// More importantly, it provides tools and wrappers to work with tarsums and
// hash.Hash-based digests with little effort.
//
// Basics
//
// The format of a digest is simply a string with two parts, dubbed the
// "algorithm" and the "digest", separated by a colon:
//
// 	<algorithm>:<digest>
//
// An example of a sha256 digest representation follows:
//
// 	sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc
//
// In this case, the string "sha256" is the algorithm and the hex bytes are
// the "digest". A tarsum example will be more illustrative of the use case
// involved in the registry:
//
// 	tarsum+sha256:e58fcf7418d4390dec8e8fb69d88c06ec07039d651fedd3aa72af9972e7d046b
//
// For this, we consider the algorithm to be "tarsum+sha256". Prudent
// applications will favor the ParseDigest function to verify the format over
// using simple type casts. However, a normal string can be cast as a digest
// with a simple type conversion:
//
// 	Digest("tarsum+sha256:e58fcf7418d4390dec8e8fb69d88c06ec07039d651fedd3aa72af9972e7d046b")
//
// Because the Digest type is simply a string, once a valid Digest is
// obtained, comparisons are cheap, quick and simple to express with the
// standard equality operator.
//
// Verification
//
// The main benefit of using the Digest type is simple verification against a
// given digest. The Verifier interface, modeled after the stdlib hash.Hash
// interface, provides a common write sink for digest verification. After
// writing is complete, calling the Verifier.Verified method will indicate
// whether or not the stream of bytes matches the target digest.
//
// Missing Features
//
// In addition to the above, we intend to add the following features to this
// package:
//
// 1. A Digester type that supports write sink digest calculation.
//
// 2. Suspend and resume of ongoing digest calculations to support efficient digest verification in the registry.
//
package digest
