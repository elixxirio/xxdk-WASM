////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/hack-pad/go-indexeddb/idb"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	wChannels "gitlab.com/elixxir/xxdk-wasm/indexedDb/worker/channels"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/xx_network/primitives/id"
)

// wasmModel implements [channels.EventModel] interface, which uses the channels
// system passed an object that adheres to in order to get events on the
// channel.
type wasmModel struct {
	db                *idb.Database
	cipher            cryptoChannel.Cipher
	receivedMessageCB wChannels.MessageReceivedCallback
	deletedMessageCB  wChannels.DeletedMessageCallback
	mutedUserCB       wChannels.MutedUserCallback
	updateMux         sync.Mutex
}

// JoinChannel is called whenever a channel is joined locally.
func (w *wasmModel) JoinChannel(channel *cryptoBroadcast.Channel) {
	parentErr := errors.New("failed to JoinChannel")

	// Build object
	newChannel := Channel{
		ID:          channel.ReceptionID.Marshal(),
		Name:        channel.Name,
		Description: channel.Description,
	}

	// Convert to jsObject
	newChannelJson, err := json.Marshal(&newChannel)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to marshal Channel: %+v", err))
		return
	}
	channelObj, err := utils.JsonToJS(newChannelJson)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to marshal Channel: %+v", err))
		return
	}

	_, err = impl.Put(w.db, channelsStoreName, channelObj)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to put Channel: %+v", err))
	}
}

// LeaveChannel is called whenever a channel is left locally.
func (w *wasmModel) LeaveChannel(channelID *id.ID) {
	parentErr := errors.New("failed to LeaveChannel")

	// Delete the channel from storage
	err := impl.Delete(w.db, channelsStoreName, js.ValueOf(channelID.String()))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to delete Channel: %+v", err))
		return
	}

	// Clean up lingering data
	err = w.deleteMsgByChannel(channelID)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Deleting Channel's Message data failed: %+v", err))
		return
	}
	jww.DEBUG.Printf("Successfully deleted channel: %s", channelID)
}

// deleteMsgByChannel is a private helper that uses messageStoreChannelIndex
// to delete all Message with the given Channel ID.
func (w *wasmModel) deleteMsgByChannel(channelID *id.ID) error {
	parentErr := errors.New("failed to deleteMsgByChannel")

	// Prepare the Transaction
	txn, err := w.db.Transaction(idb.TransactionReadWrite, messageStoreName)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to create Transaction: %+v", err)
	}
	store, err := txn.ObjectStore(messageStoreName)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to get ObjectStore: %+v", err)
	}
	index, err := store.Index(messageStoreChannelIndex)
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to get Index: %+v", err)
	}

	// Perform the operation
	channelIdStr := base64.StdEncoding.EncodeToString(channelID.Marshal())
	keyRange, err := idb.NewKeyRangeOnly(js.ValueOf(channelIdStr))
	cursorRequest, err := index.OpenCursorRange(keyRange, idb.CursorNext)
	if err != nil {
		return errors.WithMessagef(parentErr, "Unable to open Cursor: %+v", err)
	}
	ctx, cancel := impl.NewContext()
	err = cursorRequest.Iter(ctx,
		func(cursor *idb.CursorWithValue) error {
			_, err := cursor.Delete()
			return err
		})
	cancel()
	if err != nil {
		return errors.WithMessagef(parentErr,
			"Unable to delete Message data: %+v", err)
	}
	return nil
}

// ReceiveMessage is called whenever a message is received on a given channel.
//
// It may be called multiple times on the same message; it is incumbent on the
// user of the API to filter such called by message ID.
func (w *wasmModel) ReceiveMessage(channelID *id.ID, messageID message.ID,
	nickname, text string, pubKey ed25519.PublicKey, dmToken uint32,
	codeset uint8, timestamp time.Time, lease time.Duration, round rounds.Round,
	mType channels.MessageType, status channels.SentStatus, hidden bool) uint64 {
	textBytes := []byte(text)
	var err error

	// Handle encryption, if it is present
	if w.cipher != nil {
		textBytes, err = w.cipher.Encrypt([]byte(text))
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	msgToInsert := buildMessage(
		channelID.Marshal(), messageID.Bytes(), nil, nickname,
		textBytes, pubKey, dmToken, codeset, timestamp, lease, round.ID, mType,
		false, hidden, status)

	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		jww.ERROR.Printf("Failed to receive Message: %+v", err)
	}

	go w.receivedMessageCB(uuid, channelID, false)
	return uuid
}

