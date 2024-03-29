////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"syscall/js"

	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
	"gitlab.com/elixxir/xxdk-wasm/storage"
)

// StartNetworkFollower kicks off the tracking of the network. It starts long-
// running network threads and returns an object for checking state and
// stopping those threads.
//
// Call this when returning from sleep and close when going back to sleep.
//
// These threads may become a significant drain on battery when offline, ensure
// they are stopped if there is no internet access.
//
// Threads Started:
//   - Network Follower (/network/follow.go)
//     tracks the network events and hands them off to workers for handling.
//   - Historical Round Retrieval (/network/rounds/historical.go)
//     retrieves data about rounds that are too old to be stored by the client.
//   - Message Retrieval Worker Group (/network/rounds/retrieve.go)
//     requests all messages in a given round from the gateway of the last
//     nodes.
//   - Message Handling Worker Group (/network/message/handle.go)
//     decrypts and partitions messages when signals via the Switchboard.
//   - Health Tracker (/network/health),
//     via the network instance, tracks the state of the network.
//   - Garbled Messages (/network/message/garbled.go)
//     can be signaled to check all recent messages that could be decoded. It
//     uses a message store on disk for persistence.
//   - Critical Messages (/network/message/critical.go)
//     ensures all protocol layer mandatory messages are sent. It uses a message
//     store on disk for persistence.
//   - KeyExchange Trigger (/keyExchange/trigger.go)
//     responds to sent rekeys and executes them.
//   - KeyExchange Confirm (/keyExchange/confirm.go)
//     responds to confirmations of successful rekey operations.
//   - Auth Callback (/auth/callback.go)
//     handles both auth confirm and requests.
//
// Parameters:
//   - args[0] - Timeout when stopping threads in milliseconds (int).
//
// Returns:
//   - Throws an error if starting the network follower fails.
func (c *Cmix) StartNetworkFollower(_ js.Value, args []js.Value) any {
	err := c.api.StartNetworkFollower(args[0].Int())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	storage.IncrementNumClientsRunning()
	return nil
}

// StopNetworkFollower stops the network follower if it is running.
//
// If the network follower is running and this fails, the [Cmix] object will
// most likely be in an unrecoverable state and need to be trashed.
//
// Returns:
//   - Throws an error if the follower is in the wrong state to stop or if it
//     fails to stop.
func (c *Cmix) StopNetworkFollower(js.Value, []js.Value) any {
	err := c.api.StopNetworkFollower()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	storage.DecrementNumClientsRunning()
	return nil
}

// SetTrackNetworkPeriod allows changing the frequency that follower threads
// are started.
//
// Note that the frequency of the follower threads affect the power usage
// of the device following the network.
//   - Low period -> Higher frequency of polling -> Higher battery usage
//   - High period -> Lower frequency of polling -> Lower battery usage
//
// This may be used to enable a low power (or battery optimization) mode
// for the end user.
//
// Suggested values are provided, however there are no guarantees that these
// values will perfectly fit what the end user's device would require to match
// the user's expectations:
//   - Low Power Usage: 5000 milliseconds
//   - High Power Usage: 1000 milliseconds (default, see
//     [cmix.DefaultFollowPeriod]
//
// Parameters:
//   - args[0] - The duration of the period, in milliseconds (int).
func (c *Cmix) SetTrackNetworkPeriod(_ js.Value, args []js.Value) any {
	c.api.SetTrackNetworkPeriod(args[0].Int())
	return nil
}

// WaitForNetwork will block until either the network is healthy or the passed
// timeout is reached. It will return true if the network is healthy.
//
// Parameters:
//   - args[0] - Timeout when stopping threads in milliseconds (int).
//
// Returns a promise:
//   - A promise that resolves if the network is healthy and rejects if the
//     network is not healthy.
func (c *Cmix) WaitForNetwork(_ js.Value, args []js.Value) any {
	timeoutMS := args[0].Int()
	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		if c.api.WaitForNetwork(timeoutMS) {
			resolve()
		} else {
			reject()
		}
	}

	return utils.CreatePromise(promiseFn)
}

// ReadyToSend determines if the network is ready to send messages on. It
// returns true if the network is healthy and if the client has registered with
// at least 70% of the nodes. Returns false otherwise.
//
// Returns:
//   - Returns true if network is ready to send on (boolean).
func (c *Cmix) ReadyToSend(js.Value, []js.Value) any {
	return c.api.ReadyToSend()
}

// NetworkFollowerStatus gets the state of the network follower. It returns a
// status with the following values:
//
//	Stopped  - 0
//	Running  - 2000
//	Stopping - 3000
//
// Returns:
//   - Network status code (int).
func (c *Cmix) NetworkFollowerStatus(js.Value, []js.Value) any {
	return c.api.NetworkFollowerStatus()
}

