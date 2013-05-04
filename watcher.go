// Copyright (c) 2013 - Michael Woolnough <michael.woolnough@gmail.com>
// 
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met: 
// 
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer. 
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution. 
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package watcher will watch files or directories for any changes and execute the given functions.

package watcher

import (
	"github.com/howeyc/fsnotify"
	"os"
)

const (
	WATCH_CREATE uint8 = 1 << iota
	WATCH_RENAME
	WATCH_MODIFY
	WATCH_DELETE
)

type Watcher interface {
	Update(string, uint8)
}

// WatcherFunc allows any function to act as a Watcher.
type WatcherFunc func (string, uint8)

func (w WatcherFunc) Update(pathname string, mask uint8) {
	w(pathname, mask)
}

type pair struct {
	pathname string
	watcher  Watcher
}

var (
	watcher     *fsnotify.Watcher
	files       map[string][]Watcher
	remove      chan string
	add         chan pair
)

// Watch will register a Watcher to a path name to be executed upon a change.
func Watch(pathname string, w Watcher) error {
	if watcher == nil {
		var err error
		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		go watch()
	}
	add <- pair { pathname, w }
	return nil
}

// StopWatch will stop a specified path for being watched for changes.
func StopWatch(pathname string) {
	remove <- pathname
}

func watch() {
	for {
		select {
			case ev := <-watcher.Event:
				if l, ok := files[ev.Name]; ok {
					mask := uint8(0)
					if ev.IsCreate() {
						mask |= WATCH_CREATE
					}
					if ev.IsRename() {
						mask |= WATCH_RENAME
					}
					if ev.IsModify() {
						mask |= WATCH_MODIFY
					}
					if ev.IsDelete() {
						mask |= WATCH_DELETE
						if _, err := os.Stat(ev.Name); err == nil {
							watcher.RemoveWatch(ev.Name)
							watcher.Watch(ev.Name)
						} else {
							remove <- ev.Name
						}
					}
					for _, w := range l {
						go w.Update(ev.Name, mask)
					}
				}
// 			case err := <-watcher.Error:
// 				fmt.Fprintf(os.Stderr, "Watcher error: %q\n", err)
			case <-watcher.Error:
			case pathname := <-remove:
				watcher.RemoveWatch(pathname)
				delete(files, pathname)
				if len(files) == 0 {
					watcher.Close()
					watcher = nil
					return
				}
			case p := <-add:
				if f, ok := files[p.pathname]; ok {
					f = append(f, p.watcher)
				} else {
					watcher.Watch(p.pathname)
					files[p.pathname] = []Watcher { p.watcher }
				}
		}
	}
}

func init() {
	remove = make(chan string)
	add    = make(chan pair)
	files  = make(map[string][]Watcher)
}
