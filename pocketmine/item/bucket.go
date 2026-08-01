package item

// Bucket is a port of pocketmine\item\Bucket (the empty bucket). OnInteractBlock (filling from a
// liquid source block) needs a real Player/Block/World - see the Item interface's doc comment.
type Bucket struct {
	ItemBase
}

func NewBucket(identifier ItemIdentifier, name string) *Bucket {
	b := &Bucket{}
	b.Init(b, identifier, name)
	return b
}

func (b *Bucket) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bucket) GetMaxStackSize() int { return 16 }
