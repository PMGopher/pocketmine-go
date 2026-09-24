package entity

import (
	stdmath "math"
	"math/rand/v2"
	"sort"

	"pocketmine-go/pocketmine/binaryutils"
	entityutils "pocketmine-go/pocketmine/entity/utils"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/world/sound"
)

// ExperienceManager is a port of pocketmine\entity\ExperienceManager.
type ExperienceManager struct {
	entity *Human

	levelAttr    *Attribute
	progressAttr *Attribute

	totalXp int

	canAttractXpOrbs bool

	xpCooldown int
}

// NewExperienceManager is a port of ExperienceManager::__construct.
func NewExperienceManager(entity *Human) *ExperienceManager {
	return &ExperienceManager{
		entity:           entity,
		levelAttr:        fetchAttribute(&entity.Entity, AttributeExperienceLevel),
		progressAttr:     fetchAttribute(&entity.Entity, AttributeExperience),
		canAttractXpOrbs: true,
	}
}

// GetXpLevel returns the player's experience level.
func (m *ExperienceManager) GetXpLevel() int { return int(m.levelAttr.GetValue()) }

// SetXpLevel sets the player's experience level. This does not affect their total XP or their XP
// progress.
func (m *ExperienceManager) SetXpLevel(level int) bool {
	return m.SetXpAndProgress(&level, nil)
}

// AddXpLevels adds a number of XP levels to the player, playing the level-up sound every 5 levels.
func (m *ExperienceManager) AddXpLevels(amount int, playSound bool) bool {
	oldLevel := m.GetXpLevel()
	if m.SetXpLevel(oldLevel + amount) {
		if playSound {
			newLevel := m.GetXpLevel()
			if newLevel/5 > oldLevel/5 {
				m.entity.BroadcastSound(sound.XpLevelUpSound{XpLevel: newLevel})
			}
		}

		return true
	}

	return false
}

// SubtractXpLevels subtracts a number of XP levels from the player.
func (m *ExperienceManager) SubtractXpLevels(amount int) bool {
	return m.AddXpLevels(-amount, true)
}

// GetXpProgress returns a value between 0.0 and 1.0 to indicate how far through the current level
// the player is.
func (m *ExperienceManager) GetXpProgress() float64 { return m.progressAttr.GetValue() }

// SetXpProgress sets the player's progress through the current level to a value between 0.0 and
// 1.0.
func (m *ExperienceManager) SetXpProgress(progress float64) bool {
	return m.SetXpAndProgress(nil, &progress)
}

// GetRemainderXp returns the number of XP points the player has progressed into their current
// level.
func (m *ExperienceManager) GetRemainderXp() int {
	return int(float64(entityutils.GetXpToCompleteLevel(m.GetXpLevel())) * m.GetXpProgress())
}

// GetCurrentTotalXp returns the amount of XP points the player currently has, calculated from their
// current level and progress through their current level. This will be reduced by enchanting
// deducting levels and is used to calculate the amount of XP the player drops on death.
func (m *ExperienceManager) GetCurrentTotalXp() int {
	return entityutils.GetXpToReachLevel(m.GetXpLevel()) + m.GetRemainderXp()
}

// SetCurrentTotalXp sets the current total of XP the player has, recalculating their XP level and
// progress. Note that this DOES NOT update the player's lifetime total XP.
func (m *ExperienceManager) SetCurrentTotalXp(amount int) bool {
	newLevel, err := entityutils.GetLevelFromXp(amount)
	if err != nil {
		panic(err)
	}

	xpLevel := int(newLevel)
	xpProgress := newLevel - float64(int(newLevel))
	return m.SetXpAndProgress(&xpLevel, &xpProgress)
}

// AddXp adds an amount of XP to the player, recalculating their XP level and progress. XP amount
// will be added to the player's lifetime XP.
func (m *ExperienceManager) AddXp(amount int, playSound bool) bool {
	amount = min(amount, binaryutils.Int32Max-m.totalXp)
	oldLevel := m.GetXpLevel()
	oldTotal := m.GetCurrentTotalXp()

	if m.SetCurrentTotalXp(oldTotal + amount) {
		if amount > 0 {
			m.totalXp += amount
		}

		if playSound {
			newLevel := m.GetXpLevel()
			if newLevel/5 > oldLevel/5 {
				m.entity.BroadcastSound(sound.XpLevelUpSound{XpLevel: newLevel})
			} else if m.GetCurrentTotalXp() > oldTotal {
				m.entity.BroadcastSound(sound.XpCollectSound{})
			}
		}

		return true
	}

	return false
}

// SubtractXp takes an amount of XP from the player, recalculating their XP level and progress.
func (m *ExperienceManager) SubtractXp(amount int) bool { return m.AddXp(-amount, true) }

