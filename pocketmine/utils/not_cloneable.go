package utils

// NotCloneable is a port of pocketmine\utils\NotCloneable.
//
// PHP's trait throws at runtime when a `clone`d object is created. Go has no clone operator
// or copy-constructor hook, so there is no way to intercept a struct value copy at runtime.
// The closest available equivalent is embedding this marker (the same trick sync.Mutex uses):
// it makes the containing type implement sync.Locker with no-op methods, which causes
// `go vet`'s copylocks check to flag accidental copies at compile-lint time. This is a
// build-time approximation, not the hard runtime guarantee the PHP original provides.
type NotCloneable struct{}

func (*NotCloneable) Lock()   {}
func (*NotCloneable) Unlock() {}
