package keeper

// Migrator is a struct for handling in-place store migrations.
// Since we're starting fresh at v2, we don't need migration functions.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{
		keeper: keeper,
	}
}
