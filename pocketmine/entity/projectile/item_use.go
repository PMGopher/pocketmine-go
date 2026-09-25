package projectile

import (
	stdmath "math"

	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// This file is the entity half of the throwing/shooting item actions (ProjectileItem::onClickAir,
// Bow::onReleaseUsing, Trident::onReleaseUsing), installed into the item package which can't
// create entities.

// thrower is what the throwing items need from pocketmine\player\Player.
type thrower interface {
	world.Entity
	GetLocation() entity.Location
	GetEyePos() math.Vector3
	GetDirectionVector() math.Vector3
	GetItemUseDuration() int64
	HasFiniteResources() bool
	IsSpectator() bool
}

// arrowSource is the part of Player Bow::onReleaseUsing reads arrows from.
type arrowSource interface {
	// TakeArrow removes one arrow from the off-hand, or else the inventory (the inventory match
	// of Bow::onReleaseUsing), reporting whether there was one.
	TakeArrow() bool
	HasArrow() bool
}

type launchable interface {
	world.Entity
	SetMotion(motion math.Vector3) bool
	GetMotion() math.Vector3
	FlagForDespawn()
	SpawnToAll()
}

func init() {
	item.ThrowProjectileFunc = throwProjectile
	item.ReleaseBowFunc = releaseBow
	item.ReleaseTridentFunc = releaseTrident
}

// createThrowable is each ProjectileItem's createEntity.
func createThrowable(it item.ProjectileItem, location entity.Location, player world.Entity) launchable {
	switch i := it.(type) {
	case *item.Snowball:
		return NewSnowball(location, player, nil)
	case *item.Egg:
		return NewEgg(location, player, nil)
	case *item.EnderPearl:
		return NewEnderPearl(location, player, nil)
	case *item.ExperienceBottle:
		return NewExperienceBottle(location, player, nil)
	case *item.IceBomb:
		return NewIceBomb(location, player, nil)
	case *item.SplashPotion:
		p := NewSplashPotion(location, player, i.PotionTypeValue, nil)
		p.SetLinger(i.Linger)
		return p
	}
	return nil
}

// throwProjectile is ProjectileItem::onClickAir (minus the pop, done by the item).
func throwProjectile(it item.ProjectileItem, p item.Player, directionVector math.Vector3) item.ItemUseResult {
	player, ok := p.(thrower)
	if !ok {
		return item.ItemUseResultNone
	}
	location := player.GetLocation()
	projectile := createThrowable(it, entity.LocationFromObject(player.GetEyePos(), player.GetWorld(), location.Yaw, location.Pitch), player)
	if projectile == nil {
		return item.ItemUseResultNone
	}
	projectile.SetMotion(directionVector.Multiply(it.GetThrowForce()))

	projectileEv := entityevent.NewProjectileLaunchEvent(projectile)
	projectileEv.Call()
	if projectileEv.IsCancelled() {
		projectile.FlagForDespawn()
		return item.ItemUseResultFail
	}

	projectile.SpawnToAll()
	location.GetWorld().AddSound(location.Vector3, sound.ThrowSound{})
	return item.ItemUseResultSuccess
}

// isProjectileEntity is PHP's `$entity instanceof Projectile`.
func isProjectileEntity(e world.Entity) bool {
	_, ok := e.(interface{ GetBaseDamage() float64 })
	return ok
}

// releaseBow is a port of Bow::onReleaseUsing.
func releaseBow(bow *item.Bow, p item.Player, returnedItems *[]item.Item) item.ItemUseResult {
	player, ok := p.(thrower)
	if !ok {
		return item.ItemUseResultNone
	}
	arrows, _ := p.(arrowSource)
	hasArrow := arrows != nil && arrows.HasArrow()
	if player.HasFiniteResources() && !hasArrow {
		return item.ItemUseResultFail
	}

	location := player.GetLocation()
	diff := float64(player.GetItemUseDuration())
	pw := diff / 20
	baseForce := stdmath.Min(((pw*pw)+pw*2)/3, 1)

	yaw := -location.Yaw
	if location.Yaw > 180 {
		yaw = 360 - location.Yaw
	}
	arrow := NewArrow(entity.LocationFromObject(player.GetEyePos(), player.GetWorld(), yaw, -location.Pitch), player, baseForce >= 1, nil)
	arrow.SetMotion(player.GetDirectionVector())

	infinity := bow.HasEnchantment(enchantment.VanillaInfinity(), -1)
	if infinity {
		arrow.SetPickupMode(ArrowPickupCreative)
	}
	if punchLevel := bow.GetEnchantmentLevel(enchantment.VanillaPunch()); punchLevel > 0 {
		arrow.SetPunchKnockback(float64(punchLevel))
	}
	if powerLevel := bow.GetEnchantmentLevel(enchantment.VanillaPower()); powerLevel > 0 {
		arrow.SetBaseDamage(arrow.GetBaseDamage() + float64(powerLevel+1)/2)
	}
	if bow.HasEnchantment(enchantment.VanillaFlame(), -1) {
		arrow.SetOnFire(arrow.GetFireTicks()/20 + 100)
	}

	ev := entityevent.NewEntityShootBowEvent(player, bow, arrow, baseForce*3)
	if baseForce < 0.1 || diff < 5 || player.IsSpectator() {
		ev.Cancel()
	}
	ev.Call()

	shot, _ := ev.GetProjectile().(launchable) //This might have been changed by plugins
	if ev.IsCancelled() || shot == nil {
		if shot != nil {
			shot.FlagForDespawn()
		}
		return item.ItemUseResultFail
	}

	shot.SetMotion(shot.GetMotion().Multiply(ev.GetForce()))

	if isProjectileEntity(shot) {
		projectileEv := entityevent.NewProjectileLaunchEvent(shot)
		projectileEv.Call()
		if projectileEv.IsCancelled() {
			shot.FlagForDespawn()
			return item.ItemUseResultFail
		}
		shot.SpawnToAll()
		location.GetWorld().AddSound(location.Vector3, sound.BowShootSound{})
	} else {
		shot.SpawnToAll()
	}

	if player.HasFiniteResources() {
		if !infinity && arrows != nil { //TODO: tipped arrows are still consumed when Infinity is applied
			arrows.TakeArrow()
		}
		bow.ApplyDamage(1)
	}
	return item.ItemUseResultSuccess
}

// releaseTrident is a port of Trident::onReleaseUsing.
func releaseTrident(trident *item.Trident, p item.Player, returnedItems *[]item.Item) item.ItemUseResult {
	player, ok := p.(thrower)
	if !ok {
		return item.ItemUseResultNone
	}
	location := player.GetLocation()

	diff := float64(player.GetItemUseDuration())
	if diff < 14 {
		return item.ItemUseResultFail
	}

	thrown := trident.PopCount(1).(*item.Trident)
	if player.HasFiniteResources() {
		thrown.ApplyDamage(item.TridentDamageOnThrow)
	}
	if thrown.IsNull() {
		//canStartUsingItem() will normally prevent this, but it's possible the item might've been modified between
		//the start action and the release, so it's best to account for this anyway
		return item.ItemUseResultFail
	}

	yaw := -location.Yaw
	if location.Yaw > 180 {
		yaw = 360 - location.Yaw
	}
	e := NewTrident(entity.LocationFromObject(player.GetEyePos(), player.GetWorld(), yaw, -location.Pitch), thrown, player, nil)
	pw := diff / 20
	baseForce := stdmath.Min(((pw*pw)+pw*2)/3, 1) * 2.4
	e.SetMotion(player.GetDirectionVector().Multiply(baseForce))

	ev := entityevent.NewProjectileLaunchEvent(e)
	ev.Call()
	if ev.IsCancelled() {
		e.FlagForDespawn()
		return item.ItemUseResultFail
	}
	e.SpawnToAll()
	location.GetWorld().AddSound(location.Vector3, sound.TridentThrowSound{})
	return item.ItemUseResultSuccess
}
