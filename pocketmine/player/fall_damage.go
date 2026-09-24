package player

// Fall damage for players is entity.Living's real OnHitGround (damage, and the long/short fall and
// landing sounds), reached through HandleMovement -> Entity.Move -> Player.CheckGroundState ->
// Entity.UpdateFallState, exactly as PHP's Player::handleMovement -> Entity::move does. The fall
// distance is measured from the positions the client reports; flying players take none (see
// Player.CalculateFallDamage).
