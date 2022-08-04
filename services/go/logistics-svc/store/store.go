package store

import (
	"github.com/doug-martin/goqu/v9"
	"log"
)

// Mall provides all store access methods.
type Mall struct {
	dialect          goqu.DialectWrapper
	channelOperators map[ChannelType]channelOperator
}

// ChannelTypeSupplier is the global channelOperatorSupplier that is
// used when creating Mall with NewMall and also for validation of supported
// channel types. Will be set in init.
var ChannelTypeSupplier channelTypesSupplier

func init() {
	ChannelTypeSupplier = channelTypesSupplier{
		ChannelTypes: map[ChannelType]struct{}{
			ChannelTypeDirect:         {},
			ChannelTypeEmail:          {},
			ChannelTypeForwardToGroup: {},
			ChannelTypeForwardToUser:  {},
			ChannelTypePhoneCall:      {},
			ChannelTypeRadio:          {},
			ChannelTypePush:           {},
		},
	}
	_ = ChannelTypeSupplier.operators(nil)
}

// channelTypesSupplier is a central supplier for channel operators. As is
// holds a list of supported channel types and is provided via a global instance
// (ChannelTypeSupplier), it is also used in Channel.Validate.
type channelTypesSupplier struct {
	// ChannelTypes holds all supported channel types. They are held in a map in
	// order to provide fast access for entity validation.
	ChannelTypes map[ChannelType]struct{}
}

func (supplier channelTypesSupplier) operators(m *Mall) map[ChannelType]channelOperator {
	operators := make(map[ChannelType]channelOperator, len(supplier.ChannelTypes))
	for channelType := range supplier.ChannelTypes {
		var operator channelOperator
		switch channelType {
		case ChannelTypeDirect:
			operator = &directChannelOperator{m: m}
		case ChannelTypeEmail:
			operator = &emailChannelOperator{m: m}
		case ChannelTypeForwardToGroup:
			operator = &forwardToGroupChannelOperator{m: m}
		case ChannelTypeForwardToUser:
			operator = &forwardToUserChannelOperator{m: m}
		case ChannelTypePhoneCall:
			operator = &phoneCallChannelOperator{m: m}
		case ChannelTypeRadio:
			operator = &radioChannelOperator{m: m}
		case ChannelTypePush:
			operator = &pushChannelOperator{}
		default:
			log.Fatalf("missing channel operator for channel type %v", channelType)
		}
		operators[channelType] = operator
	}
	return operators
}

// NewMall creates a new Mall with postgres dialect.
func NewMall() *Mall {
	m := &Mall{
		dialect: goqu.Dialect("postgres"),
	}
	m.channelOperators = ChannelTypeSupplier.operators(m)
	return m
}
