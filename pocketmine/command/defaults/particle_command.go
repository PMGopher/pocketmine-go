package defaults

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
)

// ParticleCommand is a port of pocketmine\command\defaults\ParticleCommand.
type ParticleCommand struct{ VanillaCommand }

func NewParticleCommand() *ParticleCommand {
	return &ParticleCommand{newVanillaCommand("particle", lang.KnownTranslationFactory.PocketmineCommandParticleDescription(), lang.KnownTranslationFactory.PocketmineCommandParticleUsage(), nil, permission.CommandParticle)}
}

func (c *ParticleCommand) Execute(sender command.Sender, commandLabel string, args []string) (any, error) {
	if len(args) < 7 {
		return syntaxError()
	}

	var w *world.World
	var pos math.Vector3
	if p, ok := sender.(*player.Player); ok {
		senderPos := p.GetPosition()
		w = p.GetWorld()
		pos = math.NewVector3(
			getRelativeDouble(senderPos.X, args[1], MinCoord, MaxCoord),
			getRelativeDouble(senderPos.Y, args[2], world.YMin, world.YMax),
			getRelativeDouble(senderPos.Z, args[3], MinCoord, MaxCoord),
		)
	} else {
		w = srv(sender).GetWorldManager().GetDefaultWorld()
		pos = math.NewVector3(phpFloat(args[1]), phpFloat(args[2]), phpFloat(args[3]))
	}

	name := strings.ToLower(args[0])

	xd := phpFloat(args[4])
	yd := phpFloat(args[5])
	zd := phpFloat(args[6])

	count := 1
	if len(args) > 7 {
		count = max(1, phpInt(args[7]))
	}

	var data *string
	if len(args) > 8 {
		data = &args[8]
	}

	p := getParticle(name, data)
	if p == nil {
		sender.SendMessage(lang.KnownTranslationFactory.CommandsParticleNotFound(name).Prefix(utils.Red))
		return true, nil
	}

	sender.SendMessage(lang.KnownTranslationFactory.CommandsParticleSuccess(name, strconv.Itoa(count)))

	random := utils.NewRandom(int(time.Now().UnixMilli()) + rand.Intn(1<<31))

	for i := 0; i < count; i++ {
		w.AddParticle(pos.Add(
			random.NextSignedFloat()*xd,
			random.NextSignedFloat()*yd,
			random.NextSignedFloat()*zd,
		), p)
	}
	return true, nil
}

// dataInt is `(int) ($data ?? $default)`.
func dataInt(data *string, def int) int {
	if data == nil {
		return def
	}
	return phpInt(*data)
}

// parsedBlock is `StringToItemParser::getInstance()->parse($data)?->getBlock()`, excluding air.
func parsedBlock(data *string) block.Behavior {
	if data == nil {
		return nil
	}
	it, ok := item.GetStringToItemParser().Parse(*data)
	if !ok {
		return nil
	}
	if b := it.GetBlock(); b.GetTypeId() != block.AIR {
		return b
	}
	return nil
}

// parsedItem is `StringToItemParser::getInstance()->parse($data)`, excluding null items.
func parsedItem(data *string) item.Item {
	if data == nil {
		return nil
	}
	if it, ok := item.GetStringToItemParser().Parse(*data); ok && !it.IsNull() {
		return it
	}
	return nil
}

// getParticle is a port of ParticleCommand::getParticle.
func getParticle(name string, data *string) particle.Particle {
	switch name {
	case "explode":
		return particle.ExplodeParticle{}
	case "hugeexplosion":
		return particle.HugeExplodeParticle{}
	case "hugeexplosionseed":
		return particle.HugeExplodeSeedParticle{}
	case "bubble":
		return particle.BubbleParticle{}
	case "splash":
		return particle.SplashParticle{}
	case "wake", "water":
		return particle.WaterParticle{}
	case "crit":
		return particle.CriticalParticle{Scale: 2}
	case "smoke":
		return particle.SmokeParticle{Scale: dataInt(data, 0)}
	case "spell":
		return particle.EnchantParticle{Color: color.NewColor(0, 0, 0, 255)} //TODO: colour support
	case "instantspell":
		return particle.InstantEnchantParticle{Color: color.NewColor(0, 0, 0, 255)} //TODO: colour support
	case "dripwater":
		return particle.WaterDripParticle{}
	case "driplava":
		return particle.LavaDripParticle{}
	case "townaura", "spore":
		return particle.SporeParticle{}
	case "portal":
		return particle.PortalParticle{}
	case "flame":
		return particle.FlameParticle{}
	case "lava":
		return particle.LavaParticle{}
	case "reddust":
		return particle.RedstoneParticle{Lifetime: dataInt(data, 1)}
	case "snowballpoof":
		return convert.NewItemBreakParticle(item.VanillaItem("snowball"))
	case "slime":
		return convert.NewItemBreakParticle(item.VanillaItem("slimeball"))
	case "itembreak", "iconcrack":
		if it := parsedItem(data); it != nil {
			return convert.NewItemBreakParticle(it)
		}
	case "terrain", "blockcrack":
		if b := parsedBlock(data); b != nil {
			return particle.TerrainParticle{BlockStateID: b.GetStateId()}
		}
	case "heart":
		return particle.HeartParticle{Scale: dataInt(data, 0)}
	case "ink":
		return particle.InkParticle{Scale: dataInt(data, 0)}
	case "droplet":
		return particle.RainSplashParticle{}
	case "enchantmenttable":
		return particle.EnchantmentTableParticle{}
	case "happyvillager":
		return particle.HappyVillagerParticle{}
	case "angryvillager":
		return particle.AngryVillagerParticle{}
	case "forcefield":
		return particle.BlockForceFieldParticle{Data: dataInt(data, 0)}
	case "mobflame":
		return particle.EntityFlameParticle{}
	case "blockdust":
		if data != nil {
			//to preserve the old unlimited explode behaviour, allow this to split into at most 5 parts
			//this allows the 4th argument to be processed normally if given without forcing it to also consume
			//any unexpected parts
			//we probably ought to error in this case, but this will do for now
			d := strings.SplitN(*data, "_", 5)
			if len(d) >= 3 {
				a := 255
				if len(d) > 3 {
					a = phpInt(d[3])
				}
				return particle.DustParticle{Color: color.NewColor(uint8(phpInt(d[0])&0xff), uint8(phpInt(d[1])&0xff), uint8(phpInt(d[2])&0xff), uint8(a&0xff))}
			}
		}
	case "sonicexplosion":
		return particle.SonicExplosionParticle{}
	}
	return nil
}