// ReceiveReply is called whenever a message is received that is a reply on a
// given channel. It may be called multiple times on the same message; it is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply, in theory, can arrive before
// the initial message. As a result, it may be important to buffer replies.
func (w *wasmModel) ReceiveReply(channelID *id.ID, messageID,
	replyTo message.ID, nickname, text string, pubKey ed25519.PublicKey,
	dmToken uint32, codeset uint8, timestamp time.Time, lease time.Duration,
	round rounds.Round, mType channels.MessageType, status channels.SentStatus,
	hidden bool) uint64 {
	textBytes := []byte(text)
	var err error

	// Handle encryption, if it is present
	if w.cipher != nil {
		textBytes, err = w.cipher.Encrypt([]byte(text))
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	msgToInsert := buildMessage(channelID.Marshal(), messageID.Bytes(),
		replyTo.Bytes(), nickname, textBytes, pubKey, dmToken, codeset,
		timestamp, lease, round.ID, mType, hidden, false, status)

	uuid, err := w.receiveHelper(msgToInsert, false)

	if err != nil {
		jww.ERROR.Printf("Failed to receive reply: %+v", err)
	}
	go w.receivedMessageCB(uuid, channelID, false)
	return uuid
}

// ReceiveReaction is called whenever a reaction to a message is received on a
// given channel. It may be called multiple times on the same reaction; it is
// incumbent on the user of the API to filter such called by message ID.
//
// Messages may arrive our of order, so a reply, in theory, can arrive before
// the initial message. As a result, it may be important to buffer reactions.
func (w *wasmModel) ReceiveReaction(channelID *id.ID, messageID,
	reactionTo message.ID, nickname, reaction string, pubKey ed25519.PublicKey,
	dmToken uint32, codeset uint8, timestamp time.Time, lease time.Duration,
	round rounds.Round, mType channels.MessageType, status channels.SentStatus,
	hidden bool) uint64 {
	textBytes := []byte(reaction)
	var err error

	// Handle encryption, if it is present
	if w.cipher != nil {
		textBytes, err = w.cipher.Encrypt([]byte(reaction))
		if err != nil {
			jww.ERROR.Printf("Failed to encrypt Message: %+v", err)
			return 0
		}
	}

	msgToInsert := buildMessage(
		channelID.Marshal(), messageID.Bytes(), reactionTo.Bytes(), nickname,
		textBytes, pubKey, dmToken, codeset, timestamp, lease, round.ID, mType,
		false, hidden, status)

	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		jww.ERROR.Printf("Failed to receive reaction: %+v", err)
	}
	go w.receivedMessageCB(uuid, channelID, false)
	return uuid
}

// UpdateFromUUID is called whenever a message at the UUID is modified.
//
// messageID, timestamp, round, pinned, and hidden are all nillable and may be
// updated based upon the UUID at a later date. If a nil value is passed, then
// make no update.
func (w *wasmModel) UpdateFromUUID(uuid uint64, messageID *message.ID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) {
	parentErr := errors.New("failed to UpdateFromUUID")

	// FIXME: this is a bit of race condition without the mux.
	//        This should be done via the transactions (i.e., make a
	//        special version of receiveHelper)
	w.updateMux.Lock()
	defer w.updateMux.Unlock()

	// Convert messageID to the key generated by json.Marshal
	key := js.ValueOf(uuid)

	// Use the key to get the existing Message
	currentMsg, err := impl.Get(w.db, messageStoreName, key)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to get message: %+v", err))
		return
	}

	_, err = w.updateMessage(utils.JsToJson(currentMsg), messageID, nil, timestamp,
		round, pinned, hidden, status, nil)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to updateMessage: %+v", err))
	}
}

