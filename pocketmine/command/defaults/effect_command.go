package defaults

import (
	stdmath "math"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
)

// EffectCommand is a port of pocketmine\command\defaults\EffectCommand.
type EffectCommand struct{ VanillaCommand }

func NewEffectCommand() *EffectCommand {
	c := &EffectCommand{VanillaCommand{Command: command.InitCommand("effect", lang.KnownTranslationFactory.PocketmineCommandEffectDescription(), lang.KnownTranslationFactory.CommandsEffectUsage(), nil)}}
	_ = c.SetPermissions([]string{permission.CommandEffectSelf, permission.CommandEffectOther})
	return c
}

func (c *EffectCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 2 {
		return syntaxError()
	}

	p, err := c.fetchPermittedPlayerTarget(sender, &args[0], permission.CommandEffectSelf, permission.CommandEffectOther)
	if err != nil || p == nil {
		return true, err
	}
	effectManager := p.GetEffects()

	if strings.ToLower(args[1]) == "clear" {
		effectManager.Clear()

		sender.SendMessage(lang.KnownTranslationFactory.CommandsEffectSuccessRemovedAll(p.GetDisplayName()))
		return true, nil
	}

	eff, ok := effect.GetStringToEffectParser().Parse(args[1])
	if !ok {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsEffectNotFound(args[1]).Prefix(utils.Red))
		return true, nil
	}

	amplification := 0
	infinite := false
	var duration *int

	if len(args) >= 3 {
		if strings.ToLower(args[2]) == "infinite" {
			infinite = true
		} else {
			d, err := getBoundedInt(sender, args[2], 0, stdmath.MaxInt32/20)
			if err != nil || d == nil {
				return false, err
			}
			ticks := *d * 20 // ticks
			duration = &ticks
		}
	}

	if len(args) >= 4 {
		a, err := getBoundedInt(sender, args[3], 0, 255)
		if err != nil || a == nil {
			return false, err
		}
		amplification = *a
	}

	visible := true
	if len(args) >= 5 {
		v := strings.ToLower(args[4])
		if v == "on" || v == "true" || v == "t" || v == "1" {
			visible = false
		}
	}

	if duration != nil && *duration == 0 {
		if !effectManager.Has(eff) {
			if len(effectManager.All()) == 0 {
				sender.SendMessage(lang.KnownTranslationFactory.CommandsEffectFailureNotActiveAll(p.GetDisplayName()))
			} else {
				sender.SendMessage(lang.KnownTranslationFactory.CommandsEffectFailureNotActive(eff.GetName(), p.GetDisplayName()))
			}
			return true, nil
		}

		effectManager.Remove(eff)
		sender.SendMessage(lang.KnownTranslationFactory.CommandsEffectSuccessRemoved(eff.GetName(), p.GetDisplayName()))
	} else {
		instance := effect.NewEffectInstanceFull(eff, duration, amplification, visible, false, nil, infinite)
		effectManager.Add(instance)

		if infinite {
			command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsEffectSuccessInfinite(eff.GetName(), strconv.Itoa(instance.GetAmplifier()), p.GetDisplayName()), true)
		} else {
			command.BroadcastCommandMessage(sender, lang.KnownTranslationFactory.CommandsEffectSuccess(eff.GetName(), strconv.Itoa(instance.GetAmplifier()), p.GetDisplayName(), phpDivide(instance.GetDuration(), 20)), true)
		}
	}

	return true, nil
}

// phpDivide is `(string) ($a / $b)`: PHP's `/` gives an int when the division is exact and a
// float otherwise.
func phpDivide(a, b int) string {
	if a%b == 0 {
		return strconv.Itoa(a / b)
	}
	return strconv.FormatFloat(float64(a)/float64(b), 'f', -1, 64)
}