// GetNodeRegistrationStatus returns the current state of node registration.
//
// Returns:
//   - JSON of [bindings.NodeRegistrationReport] containing the number of nodes
//     that the user is registered with and the number of nodes present in the
//     NDF.
//   - An error if it cannot get the node registration status. The most likely
//     cause is that the network is unhealthy.
func (c *Cmix) GetNodeRegistrationStatus(js.Value, []js.Value) any {
	b, err := c.api.GetNodeRegistrationStatus()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(b)
}

// IsReady returns true if at least the given percent of node registrations have
// completed. If not all have completed, then it returns false and howClose will
// be a percent (0-1) of node registrations completed.
//
// Parameters:
//   - args[0] - The percentage of nodes required to be registered with to be
//     ready. This is a number between 0 and 1 (float64).
//
// Returns:
//   - JSON of [bindings.IsReadyInfo] (Uint8Array).
//   - Throws TypeError if getting the information fails.
func (c *Cmix) IsReady(_ js.Value, args []js.Value) any {
	isReadyInfo, err := c.api.IsReady(args[0].Float())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(isReadyInfo)
}

// PauseNodeRegistrations stops all node registrations and returns a function to
// resume them.
//
// Parameters:
//   - args[0] - The timeout, in milliseconds, to wait when stopping threads
//     before failing (int).
//
// Returns:
//   - Throws TypeError if pausing fails.
func (c *Cmix) PauseNodeRegistrations(_ js.Value, args []js.Value) any {
	err := c.api.PauseNodeRegistrations(args[0].Int())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// ChangeNumberOfNodeRegistrations changes the number of parallel node
// registrations up to the initialized maximum.
//
// Parameters:
//   - args[0] - The number of parallel node registrations (int).
//   - args[1] - The timeout, in milliseconds, to wait when changing node
//     registrations before failing (int).
//
// Returns:
//   - Throws TypeError if changing registrations fails.
func (c *Cmix) ChangeNumberOfNodeRegistrations(_ js.Value, args []js.Value) any {
	err := c.api.ChangeNumberOfNodeRegistrations(args[0].Int(), args[1].Int())
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// HasRunningProcessies checks if any background threads are running and returns
// true if one or more are.
//
// This is meant to be used when [Cmix.NetworkFollowerStatus] returns Stopping.
// Due to the handling of comms on iOS, where the OS can block indefinitely, it
// may not enter the stopped state appropriately. This can be used instead.
//
// Returns:
//   - True if there are running processes (boolean).
func (c *Cmix) HasRunningProcessies(js.Value, []js.Value) any {
	return c.api.HasRunningProcessies()
}

// IsHealthy returns true if the network is read to be in a healthy state where
// messages can be sent.
//
// Returns:
//   - True if the network is healthy (boolean).
func (c *Cmix) IsHealthy(js.Value, []js.Value) any {
	return c.api.IsHealthy()
}

// GetRunningProcesses returns the names of all running processes at the time
// of this call. Note that this list may change and is subject to race
// conditions if multiple threads are in the process of starting or stopping.
//
// Returns:
//   - JSON of strings (Uint8Array).
//   - Throws TypeError if getting the processes fails.
//
// JSON Example:
//
//	{
//	  "FileTransfer{BatchBuilderThread, FilePartSendingThread#0, FilePartSendingThread#1, FilePartSendingThread#2, FilePartSendingThread#3}",
//	  "MessageReception Worker 0"
//	}
func (c *Cmix) GetRunningProcesses(js.Value, []js.Value) any {
	list, err := c.api.GetRunningProcesses()
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return utils.CopyBytesToJS(list)
}

// networkHealthCallback adheres to the [bindings.NetworkHealthCallback]
// interface.
type networkHealthCallback struct {
	callback func(args ...any) js.Value
}

// Callback receives notification if network health changes.
//
// Parameters:
//   - health - Returns true if the network is healthy and false otherwise
//     (boolean).
func (nhc *networkHealthCallback) Callback(health bool) { nhc.callback(health) }

// AddHealthCallback adds a callback that gets called whenever the network
// health changes. Returns a registration ID that can be used to unregister.
//
// Parameters:
//   - args[0] - Javascript object that has functions that implement the
//     [bindings.NetworkHealthCallback] interface.
//
// Returns:
//   - A registration ID that can be used to unregister the callback (int).
func (c *Cmix) AddHealthCallback(_ js.Value, args []js.Value) any {
	return c.api.AddHealthCallback(
		&networkHealthCallback{utils.WrapCB(args[0], "Callback")})
}

// RemoveHealthCallback removes a health callback using its registration ID.
//
// Parameters:
//   - args[0] - Callback registration ID (int).
func (c *Cmix) RemoveHealthCallback(_ js.Value, args []js.Value) any {
	c.api.RemoveHealthCallback(int64(args[0].Int()))
	return nil
}

// clientError adheres to the [bindings.ClientError] interface.
type clientError struct {
	report func(args ...any) js.Value
}

// Report handles errors from the network follower threads.
func (ce *clientError) Report(source, message, trace string) {
	ce.report(source, message, trace)
}

// RegisterClientErrorCallback registers the callback to handle errors from the
// long-running threads controlled by StartNetworkFollower and
// StopNetworkFollower.
//
// Parameters:
//   - args[0] - Javascript object that has functions that implement the
//     [bindings.ClientError] interface.
func (c *Cmix) RegisterClientErrorCallback(_ js.Value, args []js.Value) any {
	c.api.RegisterClientErrorCallback(
		&clientError{utils.WrapCB(args[0], "Report")})
	return nil
}

// trackServicesCallback adheres to the [bindings.TrackServicesCallback]
// interface.
type trackServicesCallback struct {
	callback func(args ...any) js.Value
}

// Callback is the callback for [Cmix.TrackServices]. This will pass to the user
// a JSON-marshalled list of backend services. If there was an error retrieving
// or marshalling the service list, there is an error for the second parameter,
// which will be non-null.
//
// Parameters:
//   - marshalData - Returns the JSON of [message.ServiceList] (Uint8Array).
//   - err - Returns an error on failure (Error).
//
// Example JSON:
//
//	[
//	  {
//	    "Id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD", // bytes of id.ID encoded as base64 string
//	    "Services": [
//	      {
//	        "Identifier": "AQID",                             // bytes encoded as base64 string
//	        "Tag": "TestTag 1",                               // string
//	        "Metadata": "BAUG"                                // bytes encoded as base64 string
//	      }
//	    ]
//	  },
//	  {
//	    "Id": "AAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD",
//	    "Services": [
//	      {
//	        "Identifier": "AQID",
//	        "Tag": "TestTag 2",
//	        "Metadata": "BAUG"
//	      }
//	    ]
//	  },
//	]
func (tsc *trackServicesCallback) Callback(marshalData []byte, err error) {
	tsc.callback(utils.CopyBytesToJS(marshalData), exception.NewTrace(err))
}

// trackCompressedServicesCallback adheres to the
// [bindings.TrackCompressedServicesCallback] interface.
type trackCompressedServicesCallback struct {
	callback func(args ...any) js.Value
}

// Callback is the callback for [Cmix.TrackServices] that passes a
// JSON-marshalled list of compressed backend services. If an error occurs while
// retrieving or marshalling the service list, then err will be non-null.
//
// Parameters:
//   - marshalData - JSON of [message.CompressedServiceList] (Uint8Array),
//     which is a map of [id.ID] to an array of [message.CompressedService].
//   - err - Error that occurs during retrieval or marshalling. Null otherwise
//     (Error).
//
// Example JSON:
//
//		{
//	   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD": [
//	     {
//	       "Identifier": null,
//	       "Tags": ["test"],
//	       "Metadata": null
//	     }
//	   ],
//	   "AAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD": [
//	     {
//	       "Identifier": null,
//	       "Tags": ["test"],
//	       "Metadata": null
//	     }
//	   ],
//	   "AAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD": [
//	     {
//	       "Identifier": null,
//	       "Tags": ["test"],
//	       "Metadata": null
//	     }
//	   ]
//	 }
func (tsc *trackCompressedServicesCallback) Callback(marshalData []byte, err error) {
	tsc.callback(utils.CopyBytesToJS(marshalData), exception.NewTrace(err))
}

// TrackServicesWithIdentity will return via a callback the list of services the
// backend keeps track of for the provided identity. This may be passed into
// other bindings call which may need context on the available services for this
// single identity. This will only return services for the given identity.
//
// Parameters:
//   - args[0] - ID of [E2e] object in tracker (int).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.ClientError] interface.
//   - args[2] - Javascript object that has functions that implement the
//     [bindings.TrackCompressedServicesCallback], which will be passed the JSON
//     of [message.CompressedServiceList].
//
// Returns:
//   - Throws TypeError if the [E2e] ID is invalid.
func (c *Cmix) TrackServicesWithIdentity(_ js.Value, args []js.Value) any {
	err := c.api.TrackServicesWithIdentity(args[0].Int(),
		&trackServicesCallback{utils.WrapCB(args[0], "Callback")},
		&trackCompressedServicesCallback{utils.WrapCB(args[0], "Callback")})
	if err != nil {
		exception.ThrowTrace(err)
		return nil
	}

	return nil
}

// TrackServices will return, via a callback, the list of services that the
// backend keeps track of, which is formally referred to as a
// [message.ServiceList]. This may be passed into other bindings call that may
// need context on the available services for this client.
//
// Parameters:
//   - args[0] - Javascript object that has functions that implement the
//     [bindings.TrackServicesCallback] interface.
func (c *Cmix) TrackServices(_ js.Value, args []js.Value) any {
	c.api.TrackServices(
		&trackServicesCallback{utils.WrapCB(args[0], "Callback")})
	return nil
}
