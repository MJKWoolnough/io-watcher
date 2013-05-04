# package watcher
	import "/home/michael/Programming/Go/src/github.com/MJKWoolnough/io-watcher"


# CONSTANTS

	const (
	    WATCH_CREATE uint8 = 1 << iota
	    WATCH_RENAME
	    WATCH_MODIFY
	    WATCH_DELETE
	)


# FUNCTIONS

	func StopWatch(pathname string)
StopWatch will stop a specified path for being watched for changes.

	func Watch(pathname string, w Watcher) error
Watch will register a Watcher to a path name to be executed upon a change.


# TYPES

	type Watcher interface {
	    Update(string, uint8)
	}

	type WatcherFunc func(string, uint8)
WatcherFunc allows any function to act as a Watcher.

	func (w WatcherFunc) Update(pathname string, mask uint8)


