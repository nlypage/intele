package intele

import (
	"encoding/json"
	"fmt"
	tele "gopkg.in/telebot.v3"
	"time"
)

// messageCollector implements the MessageCollector interface.
type messageCollector struct {
	ctrl *controller
}

// Collect adds a message to the collector storage.
func (mc *messageCollector) Collect(messageID int, chatID int64) error {
	collectedMessage := CollectedMessage{
		MessageID: messageID,
		ChatID:    chatID,
		Timestamp: time.Now(),
	}

	// Get existing collected messages
	existingMessages, err := mc.GetMessages()
	if err != nil {
		return &ErrMessageCollection{Operation: "get existing messages", Err: err}
	}

	// Add new message
	existingMessages = append(existingMessages, collectedMessage)

	// Save back to storage
	if err := mc.saveMessages(existingMessages); err != nil {
		return &ErrMessageCollection{Operation: "save messages", Err: err}
	}

	return nil
}

// Send sends a message and automatically collects it.
func (mc *messageCollector) Send(what interface{}, opts ...interface{}) error {
	if mc.ctrl.teleCtx == nil {
		return &ErrMessageSend{ChatID: 0, Err: fmt.Errorf("no telegram context available")}
	}

	message, err := mc.ctrl.teleCtx.Bot().Send(mc.ctrl.teleCtx.Chat(), what, opts...)
	if err != nil {
		return &ErrMessageSend{ChatID: mc.ctrl.teleCtx.Chat().ID, Err: err}
	}

	return mc.Collect(message.ID, message.Chat.ID)
}

// GetMessages returns all collected messages from storage.
func (mc *messageCollector) GetMessages() ([]CollectedMessage, error) {
	messagesKey := fmt.Sprintf("collected_messages_%d_%s", mc.ctrl.session.Data.UserID, mc.ctrl.session.Flow.id)

	messagesData, exists := mc.ctrl.Storage().GetString(messagesKey)
	if !exists {
		return []CollectedMessage{}, nil
	}

	var messages []CollectedMessage
	if err := json.Unmarshal([]byte(messagesData), &messages); err != nil {
		return nil, &ErrMessageCollection{Operation: "unmarshal messages", Err: err}
	}

	return messages, nil
}

// Clear deletes all collected messages and cleans the collector.
func (mc *messageCollector) Clear(opts ClearOptions) error {
	if mc.ctrl.teleCtx == nil {
		return &ErrMessageCollection{Operation: "clear", Err: fmt.Errorf("no telegram context available")}
	}

	messages, err := mc.GetMessages()
	if err != nil {
		return err
	}

	var messagesToKeep []CollectedMessage

	for i, message := range messages {
		shouldDelete := true

		// Check if we should exclude the last message
		if opts.ExcludeLast && i == len(messages)-1 {
			shouldDelete = false
			messagesToKeep = append(messagesToKeep, message)
		}

		// Check if message is too new (MaxAge filter)
		if opts.MaxAge != nil && time.Since(message.Timestamp) < *opts.MaxAge {
			shouldDelete = false
			messagesToKeep = append(messagesToKeep, message)
		}

		if shouldDelete {
			// Create a tele.Message for deletion
			msgToDelete := &tele.Message{
				ID: message.MessageID,
				Chat: &tele.Chat{
					ID: message.ChatID,
				},
			}

			if err := mc.ctrl.teleCtx.Bot().Delete(msgToDelete); err != nil {
				deleteErr := &ErrMessageDeletion{
					MessageID: message.MessageID,
					ChatID:    message.ChatID,
					Err:       err,
				}

				return deleteErr
			}
		}
	}

	// Save remaining messages back to storage
	if err := mc.saveMessages(messagesToKeep); err != nil {
		return &ErrMessageCollection{Operation: "save remaining messages", Err: err}
	}

	return nil
}

// saveMessages saves the collected messages to storage.
func (mc *messageCollector) saveMessages(messages []CollectedMessage) error {
	messagesKey := fmt.Sprintf(
		"collected_messages_%d_%s",
		mc.ctrl.session.Data.UserID,
		mc.ctrl.session.Flow.id,
	)

	if len(messages) == 0 {
		mc.ctrl.Storage().Delete(messagesKey)
		return nil
	}

	jsonData, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	mc.ctrl.Storage().Set(messagesKey, string(jsonData))
	return nil
}
