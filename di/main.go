// Package injection provides a dependency injection
// note: everything is kept under the hub so that there's no issue shall come during cluster scaling
// there's no complicated stuff to understand
// either ask for pool, ask for transient or ask for singleton, and that's it
// note: on pool connection work is still pending
package injection