// SetXpAndProgress is a port of ExperienceManager::setXpAndProgress: nil leaves the value unchanged.
func (m *ExperienceManager) SetXpAndProgress(level *int, progress *float64) bool {
	ev := playerevent.NewPlayerExperienceChangeEvent(m.entity.hself, m.GetXpLevel(), m.GetXpProgress(), level, progress)
	ev.Call()

	if ev.IsCancelled() {
		return false
	}

	level = ev.GetNewLevel()
	progress = ev.GetNewProgress()

	if level != nil {
		m.levelAttr.SetValue(float64(*level), false, false)
	}

	if progress != nil {
		m.progressAttr.SetValue(*progress, false, false)
	}

	return true
}

// SetXpAndProgressNoEvent is ExperienceManager::setXpAndProgressNoEvent: used for loading NBT data
// without firing events.
func (m *ExperienceManager) SetXpAndProgressNoEvent(level int, progress float64) {
	m.levelAttr.SetValue(float64(level), false, false)
	m.progressAttr.SetValue(progress, false, false)
}

// GetLifetimeTotalXp returns the total XP the player has collected in their lifetime. Resets when
// the player dies. XP levels being removed in enchanting do not reduce this number.
func (m *ExperienceManager) GetLifetimeTotalXp() int { return m.totalXp }

// SetLifetimeTotalXp panics outside 0..INT32_MAX (PHP's InvalidArgumentException).
func (m *ExperienceManager) SetLifetimeTotalXp(amount int) {
	if amount < 0 || amount > binaryutils.Int32Max {
		panic("XP must be greater than 0 and less than INT32_MAX")
	}
	m.totalXp = amount
}

// CanPickupXp returns whether the human can pickup XP orbs (checks cooldown time).
func (m *ExperienceManager) CanPickupXp() bool { return m.xpCooldown == 0 }

// mendable is the surface onPickupXp needs from item.Durable.
type mendable interface {
	item.Item
	GetDamage() int
	SetDamage(damage int)
}

// OnPickupXp is a port of ExperienceManager::onPickupXp: Mending repairs a random mending item
// before the rest of the XP is added.
func (m *ExperienceManager) OnPickupXp(xpValue int) {
	const mainHandIndex = -1
	const offHandIndex = -2

	//TODO: replace this with a more generic equipment getting/setting interface
	equipment := map[int]mendable{}

	if it, ok := m.entity.GetInventory().GetItemInHand().(mendable); ok && isDurable(it) && it.HasEnchantment(enchantment.VanillaMending(), -1) {
		equipment[mainHandIndex] = it
	}
	if it, ok := m.entity.GetOffHandInventory().GetItem(0).(mendable); ok && isDurable(it) && it.HasEnchantment(enchantment.VanillaMending(), -1) {
		equipment[offHandIndex] = it
	}
	for k, armorItem := range m.entity.GetArmorInventory().GetContents(false) {
		if it, ok := armorItem.(mendable); ok && isDurable(it) && it.HasEnchantment(enchantment.VanillaMending(), -1) {
			equipment[k] = it
		}
	}

	if len(equipment) > 0 {
		keys := make([]int, 0, len(equipment))
		for k := range equipment {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		k := keys[rand.IntN(len(keys))]
		repairItem := equipment[k]
		if repairItem.GetDamage() > 0 {
			repairAmount := min(repairItem.GetDamage(), xpValue*2)
			repairItem.SetDamage(repairItem.GetDamage() - repairAmount)
			xpValue -= int(stdmath.Ceil(float64(repairAmount) / 2))

			if k == mainHandIndex {
				m.entity.GetInventory().SetItemInHand(repairItem)
			} else if k == offHandIndex {
				m.entity.GetOffHandInventory().SetItem(0, repairItem)
			} else {
				m.entity.GetArmorInventory().SetItem(k, repairItem)
			}
		}
	}

	m.AddXp(xpValue, true) //this will still get fired even if the value is 0 due to mending, to play sounds
	m.ResetXpCooldown(2)
}

func isDurable(it item.Item) bool {
	_, ok := it.(durableItem)
	return ok
}

// ResetXpCooldown sets the duration in ticks until the human can pick up another XP orb (PHP's
// default is 2).
func (m *ExperienceManager) ResetXpCooldown(value int) { m.xpCooldown = value }

func (m *ExperienceManager) Tick(tickDiff int) {
	if m.xpCooldown > 0 {
		m.xpCooldown = max(0, m.xpCooldown-tickDiff)
	}
}

func (m *ExperienceManager) CanAttractXpOrbs() bool { return m.canAttractXpOrbs }

func (m *ExperienceManager) SetCanAttractXpOrbs(v bool) { m.canAttractXpOrbs = v }
