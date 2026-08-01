package item

import blockutils "pocketmine-go/pocketmine/block/utils"

// Record is a port of pocketmine\item\Record.
type Record struct {
	ItemBase

	RecordTypeValue blockutils.RecordType
}

func NewRecord(identifier ItemIdentifier, recordType blockutils.RecordType, name string) *Record {
	r := &Record{RecordTypeValue: recordType}
	r.Init(r, identifier, name)
	return r
}

func (r *Record) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *Record) GetRecordType() blockutils.RecordType { return r.RecordTypeValue }

func (r *Record) GetMaxStackSize() int { return 1 }
