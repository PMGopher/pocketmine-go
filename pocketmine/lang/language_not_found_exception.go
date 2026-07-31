package lang

// LanguageNotFoundException is a port of pocketmine\lang\LanguageNotFoundException.
type LanguageNotFoundException struct{ Message string }

func (e *LanguageNotFoundException) Error() string { return e.Message }
