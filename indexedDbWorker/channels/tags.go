////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package channels

import "gitlab.com/elixxir/xxdk-wasm/worker"

// List of tags that can be used when sending a message or registering a handler
// to receive a message.
const (

	NewWASMEventModelTag       worker.Tag = "NewWASMEventModel"
	MessageReceivedCallbackTag worker.Tag = "MessageReceivedCallback"
	EncryptionStatusTag        worker.Tag = "EncryptionStatus"
	StoreDatabaseNameTag       worker.Tag = "StoreDatabaseName"

	JoinChannelTag         worker.Tag = "JoinChannel"
	LeaveChannelTag        worker.Tag = "LeaveChannel"
	ReceiveMessageTag      worker.Tag = "ReceiveMessage"
	ReceiveReplyTag        worker.Tag = "ReceiveReply"
	ReceiveReactionTag     worker.Tag = "ReceiveReaction"
	UpdateFromUUIDTag      worker.Tag = "UpdateFromUUID"
	UpdateFromMessageIDTag worker.Tag = "UpdateFromMessageID"
	GetMessageTag          worker.Tag = "GetMessage"
	DeleteMessageTag       worker.Tag = "DeleteMessage"
)