// generated by stringer -type=itemType; DO NOT EDIT

package biblexer

import "fmt"

const _itemType_name = "itemErroritemEOFitemCommentitemEntryTypeDelimitemEntryTypeitemEntryStartDelimitemEntryStopDelimitemCiteKeyitemTagNameitemTagNameContentDelimitemTagContentitemTagDelimitemTagContentStartDelimitemTagContentStopDelimitemTagContentQuoteDelimitemConcatitemStringKeyitemStringDelim"

var _itemType_index = [...]uint16{0, 9, 16, 27, 45, 58, 77, 95, 106, 117, 140, 154, 166, 190, 213, 237, 247, 260, 275}

func (i itemType) String() string {
	if i < 0 || i >= itemType(len(_itemType_index)-1) {
		return fmt.Sprintf("itemType(%d)", i)
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}