// UpdateFromMessageID is called whenever a message with the message ID is
// modified.
//
// The API needs to return the UUID of the modified message that can be
// referenced at a later time.
//
// timestamp, round, pinned, and hidden are all nillable and may be updated
// based upon the UUID at a later date. If a nil value is passed, then make
// no update.
func (w *wasmModel) UpdateFromMessageID(messageID message.ID,
	timestamp *time.Time, round *rounds.Round, pinned, hidden *bool,
	status *channels.SentStatus) uint64 {
	parentErr := errors.New("failed to UpdateFromMessageID")

	// FIXME: this is a bit of race condition without the mux.
	//        This should be done via the transactions (i.e., make a
	//        special version of receiveHelper)
	w.updateMux.Lock()
	defer w.updateMux.Unlock()

	msgIDStr := base64.StdEncoding.EncodeToString(messageID.Marshal())
	currentMsgObj, err := impl.GetIndex(w.db, messageStoreName,
		messageStoreMessageIndex, js.ValueOf(msgIDStr))
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Failed to get message by index: %+v", err))
		return 0
	}

	currentMsg := utils.JsToJson(currentMsgObj)
	uuid, err := w.updateMessage(currentMsg, &messageID, nil, timestamp,
		round, pinned, hidden, status, nil)
	if err != nil {
		jww.ERROR.Printf("%+v", errors.WithMessagef(parentErr,
			"Unable to updateMessage: %+v", err))
	}
	return uuid
}

// updateMessage is a helper for updating a stored message.
func (w *wasmModel) updateMessage(currentMsgJson string, messageID *message.ID,
	messageData []byte, timestamp *time.Time, round *rounds.Round, pinned,
	hidden *bool, status *channels.SentStatus, fileData []byte) (uint64, error) {

	newMessage := &Message{}
	err := json.Unmarshal([]byte(currentMsgJson), newMessage)
	if err != nil {
		return 0, err
	}

	if status != nil {
		newMessage.Status = uint8(*status)
	}
	if messageID != nil {
		newMessage.MessageID = messageID.Bytes()
	}

	if round != nil {
		newMessage.Round = uint64(round.ID)
	}

	if timestamp != nil {
		newMessage.Timestamp = *timestamp
	}

	if pinned != nil {
		newMessage.Pinned = *pinned
	}

	if hidden != nil {
		newMessage.Hidden = *hidden
	}

	if messageData != nil {
		newMessage.Text = messageData
	}

	if fileData != nil {
		newMessage.FileData = fileData
	}

	// Store the updated Message
	uuid, err := w.receiveHelper(newMessage, true)
	if err != nil {
		return 0, err
	}
	channelID := &id.ID{}
	copy(channelID[:], newMessage.ChannelID)
	go w.receivedMessageCB(uuid, channelID, true)

	return uuid, nil
}

// buildMessage is a private helper that converts typical [channels.EventModel]
// inputs into a basic Message structure for insertion into storage.
//
// NOTE: ID is not set inside this function because we want to use the
// autoincrement key by default. If you are trying to overwrite an existing
// message, then you need to set it manually yourself.
func buildMessage(channelID, messageID, parentID []byte, nickname string,
	text []byte, pubKey ed25519.PublicKey, dmToken uint32, codeset uint8,
	timestamp time.Time, lease time.Duration, round id.Round,
	mType channels.MessageType, pinned, hidden bool,
	status channels.SentStatus) *Message {
	return &Message{
		MessageID:       messageID,
		Nickname:        nickname,
		ChannelID:       channelID,
		ParentMessageID: parentID,
		Timestamp:       timestamp,
		Lease:           lease,
		Status:          uint8(status),
		Hidden:          hidden,
		Pinned:          pinned,
		Text:            text,
		Type:            uint16(mType),
		Round:           uint64(round),
		// User Identity Info
		Pubkey:         pubKey,
		DmToken:        dmToken,
		CodesetVersion: codeset,
		FileData:       nil,
		FileId:         nil,
	}
}

