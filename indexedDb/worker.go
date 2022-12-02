////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package indexedDb

import (
	"encoding/json"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"sync"
	"syscall/js"
	"time"
)

const (
	// InitialID is the ID for the first item in the handler list.
	InitialID = 0
)

// Worker response timeouts.
const (
	workerInitialConnectionTimeout   = 2 * time.Second
	databaseInitialConnectionTimeout = 1 * time.Second
)

// List of tags that can be used when sending a message.
const (
	NewWASMEventModelTag   Tag = "NewWASMEventModel"
	JoinChannelTag         Tag = "JoinChannel"
	LeaveChannelTag        Tag = "LeaveChannel"
	ReceiveMessageTag      Tag = "ReceiveMessage"
	ReceiveReplyTag        Tag = "ReceiveReply"
	ReceiveReactionTag     Tag = "ReceiveReaction"
	UpdateFromMessageIDTag Tag = "UpdateFromMessageID"
	UpdateFromUUIDTag      Tag = "UpdateFromUUID"
	GetMessageTag          Tag = "GetMessage"
	ReadyTag               Tag = "Ready"
	WorkerResponseTag      Tag = "WorkerResponse"
)

// deleteAfterReceiving is a list of Tags that will have their handler deleted
// after a message is received. This is mainly used for responses where the
// handler will only handle it once and never again.
var deleteAfterReceiving = map[Tag]struct{}{
	ReadyTag:          {},
	WorkerResponseTag: {},
}

// handlerFn is the function that handles incoming data from the worker.
type handlerFn func(data []byte)

// workerHandler manages the handling of messages received from the worker.
type workerHandler struct {
	// worker represents is the Worker Javascript object
	worker js.Value

	// handlers are a list of handlers that handle a specific message received
	// from the worker. Each handler is keyed on a tag specifying how the
	// received message should be handled. If the message is a reply to a
	// message sent to the worker, then the handler is also keyed on a unique
	// ID.
	handlers map[Tag]map[uint64]handlerFn

	// idCount tracks the newest ID to assign to new handlers.
	idCount uint64

	mux sync.Mutex
}

// WorkerMessage is the message that is serialised and sent to the worker.
type WorkerMessage struct {
	Tag  Tag    `json:"tag"`
	ID   uint64 `json:"id"`
	Data []byte `json:"data"`
}

// Tag describes how a message sent to or from the worker should be handled.
type Tag string

// newWorkerHandler generates a new workerHandler. This functions will only
// return once communication with the worker has been established.
func newWorkerHandler(aURL, name string) (*workerHandler, error) {
	wh := &workerHandler{
		worker: js.Global().Get("Worker").New(
			aURL, newWorkerOptions("", "", name)),
		handlers: make(map[Tag]map[uint64]handlerFn),
		idCount:  InitialID,
	}

	wh.addEventListeners()

	ready := make(chan struct{})
	wh.registerHandler(
		ReadyTag, InitialID, false, func([]byte) { ready <- struct{}{} })

	// Wait for the ready signal from the worker
	select {
	case <-ready:
	case <-time.After(workerInitialConnectionTimeout):
		return nil, errors.Errorf("timed out after %s waiting for initial "+
			"message from worker", workerInitialConnectionTimeout)
	}

	return wh, nil
}

// getNextID returns the next unique ID.
func (wh *workerHandler) getNextID() uint64 {
	id := wh.idCount
	wh.idCount++
	return id
}

// sendMessage sends a message to the worker with the given tag. If a reception
// handler is specified, then the message is given a unique ID to handle the
// reply.
func (wh *workerHandler) sendMessage(
	tag Tag, data []byte, receptionHandler handlerFn) {
	var id uint64
	if receptionHandler != nil {
		id = wh.registerHandler(tag, 0, true, receptionHandler)
	}

	message := WorkerMessage{
		Tag:  tag,
		ID:   id,
		Data: data,
	}
	payload, err := json.Marshal(message)
	if err != nil {
		jww.FATAL.Panicf(
			"Failed to marshal payload for %q going to worker: %+v", tag, err)
	}

	go wh.postMessage(string(payload))
}

// receiveMessage is registered with the Javascript event listener and is called
// every time a new message from the worker is received.
func (wh *workerHandler) receiveMessage(data []byte) error {
	var message WorkerMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return errors.Errorf(
			"could not to unmarshal payload from worker: %+v", err)
	}

	handler, err := wh.getHandler(message.Tag, message.ID)
	if err != nil {
		return err
	}

	go handler(message.Data)

	return nil
}

