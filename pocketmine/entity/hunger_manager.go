package entity

import (
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/world"
)

// HungerManager is a port of pocketmine\entity\HungerManager.
type HungerManager struct {
	entity *Human

	hungerAttr     *Attribute
	saturationAttr *Attribute
	exhaustionAttr *Attribute

	foodTickTimer int

	enabled bool
}

// NewHungerManager is a port of HungerManager::__construct.
func NewHungerManager(entity *Human) *HungerManager {
	return &HungerManager{
		entity:         entity,
		hungerAttr:     fetchAttribute(&entity.Entity, AttributeHunger),
		saturationAttr: fetchAttribute(&entity.Entity, AttributeSaturation),
		exhaustionAttr: fetchAttribute(&entity.Entity, AttributeExhaustion),
		enabled:        true,
	}
}

func fetchAttribute(entity *Entity, attributeID string) *Attribute {
	attribute := GetAttributeFactory().MustGet(attributeID)
	entity.GetAttributeMap().Add(attribute)
	return attribute
}

func (m *HungerManager) GetFood() float64 { return m.hungerAttr.GetValue() }

// SetFood is a port of HungerManager::setFood (panicking on an out-of-range value).
func (m *HungerManager) SetFood(newFood float64) {
	old := m.hungerAttr.GetValue()
	m.hungerAttr.SetValue(newFood, false, false)

	// ranges: 18-20 (regen), 7-17 (none), 1-6 (no sprint), 0 (health depletion)
	for _, bound := range []float64{17, 6, 0} {
		if (old > bound) != (newFood > bound) {
			m.foodTickTimer = 0
			break
		}
	}
}

func (m *HungerManager) GetMaxFood() float64 { return m.hungerAttr.GetMaxValue() }

func (m *HungerManager) AddFood(amount float64) {
	amount += m.hungerAttr.GetValue()
	amount = max(min(amount, m.hungerAttr.GetMaxValue()), m.hungerAttr.GetMinValue())
	m.SetFood(amount)
}

// IsHungry returns whether this Human may consume objects requiring hunger.
func (m *HungerManager) IsHungry() bool { return m.GetFood() < m.GetMaxFood() }

func (m *HungerManager) GetSaturation() float64 { return m.saturationAttr.GetValue() }

// SetSaturation sets the saturation level (panicking on an out-of-range value).
func (m *HungerManager) SetSaturation(saturation float64) {
	m.saturationAttr.SetValue(saturation, false, false)
}

func (m *HungerManager) AddSaturation(amount float64) {
	m.saturationAttr.SetValue(m.saturationAttr.GetValue()+amount, true, false)
}

func (m *HungerManager) GetExhaustion() float64 { return m.exhaustionAttr.GetValue() }

// SetExhaustion is a port of HungerManager::setExhaustion (panicking on an out-of-range value).
func (m *HungerManager) SetExhaustion(exhaustion float64) {
	m.exhaustionAttr.SetValue(exhaustion, false, false)
}

// Exhaust is a port of HungerManager::exhaust: returns the amount of exhaustion level increased.
func (m *HungerManager) Exhaust(amount float64, cause int) float64 {
	if !m.enabled {
		return 0
	}
	evAmount := amount
	if entityevent.HasHandlers[playerevent.PlayerExhaustEvent]() {
		ev := playerevent.NewPlayerExhaustEvent(m.entity.hself, amount, cause)
		ev.Call()
		if ev.IsCancelled() {
			return 0.0
		}
		evAmount = ev.GetAmount()
	}

	exhaustion := m.GetExhaustion()
	exhaustion += evAmount

	for exhaustion >= 4.0 {
		exhaustion -= 4.0

		saturation := m.GetSaturation()
		if saturation > 0 {
			saturation = max(0, saturation-1.0)
			m.SetSaturation(saturation)
		} else {
			food := m.GetFood()
			if food > 0 {
				food--
				m.SetFood(max(food, 0))
			}
		}
	}
	m.SetExhaustion(exhaustion)

	return evAmount
}

func (m *HungerManager) GetFoodTickTimer() int { return m.foodTickTimer }

// SetFoodTickTimer panics on a negative value (PHP's InvalidArgumentException).
func (m *HungerManager) SetFoodTickTimer(foodTickTimer int) {
	if foodTickTimer < 0 {
		panic("Expected a non-negative value")
	}
	m.foodTickTimer = foodTickTimer
}

// Tick is a port of HungerManager::tick: natural regeneration, starvation and the sprint cutoff.
func (m *HungerManager) Tick(tickDiff int) {
	if !m.entity.IsAlive() || !m.enabled {
		return
	}
	food := m.GetFood()
	health := m.entity.hself.GetHealth()
	difficulty := m.entity.GetWorld().GetDifficulty()

	m.foodTickTimer += tickDiff
	if m.foodTickTimer >= 80 {
		m.foodTickTimer = 0
	}

	if difficulty == world.DifficultyPeaceful && m.foodTickTimer%10 == 0 {
		if food < m.GetMaxFood() {
			m.AddFood(1.0)
			food = m.GetFood()
		}
		if m.foodTickTimer%20 == 0 && health < float64(m.entity.hself.GetMaxHealth()) {
			m.entity.Heal(entityevent.NewEntityRegainHealthEvent(m.entity.hself, 1, entityevent.RegainCauseSaturation))
		}
	}

	if m.foodTickTimer == 0 {
		if food >= 18 {
			if health < float64(m.entity.hself.GetMaxHealth()) {
				m.entity.Heal(entityevent.NewEntityRegainHealthEvent(m.entity.hself, 1, entityevent.RegainCauseSaturation))
				m.Exhaust(6.0, playerevent.ExhaustCauseHealthRegen)
			}
		} else if food <= 0 {
			if (difficulty == world.DifficultyEasy && health > 10) || (difficulty == world.DifficultyNormal && health > 1) || difficulty == world.DifficultyHard {
				m.entity.hself.Attack(entityevent.NewEntityDamageEvent(m.entity.hself, entityevent.CauseStarvation, 1, nil))
			}
		}
	}

	if food <= 6 {
		m.entity.hself.SetSprinting(false)
	}
}

func (m *HungerManager) IsEnabled() bool { return m.enabled }

func (m *HungerManager) SetEnabled(enabled bool) { m.enabled = enabled }
