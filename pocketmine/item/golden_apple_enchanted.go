package item

// GoldenAppleEnchanted is a port of pocketmine\item\GoldenAppleEnchanted. Its only override in
// PHP is GetAdditionalEffects (a stronger effect set than GoldenApple's) - not ported, same gap
// as GoldenApple's doc comment - so this behaves identically to GoldenApple here, existing purely
// as its own named type (matching the PHP class hierarchy) rather than a functional difference.
type GoldenAppleEnchanted struct {
	GoldenApple
}

func NewGoldenAppleEnchanted(identifier ItemIdentifier, name string) *GoldenAppleEnchanted {
	g := &GoldenAppleEnchanted{}
	g.Init(g, identifier, name)
	return g
}

func (g *GoldenAppleEnchanted) Clone() Item {
	c := *g
	c.rebind(&c)
	return &c
}