// receiveHelper is a private helper for receiving any sort of message.
func (w *wasmModel) receiveHelper(
	newMessage *Message, isUpdate bool) (uint64, error) {
	// Convert to jsObject
	newMessageJson, err := json.Marshal(newMessage)
	if err != nil {
		return 0, errors.Errorf("Unable to marshal Message: %+v", err)
	}
	messageObj, err := utils.JsonToJS(newMessageJson)
	if err != nil {
		return 0, errors.Errorf("Unable to marshal Message: %+v", err)
	}

	// Unset the primaryKey for inserts so that it can be auto-populated and
	// incremented
	if !isUpdate {
		messageObj.Delete("id")
	}

	// Store message to database
	result, err := impl.Put(w.db, messageStoreName, messageObj)
	if err != nil && !strings.Contains(err.Error(),
		"at least one key does not satisfy the uniqueness requirements") {
		// Only return non-unique constraint errors so that the case
		// below this one can be hit and handle duplicate entries properly.
		return 0, errors.Errorf("Unable to put Message: %+v", err)
	}

	// NOTE: Sometimes the insert fails to return an error but hits a duplicate
	//  insert, so this fallthrough returns the UUID entry in that case.
	if result.IsUndefined() {
		msgID := message.ID{}
		copy(msgID[:], newMessage.MessageID)
		msg, errLookup := w.msgIDLookup(msgID)
		if errLookup == nil && msg.ID != 0 {
			return msg.ID, nil
		}
		return 0, errors.Errorf("uuid lookup failure: %+v", err)
	}
	uuid := uint64(result.Int())
	jww.DEBUG.Printf("Successfully stored message %d", uuid)

	return uuid, nil
}

// GetMessage returns the message with the given [channel.MessageID].
func (w *wasmModel) GetMessage(
	messageID message.ID) (channels.ModelMessage, error) {
	lookupResult, err := w.msgIDLookup(messageID)
	if err != nil {
		return channels.ModelMessage{}, err
	}

	var channelId *id.ID
	if lookupResult.ChannelID != nil {
		channelId, err = id.Unmarshal(lookupResult.ChannelID)
		if err != nil {
			return channels.ModelMessage{}, err
		}
	}

	var parentMsgId message.ID
	if lookupResult.ParentMessageID != nil {
		parentMsgId, err = message.UnmarshalID(lookupResult.ParentMessageID)
		if err != nil {
			return channels.ModelMessage{}, err
		}
	}

	return channels.ModelMessage{
		UUID:            lookupResult.ID,
		Nickname:        lookupResult.Nickname,
		MessageID:       messageID,
		ChannelID:       channelId,
		ParentMessageID: parentMsgId,
		Timestamp:       lookupResult.Timestamp,
		Lease:           lookupResult.Lease,
		Status:          channels.SentStatus(lookupResult.Status),
		Hidden:          lookupResult.Hidden,
		Pinned:          lookupResult.Pinned,
		Content:         lookupResult.Text,
		Type:            channels.MessageType(lookupResult.Type),
		Round:           id.Round(lookupResult.Round),
		PubKey:          lookupResult.Pubkey,
		CodesetVersion:  lookupResult.CodesetVersion,
	}, nil
}

// DeleteMessage removes a message with the given messageID from storage.
func (w *wasmModel) DeleteMessage(messageID message.ID) error {
	msgId := js.ValueOf(base64.StdEncoding.EncodeToString(messageID.Bytes()))

	err := impl.DeleteIndex(
		w.db, messageStoreName, messageStoreMessageIndex, pkeyName, msgId)
	if err != nil {
		return err
	}

	go w.deletedMessageCB(messageID)

	return nil
}

// MuteUser is called whenever a user is muted or unmuted.
func (w *wasmModel) MuteUser(
	channelID *id.ID, pubKey ed25519.PublicKey, unmute bool) {
	go w.mutedUserCB(channelID, pubKey, unmute)
}

// msgIDLookup gets the UUID of the Message with the given messageID.
func (w *wasmModel) msgIDLookup(messageID message.ID) (*Message, error) {
	msgIDStr := js.ValueOf(base64.StdEncoding.EncodeToString(messageID.Bytes()))
	resultObj, err := impl.GetIndex(w.db, messageStoreName,
		messageStoreMessageIndex, msgIDStr)
	if err != nil {
		return nil, err
	} else if resultObj.IsUndefined() {
		return nil, errors.Errorf("no message for %s found", msgIDStr)
	}

	// Process result into string
	resultMsg := &Message{}
	err = json.Unmarshal([]byte(utils.JsToJson(resultObj)), resultMsg)
	if err != nil {
		return nil, err
	}
	return resultMsg, nil
}

