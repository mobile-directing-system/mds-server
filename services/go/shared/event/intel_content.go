package event

// IntelTypeAnalogRadioMessage for received radio messages.
const IntelTypeAnalogRadioMessage IntelType = "analog-radio-message"

// IntelTypeAnalogRadioMessageContent is the content for intel with
// IntelTypeAnalogRadioMessage.
type IntelTypeAnalogRadioMessageContent struct {
	// Channel used for radio communication.
	Channel string `json:"channel"`
	// Callsign of the sender.
	Callsign string `json:"callsign"`
	// Head of the message.
	Head string `json:"head"`
	// Content is the actual message content.
	Content string `json:"content"`
}

// IntelTypePlaintextMessage for simple plaintext messages.
const IntelTypePlaintextMessage IntelType = "plaintext-message"

// IntelTypePlaintextMessageContent is the content for intel with
// IntelTypePlaintextMessage.
type IntelTypePlaintextMessageContent struct {
	// Text is the actual text content.
	Text string `json:"text"`
}
