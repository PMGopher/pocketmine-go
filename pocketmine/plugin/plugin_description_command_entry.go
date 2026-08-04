package plugin

// DescriptionCommandEntry is a port of pocketmine\plugin\PluginDescriptionCommandEntry.
// Description/UsageMessage/PermissionDeniedMessage use *string for PHP's ?string (nil for
// "absent"), matching this port's established convention for nullable strings (see e.g.
// command.Command's own PermissionMessage *string).
type DescriptionCommandEntry struct {
	Description             *string
	UsageMessage            *string
	Aliases                 []string
	Permission              string
	PermissionDeniedMessage *string
}
