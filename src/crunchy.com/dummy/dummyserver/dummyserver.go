/*
 Copyright 2015 Crunchy Data Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"crunchy.com/dummy"
	"crunchy.com/logutil"
	"net"
	"net/http"
	"net/rpc"
)

func main() {

	logutil.Log("dummy starting\n")
	command := new(dummy.Command)
	rpc.Register(command)
	logutil.Log("Command registered\n")
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":13013")
	logutil.Log("listening\n")
	if e != nil {
		logutil.Log(e.Error())
		panic("could not listen on rpc socker")
	}
	logutil.Log("about to serve\n")
	http.Serve(l, nil)
	logutil.Log("after serve\n")
}
