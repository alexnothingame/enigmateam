
package model

// Outgoing message instruction from Bot Logic to Telegram Client.
type OutMessage struct {
  ChatID int64 `json:"chat_id"`
  Text   string `json:"text"`
  // Inline keyboard optional
  Keyboard *InlineKeyboard `json:"keyboard,omitempty"`
  ParseMode string `json:"parse_mode,omitempty"` // "MarkdownV2" etc
  DisablePreview bool `json:"disable_preview,omitempty"`
}

type InlineKeyboard struct {
  Rows [][]InlineButton `json:"rows"`
}

type InlineButton struct {
  Text string `json:"text"`
  Data string `json:"data"`
}

// Bot Logic response to Telegram Client for one Update or cron tick.
type Response struct {
  Messages []OutMessage `json:"messages"`
  // CallbackAnswer may be set to stop loading indicator in Telegram.
  CallbackAnswer *CallbackAnswer `json:"callback_answer,omitempty"`
}

type CallbackAnswer struct {
  CallbackQueryID string `json:"callback_query_id"`
  Text string `json:"text,omitempty"`
  ShowAlert bool `json:"show_alert,omitempty"`
}