// ReceiveFileMessage is called when a file upload begins or when a message
// to download a file is received.
//
// The API needs to return a UUID of the message that can be referenced at a
// later time.
//
// timestamp and round are nillable and may be updated based upon the UUID
// at a later date. A time of time.Time{} will be passed for a nilled
// timestamp.
//
// nickname may be empty, in which case the UI is expected to display the
// codename.
func (w *wasmModel) ReceiveFileMessage(channelID *id.ID, fileID fileTransfer.ID, nickname string,
	fileInfo, fileData []byte, pubKey ed25519.PublicKey, dmToken uint32,
	codeset uint8, timestamp time.Time, lease time.Duration,
	round rounds.Round, messageType channels.MessageType, status channels.SentStatus,
	hidden bool) (uint64, error) {
	var err error

	// Handle encryption, if it is present
	if w.cipher != nil {
		fileInfo, err = w.cipher.Encrypt(fileInfo)
		if err != nil {
			return 0, errors.Errorf("Failed to encrypt fileInfo: %+v", err)
		}
		fileData, err = w.cipher.Encrypt(fileData)
		if err != nil {
			return 0, errors.Errorf("Failed to encrypt fileData: %+v", err)
		}
	}

	msgToInsert := buildMessage(
		channelID.Marshal(), nil, nil, nickname,
		fileInfo, pubKey, dmToken, codeset, timestamp, lease, round.ID, messageType,
		false, hidden, status)
	msgToInsert.FileId = fileID.Marshal()
	msgToInsert.FileData = fileData

	uuid, err := w.receiveHelper(msgToInsert, false)
	if err != nil {
		return 0, errors.Errorf("Failed to receive File Message: %+v", err)
	}
	return uuid, nil
}

// UpdateFile is called when a file upload completed, a download starts, or
// a download completes. Each use will be identified in a SentStatus change
// (SendProcessingComplete, ReceptionProcessing, and
// ReceptionProcessingComplete).
//
// timestamp, round, pinned, hidden, and status are all nillable and may be
// updated based upon the fileID at a later date. If a nil value is passed,
// then make no update.
func (w *wasmModel) UpdateFile(fileID fileTransfer.ID, fileInfo, fileData []byte, timestamp *time.Time,
	round *rounds.Round, pinned, hidden *bool, status *channels.SentStatus) error {
	parentErr := "failed to UpdateFile"

	currentMsg, err := impl.GetIndex(w.db, messageStoreName, messageStoreFileIndex,
		js.ValueOf(fileID.Marshal()))
	if err != nil {
		return errors.WithMessage(err, parentErr)
	}

	// Handle encryption, if it is present
	if w.cipher != nil {
		if fileInfo != nil {
			fileInfo, err = w.cipher.Encrypt(fileInfo)
			if err != nil {
				return errors.WithMessage(errors.Errorf("Failed to encrypt fileInfo: %+v", err), parentErr)

			}
		}
		if fileData != nil {
			fileData, err = w.cipher.Encrypt(fileData)
			if err != nil {
				return errors.WithMessage(errors.Errorf("Failed to encrypt fileData: %+v", err), parentErr)
			}
		}
	}

	_, err = w.updateMessage(utils.JsToJson(currentMsg), nil, fileInfo,
		timestamp, round, pinned, hidden, status, fileData)
	if err != nil {
		return errors.WithMessage(err, parentErr)
	}
	return nil
}

// GetFile returns the file data and info at the given file ID.
func (w *wasmModel) GetFile(fileID fileTransfer.ID) (fileInfo, fileData []byte, err error) {
	resultObj, err := impl.GetIndex(w.db, messageStoreName, messageStoreFileIndex,
		js.ValueOf(fileID.Marshal()))
	if err != nil {
		return nil, nil, err
	}

	resultMsg := &Message{}
	err = json.Unmarshal([]byte(utils.JsToJson(resultObj)), resultMsg)
	if err != nil {
		return nil, nil, err
	}
	return resultMsg.Text, resultMsg.FileData, nil
}
