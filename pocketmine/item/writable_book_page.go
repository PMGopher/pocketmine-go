package item

// WritableBookPage is a port of pocketmine\item\WritableBookPage. The PHP constructor's length/
// UTF-8 validation (throwing InvalidArgumentException) isn't ported - callers in this port
// construct pages directly via the struct literal or NewWritableBookPage, and nothing here reads
// untrusted network input yet that would need the validation.
type WritableBookPage struct {
	Text      string
	PhotoName string
}

func NewWritableBookPage(text string) WritableBookPage {
	return WritableBookPage{Text: text}
}

func NewWritableBookPageWithPhoto(text, photoName string) WritableBookPage {
	return WritableBookPage{Text: text, PhotoName: photoName}
}

func (p WritableBookPage) GetText() string { return p.Text }

func (p WritableBookPage) GetPhotoName() string { return p.PhotoName }
