/*
Package tree is a minimal interval tree implementation.

All interval types implementing the tree.Interface can use this library for fast lookups
and a stringified tree representation.

Application example:
The author uses it mainly for fast O(log n) lookups in IP ranges
where patricia-tries with O(1) are not feasible.
The tree representation allows a clear overview of the IP address block nestings.
*/
package tree