// registerHandler registers the handler for the given tag and ID unless autoID
// is true, in which case a unique ID is used. Returns the ID that was
// registered. If a previous handler was registered, it is overwritten.
// This function is thread safe.
func (wh *workerHandler) registerHandler(
	tag Tag, id uint64, autoID bool, handler handlerFn) uint64 {
	wh.mux.Lock()
	defer wh.mux.Unlock()

	// FIXME: This ID system only really works correctly when used with a
	//  single tag. This should probably be fixed.
	if autoID {
		id = wh.getNextID()
	}

	if _, exists := wh.handlers[tag]; !exists {
		wh.handlers[tag] = make(map[uint64]handlerFn)
	}
	wh.handlers[tag][id] = handler

	return id
}

// getHandler returns the handler with the given ID or returns an error if no
// handler is found. The handler is deleted from the map if specified in
// deleteAfterReceiving. This function is thread safe.
func (wh *workerHandler) getHandler(tag Tag, id uint64) (handlerFn, error) {
	wh.mux.Lock()
	defer wh.mux.Unlock()
	handlers, exists := wh.handlers[tag]
	if !exists {
		return nil, errors.Errorf("no handlers found for tag %q", tag)
	}

	handler, exists := handlers[id]
	if !exists {
		return nil, errors.Errorf("no %q handler found for ID %d", tag, id)
	}

	if _, exists = deleteAfterReceiving[tag]; exists {
		delete(wh.handlers[tag], id)
		if len(wh.handlers[tag]) == 0 {
			delete(wh.handlers, tag)
		}
	}

	return handler, nil
}

// postMessage sends a message to the worker.
//
// message is the object to deliver to the worker; this will be in the data
// field in the event delivered to the worker. It must be a js.Value or a
// primitive type that can be converted via js.ValueOf. The Javascript object
// must be "any value or JavaScript object handled by the structured clone
// algorithm, which includes cyclical references.". See the doc for more
// information.
//
// If the message parameter is not provided, a SyntaxError will be thrown by the
// parser. If the data to be passed to the worker is unimportant, js.Null or
// js.Undefined can be passed explicitly.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/postMessage
func (wh *workerHandler) postMessage(message any) {
	wh.worker.Call("postMessage", message)
}

// postMessageTransferList sends an array of Transferable objects to transfer to
// the worker. This is meant to be used to transfer large amounts of binary data
// using a high-performance, zero-copy operation. Refer to the doc for more
// information.
//
// Note: The binary data cannot simply be passed as the transferList. The
// traversable objects must be specified in the aMessage and included in the
// transferList
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/postMessage#transfer
func (wh *workerHandler) postMessageTransferList(message, transferList any) {
	wh.worker.Call("postMessage", message, transferList)
}

// addEventListeners adds the event listeners needed to receive a message from
// the worker. Two listeners were added; one to receive messages from the worker
// and the other to receive errors.
func (wh *workerHandler) addEventListeners() {
	// Create a listener for when the message event is fired on the worker. This
	// occurs when a message is received from the worker.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/message_event
	messageEvent := js.FuncOf(func(this js.Value, args []js.Value) any {
		err := wh.receiveMessage([]byte(args[0].Get("data").String()))
		if err != nil {
			jww.ERROR.Printf("Failed to receive message from worker: %+v", err)
		}
		return nil
	})

	// Create listener for when a messageerror event is fired on the worker.
	// This occurs when it receives a message that can't be deserialized.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/messageerror_event
	messageError := js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		jww.ERROR.Printf(
			"Error receiving message from worker: %s", utils.JsToJson(event))
		return nil
	})

	wh.worker.Call("addEventListener", "message", messageEvent)
	wh.worker.Call("addEventListener", "messageerror", messageError)
}

// newWorkerOptions creates a new Javascript object containing optional
// properties that can be set when creating a new worker.
//
// Each property is optional; leave a property empty to use the defaults (as
// documented). The available properties are:
//   - type - The type of worker to create. The value can be either "classic" or
//     "module". If not specified, the default used is classic.
//   - credentials - The type of credentials to use for the worker. The value
//     can be "omit", "same-origin", or "include". If it is not specified, or if
//     the type is "classic", then the default used is "omit" (no credentials
//     are required).
//   - name - An identifying name for the worker, used mainly for debugging
//     purposes.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Worker/Worker#options
func newWorkerOptions(workerType, credentials, name string) js.Value {
	options := js.Global().Get("Object").New()
	if workerType != "" {
		options.Set("type", workerType)
	}
	if credentials != "" {
		options.Set("credentials", credentials)
	}
	if name != "" {
		options.Set("name", name)
	}
	return options
}
